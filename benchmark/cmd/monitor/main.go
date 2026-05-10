// Package main: monitor — sonde CPU%/RAM/threads d'un PID pendant la durée d'un benchmark.
//
// Usage :
//
//	monitor -pid <PID> -duration <s> -interval <ms> -out <chemin.csv>
//
// Sortie : CSV avec colonnes
//
//	timestamp_unix_ms,elapsed_s,cpu_percent,mem_rss_mb,num_threads
//
// Le programme s'arrête à la fin de la durée OU si le PID disparaît, selon le
// premier événement. Sur SIGINT, il flushe le CSV et quitte proprement.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	pid := flag.Int("pid", 0, "PID du processus à observer (obligatoire)")
	duration := flag.Duration("duration", 60*time.Second, "Durée totale d'observation (ex: 90s, 4m)")
	interval := flag.Duration("interval", 1*time.Second, "Intervalle de sampling (ex: 500ms, 1s)")
	out := flag.String("out", "benchmark/results/monitor.csv", "Chemin du CSV de sortie")
	flag.Parse()

	if *pid <= 0 {
		log.Fatal("flag -pid obligatoire")
	}

	proc, err := process.NewProcess(int32(*pid))
	if err != nil {
		log.Fatalf("impossible d'attacher le PID %d : %v", *pid, err)
	}

	// Premier appel de CPUPercent : initialise la baseline.
	if _, err := proc.CPUPercent(); err != nil {
		log.Printf("avertissement : CPUPercent initial : %v", err)
	}

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("création du CSV : %v", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"timestamp_unix_ms", "elapsed_s", "cpu_percent", "mem_rss_mb", "num_threads"}); err != nil {
		log.Fatalf("écriture en-tête : %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	start := time.Now()
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	samples := 0

	// Échantillon immédiat au démarrage pour les benchs très courts (< interval).
	if cpu, errCPU := proc.CPUPercent(); errCPU == nil {
		if mem, errMem := proc.MemoryInfo(); errMem == nil {
			if threads, errThreads := proc.NumThreads(); errThreads == nil {
				rec := []string{
					strconv.FormatInt(start.UnixMilli(), 10),
					"0.000",
					strconv.FormatFloat(cpu, 'f', 2, 64),
					strconv.FormatFloat(float64(mem.RSS)/1024.0/1024.0, 'f', 2, 64),
					strconv.FormatInt(int64(threads), 10),
				}
				if err := w.Write(rec); err == nil {
					samples++
					w.Flush()
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("monitor : durée écoulée, %d échantillons écrits dans %s", samples, *out)
			return
		case <-sigCh:
			log.Printf("monitor : interruption, %d échantillons écrits dans %s", samples, *out)
			return
		case t := <-ticker.C:
			running, err := proc.IsRunning()
			if err != nil || !running {
				log.Printf("monitor : PID %d n'est plus actif, %d échantillons écrits dans %s", *pid, samples, *out)
				return
			}
			cpu, errCPU := proc.CPUPercent()
			mem, errMem := proc.MemoryInfo()
			threads, errThreads := proc.NumThreads()

			if errCPU != nil || errMem != nil || errThreads != nil {
				log.Printf("monitor : erreur de sample (cpu=%v mem=%v threads=%v)", errCPU, errMem, errThreads)
				continue
			}

			elapsed := t.Sub(start).Seconds()
			rec := []string{
				strconv.FormatInt(t.UnixMilli(), 10),
				strconv.FormatFloat(elapsed, 'f', 3, 64),
				strconv.FormatFloat(cpu, 'f', 2, 64),
				strconv.FormatFloat(float64(mem.RSS)/1024.0/1024.0, 'f', 2, 64),
				strconv.FormatInt(int64(threads), 10),
			}
			if err := w.Write(rec); err != nil {
				log.Printf("monitor : écriture ligne : %v", err)
				continue
			}
			samples++
			// Flush toutes les 5s pour ne pas perdre tout en cas de kill.
			if samples%5 == 0 {
				w.Flush()
			}
		}
	}
}

