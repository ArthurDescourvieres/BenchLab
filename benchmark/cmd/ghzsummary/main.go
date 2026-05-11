// ghzsummary lit un rapport ghz JSON sur stdin et affiche un résumé lisible.
// Usage : ghz ... --format=json | tee result.json | ghzsummary
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ghzResult struct {
	Count                  int            `json:"count"`
	Total                  int64          `json:"total"`
	Average                int64          `json:"average"`
	Fastest                int64          `json:"fastest"`
	Slowest                int64          `json:"slowest"`
	RPS                    float64        `json:"rps"`
	ErrorDistribution      map[string]int `json:"errorDistribution"`
	StatusCodeDistribution map[string]int `json:"statusCodeDistribution"`
	LatencyDistribution    []struct {
		Percentage int   `json:"percentage"`
		Latency    int64 `json:"latency"`
	} `json:"latencyDistribution"`
}

func ms(ns int64) string {
	return fmt.Sprintf("%.2f ms", float64(ns)/1e6)
}

func main() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghzsummary: lecture stdin: %v\n", err)
		os.Exit(1)
	}

	var r ghzResult
	if err := json.Unmarshal(raw, &r); err != nil {
		// JSON invalide : on affiche brut et on sort proprement
		os.Stdout.Write(raw)
		return
	}

	sep := "────────────────────────────────────────────────────────────"
	fmt.Println(sep)
	fmt.Printf("  Requêtes  : %d  |  RPS : %.0f req/s  |  Durée : %.2f s\n",
		r.Count, r.RPS, float64(r.Total)/1e9)
	fmt.Printf("  Latence   : avg=%-12s  min=%-12s  max=%s\n",
		ms(r.Average), ms(r.Fastest), ms(r.Slowest))
	for _, p := range r.LatencyDistribution {
		if p.Percentage > 0 {
			fmt.Printf("              p%-5d %s\n", p.Percentage, ms(p.Latency))
		}
	}
	if len(r.ErrorDistribution) > 0 {
		fmt.Printf("  Erreurs   : %v\n", r.ErrorDistribution)
	} else {
		fmt.Println("  Erreurs   : aucune")
	}
	fmt.Printf("  Statuts   : %v\n", r.StatusCodeDistribution)
	fmt.Println(sep)
}
