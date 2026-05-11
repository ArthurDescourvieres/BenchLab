package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed static
var staticFiles embed.FS

const resultsDir = "benchmark/results"

var allowedTargets = map[string]bool{
	"bench-rest-a":    true,
	"bench-rest-b":    true,
	"bench-rest-c":    true,
	"bench-rest-gzip": true,
	"bench-grpc-a":    true,
	"bench-grpc-b":    true,
	"bench-grpc-c":    true,
	"sysinfo":         true,
	"payload":         true,
}

func main() {
	mux := http.NewServeMux()

	sub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/api/results", handleListResults)
	mux.HandleFunc("/api/results/", handleGetResult)
	mux.HandleFunc("/api/sysinfo", handleFile("system-info.json"))
	mux.HandleFunc("/api/payload", handleFile("payload-size.json"))
	mux.HandleFunc("/api/run/", handleRun)

	println("BenchLab Dashboard → http://localhost:8090")
	if err := http.ListenAndServe(":8090", mux); err != nil {
		panic(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func handleListResults(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		http.Error(w, "cannot read results dir", http.StatusInternalServerError)
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	jsonResp(w, names)
}

func handleGetResult(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/results/")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, `\`) {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	serveJSONFile(w, filepath.Join(resultsDir, name))
}

func handleFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveJSONFile(w, filepath.Join(resultsDir, filename))
	}
}

func serveJSONFile(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	target := strings.TrimPrefix(r.URL.Path, "/api/run/")
	if !allowedTargets[target] {
		http.Error(w, "unknown target", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "make", target)
	out, err := cmd.CombinedOutput()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     err == nil,
		"output": string(out),
	})
}

func jsonResp(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
