// k6summary lit un rapport k6 JSON (handleSummary) et affiche un résumé lisible.
// Usage : k6summary benchmark/results/k6-rest-a-summary.json
//         ou       : cat summary.json | k6summary
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type k6Summary struct {
	State struct {
		DurationMs float64 `json:"testRunDurationMs"`
	} `json:"state"`
	Metrics map[string]struct {
		Values map[string]float64 `json:"values"`
	} `json:"metrics"`
}

func ms(v float64) string {
	return fmt.Sprintf("%.2f ms", v)
}

func main() {
	var raw []byte
	var err error
	if len(os.Args) > 1 {
		raw, err = os.ReadFile(os.Args[1])
	} else {
		raw, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "k6summary: %v\n", err)
		os.Exit(1)
	}

	var s k6Summary
	if err := json.Unmarshal(raw, &s); err != nil {
		os.Stdout.Write(raw)
		return
	}

	iter := s.Metrics["iterations"].Values
	dur := s.Metrics["http_req_duration"].Values
	failed := int(s.Metrics["http_req_failed"].Values["passes"])
	count := int(iter["count"])
	rps := iter["rate"]
	durSec := s.State.DurationMs / 1000

	sep := "────────────────────────────────────────────────────────────"
	fmt.Println(sep)
	fmt.Printf("  Requêtes  : %d  |  RPS : %.0f req/s  |  Durée : %.2f s\n", count, rps, durSec)
	fmt.Printf("  Latence   : avg=%-12s  min=%-12s  max=%s\n", ms(dur["avg"]), ms(dur["min"]), ms(dur["max"]))
	fmt.Printf("              p50    %s\n", ms(dur["med"]))
	fmt.Printf("              p75    %s\n", ms(dur["p(75)"]))
	fmt.Printf("              p90    %s\n", ms(dur["p(90)"]))
	fmt.Printf("              p95    %s\n", ms(dur["p(95)"]))
	fmt.Printf("              p99    %s\n", ms(dur["p(99)"]))
	if failed > 0 {
		fmt.Printf("  Erreurs   : %d requête(s) échouée(s)\n", failed)
	} else {
		fmt.Println("  Erreurs   : aucune")
	}
	if failed == 0 {
		fmt.Printf("  Statuts   : %d OK\n", count)
	} else {
		fmt.Printf("  Statuts   : %d OK, %d KO\n", count-failed, failed)
	}
	fmt.Println(sep)
}
