# BenchLab — Makefile d'orchestration
#
# Cibles principales :
#   make build         → compile rest + grpc + monitor dans bin/
#   make sysinfo       → collecte les conditions de test (système, versions)
#   make payload       → régénère benchmark/results/payload-size.json
#   make bench-rest-a  → benchmark REST scénario A (lecture)
#   make bench-rest-b  → benchmark REST scénario B (écriture)
#   make bench-rest-c  → benchmark REST scénario C (ramp)
#   make bench-rest-gzip → benchmark REST scénario A avec gzip activé
#   make bench-grpc-a  → benchmark gRPC scénario A (lecture)
#   make bench-grpc-b  → benchmark gRPC scénario B (écriture)
#   make bench-grpc-c  → benchmark gRPC scénario C (paliers de concurrence)
#   make bench-all     → enchaîne tous les benchmarks + sysinfo + payload
#   make clean         → supprime bin/ et benchmark/results/*.{json,csv,log}
#
# Pré-requis : Go, k6, ghz dans le PATH. Sous Windows, utiliser Git Bash.

SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

EXE := $(shell go env GOEXE)
BIN_REST := bin/rest$(EXE)
BIN_GRPC := bin/grpc$(EXE)
BIN_MONITOR := bin/monitor$(EXE)
BIN_DASHBOARD := bin/dashboard$(EXE)
BIN_GHZSUMMARY := bin/ghzsummary$(EXE)

REST_PORT ?= 8080
GRPC_PORT ?= 9090
REST_BASE := http://localhost:$(REST_PORT)
GRPC_ADDR := localhost:$(GRPC_PORT)

RESULTS := benchmark/results

.PHONY: help build sysinfo payload \
        bench-rest-a bench-rest-b bench-rest-c bench-rest-gzip \
        bench-grpc-a bench-grpc-b bench-grpc-c \
        bench-all clean \
        build-dashboard run-dashboard

help:
	@grep -E '^[a-zA-Z_-]+:.*?$$' $(MAKEFILE_LIST) | grep -v '^\.' | sort

build: $(BIN_REST) $(BIN_GRPC) $(BIN_MONITOR) $(BIN_GHZSUMMARY)

$(BIN_REST):
	@mkdir -p bin
	go build -o "$(BIN_REST)" ./rest-service

$(BIN_GRPC):
	@mkdir -p bin
	go build -o "$(BIN_GRPC)" ./grpc-service

$(BIN_MONITOR):
	@mkdir -p bin
	go build -o "$(BIN_MONITOR)" ./benchmark/cmd/monitor

$(BIN_GHZSUMMARY):
	@mkdir -p bin
	go build -o "$(BIN_GHZSUMMARY)" ./benchmark/cmd/ghzsummary

sysinfo:
	bash benchmark/scripts/collect-system-info.sh

payload:
	@mkdir -p $(RESULTS)
	go run ./benchmark/cmd/payloadsize

# --- REST ---------------------------------------------------------------------

bench-rest-a: $(BIN_REST) $(BIN_MONITOR)
	@$(MAKE) _bench-rest SCRIPT=benchmark/scripts/k6-rest-a-read.js MONFILE=monitor-rest-a.csv DUR=180s

bench-rest-b: $(BIN_REST) $(BIN_MONITOR)
	@$(MAKE) _bench-rest SCRIPT=benchmark/scripts/k6-rest-b-write.js MONFILE=monitor-rest-b.csv DUR=120s

bench-rest-c: $(BIN_REST) $(BIN_MONITOR)
	@$(MAKE) _bench-rest SCRIPT=benchmark/scripts/k6-rest-c-ramp.js MONFILE=monitor-rest-c.csv DUR=300s

bench-rest-gzip: $(BIN_REST) $(BIN_MONITOR)
	@$(MAKE) _bench-rest SCRIPT=benchmark/scripts/k6-rest-a-read-gzip.js MONFILE=monitor-rest-a-gzip.csv DUR=180s GZIP=1

# Cible interne : démarre le service REST en background, attache le monitor,
# lance k6, puis nettoie. Variables : SCRIPT, MONFILE, DUR, GZIP (optionnel).
_bench-rest:
	@mkdir -p $(RESULTS)
	@if [ "$${GZIP:-0}" = "1" ]; then export REST_GZIP=1; echo "[bench] REST_GZIP=1 (gzip activé)"; fi
	export PORT=$(REST_PORT); \
	"$(BIN_REST)" > $(RESULTS)/rest-service.log 2> $(RESULTS)/rest-service-err.log & \
	BASH_PID=$$!; \
	trap 'kill $$BASH_PID 2>/dev/null || true; kill $$MON_PID 2>/dev/null || true' EXIT INT TERM; \
	echo "[bench] REST démarré, attente 1s..."; \
	sleep 1; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
	  if curl -fsS -o /dev/null "$(REST_BASE)/health"; then break; fi; \
	  sleep 0.5; \
	done; \
	WIN_PID=$$(cat /proc/$$BASH_PID/winpid 2>/dev/null | tr -d '\r\n'); \
	if [ -z "$$WIN_PID" ]; then WIN_PID=$$(netstat -ano 2>/dev/null | grep " $(REST_PORT) " | grep "LISTENING" | awk '{print $$NF}' | head -1 | tr -d '\r'); fi; \
	REST_PID=$${WIN_PID:-$$BASH_PID}; \
	echo "[bench] REST PID monitor: $$REST_PID"; \
	"$(BIN_MONITOR)" -pid $$REST_PID -duration $(DUR) -interval 1s -out "$(RESULTS)/$(MONFILE)" & \
	MON_PID=$$!; \
	echo "[bench] monitor démarré (PID=$$MON_PID) → $(RESULTS)/$(MONFILE)"; \
	BASE_URL=$(REST_BASE) k6 run "$(SCRIPT)"; \
	BENCH_RC=$$?; \
	kill $$BASH_PID 2>/dev/null || true; \
	wait $$MON_PID 2>/dev/null || true; \
	exit $$BENCH_RC

