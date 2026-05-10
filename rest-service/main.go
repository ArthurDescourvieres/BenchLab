package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ArthurDescourvieres/BenchLab/internal/sensorsvc"
	"github.com/ArthurDescourvieres/BenchLab/store"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		if p[0] == ':' {
			addr = p
		} else {
			addr = ":" + p
		}
	}

	if pf := os.Getenv("PID_FILE"); pf != "" {
		_ = os.WriteFile(pf, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	if os.Getenv("REST_GZIP") == "1" {
		r.Use(gzip.Gzip(gzip.DefaultCompression))
		log.Print("REST gzip middleware enabled (REST_GZIP=1)")
	}
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	st := store.NewMemoryStore()
	svc := sensorsvc.New(st)
	RegisterRoutes(r, svc)

	log.Printf("REST listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