# --- gRPC ---------------------------------------------------------------------

bench-grpc-a: $(BIN_GRPC) $(BIN_MONITOR) $(BIN_GHZSUMMARY)
	@$(MAKE) _bench-grpc TOOL=ghz MONFILE=monitor-grpc-a.csv DUR=180s GHZ_SCRIPT=benchmark/scripts/grpc-a-read.sh

bench-grpc-b: $(BIN_GRPC) $(BIN_MONITOR) $(BIN_GHZSUMMARY)
	@$(MAKE) _bench-grpc TOOL=ghz MONFILE=monitor-grpc-b.csv DUR=120s GHZ_SCRIPT=benchmark/scripts/grpc-b-write.sh

bench-grpc-c: $(BIN_GRPC) $(BIN_MONITOR) $(BIN_GHZSUMMARY)
	@$(MAKE) _bench-grpc TOOL=ghz MONFILE=monitor-grpc-c.csv DUR=600s GHZ_SCRIPT=benchmark/scripts/grpc-c-ramp.sh

# Cible interne : démarre le service gRPC en background, attache le monitor,
# lance le script ghz (ou k6 gRPC si on l'ajoutait plus tard), puis nettoie.
_bench-grpc:
	@mkdir -p $(RESULTS)
	export GRPC_PORT=$(GRPC_PORT); \
	"$(BIN_GRPC)" > $(RESULTS)/grpc-service.log 2> $(RESULTS)/grpc-service-err.log & \
	BASH_PID=$$!; \
	trap 'kill $$BASH_PID 2>/dev/null || true; kill $$MON_PID 2>/dev/null || true' EXIT INT TERM; \
	echo "[bench] gRPC démarré, attente 1.5s..."; \
	sleep 1.5; \
	if ! kill -0 $$BASH_PID 2>/dev/null; then \
	  echo "[bench] ERREUR : le service gRPC a quitté prématurément."; \
	  echo "[bench] Port $(GRPC_PORT) déjà utilisé ? Logs :"; \
	  cat $(RESULTS)/grpc-service-err.log 2>/dev/null || true; \
	  exit 1; \
	fi; \
	WIN_PID=$$(cat /proc/$$BASH_PID/winpid 2>/dev/null | tr -d '\r\n'); \
	if [ -z "$$WIN_PID" ]; then WIN_PID=$$(netstat -ano 2>/dev/null | grep " $(GRPC_PORT) " | grep "LISTENING" | awk '{print $$NF}' | head -1 | tr -d '\r'); fi; \
	GRPC_PID=$${WIN_PID:-$$BASH_PID}; \
	echo "[bench] gRPC PID monitor: $$GRPC_PID"; \
	"$(BIN_MONITOR)" -pid $$GRPC_PID -duration $(DUR) -interval 1s -out "$(RESULTS)/$(MONFILE)" & \
	MON_PID=$$!; \
	echo "[bench] monitor démarré (PID=$$MON_PID) → $(RESULTS)/$(MONFILE)"; \
	GRPC_ADDR=$(GRPC_ADDR) bash "$(GHZ_SCRIPT)"; \
	BENCH_RC=$$?; \
	kill $$BASH_PID 2>/dev/null || true; \
	wait $$MON_PID 2>/dev/null || true; \
	exit $$BENCH_RC

# --- Tout en un ---------------------------------------------------------------

bench-all: build sysinfo payload \
           bench-rest-a bench-rest-b bench-rest-c \
           bench-grpc-a bench-grpc-b bench-grpc-c \
           bench-rest-gzip
	@echo "[bench] tous les scénarios terminés. Résultats dans $(RESULTS)/"

# --- Nettoyage ---------------------------------------------------------------

build-dashboard:
	@mkdir -p bin
	go build -o "$(BIN_DASHBOARD)" ./dashboard

run-dashboard: build-dashboard
	./$(BIN_DASHBOARD)

clean:
	rm -rf bin
	rm -f $(RESULTS)/*.json $(RESULTS)/*.csv $(RESULTS)/*.log $(RESULTS)/*.txt
	@echo "[clean] bin/ et résultats supprimés (le dossier $(RESULTS)/ est conservé)"
