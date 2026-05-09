#!/usr/bin/env bash
# bench-all.sh — orchestrateur unique : compile, collecte sysinfo, lance les
# 3 scénarios REST + 3 scénarios gRPC + variante gzip, en attachant à chaque
# fois le monitor CPU/RAM. Sortie dans benchmark/results/.
#
# Usage : bash benchmark/scripts/bench-all.sh [scénario]
#   sans argument           → enchaîne tout (sysinfo + payload + 7 benchs)
#   rest-a / rest-b / rest-c / rest-gzip → un seul scénario REST
#   grpc-a / grpc-b / grpc-c → un seul scénario gRPC
#   sysinfo                 → ne fait que collecter les conditions de test
#   payload                 → ne fait que régénérer payload-size.json
#   build                   → ne fait que compiler les binaires
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

EXE="$(go env GOEXE)"
BIN_REST="bin/rest$EXE"
BIN_GRPC="bin/grpc$EXE"
BIN_MONITOR="bin/monitor$EXE"

REST_PORT="${REST_PORT:-8080}"
GRPC_PORT="${GRPC_PORT:-9090}"
REST_BASE="http://localhost:${REST_PORT}"
GRPC_ADDR="localhost:${GRPC_PORT}"

RESULTS="benchmark/results"
mkdir -p "$RESULTS"

log() { echo "[$(date +%H:%M:%S)] $*"; }
die() { echo "[fatal] $*" >&2; exit 1; }

build_all() {
  log "Compilation des binaires"
  mkdir -p bin
  go build -o "$BIN_REST" ./rest-service || die "build rest"
  go build -o "$BIN_GRPC" ./grpc-service || die "build grpc"
  go build -o "$BIN_MONITOR" ./benchmark/cmd/monitor || die "build monitor"
}

collect_sysinfo() {
  log "Collecte des conditions de test"
  bash benchmark/scripts/collect-system-info.sh || die "collect-system-info"
}

regen_payload() {
  log "Régénération de payload-size.json"
  go run ./benchmark/cmd/payloadsize || die "payloadsize"
}

# Lance le service REST en background, le monitor sur son PID, puis k6.
# $1 : chemin du script k6
# $2 : nom du fichier monitor CSV (relatif à $RESULTS)
# $3 : durée max du monitor (ex: 180s)
# $4 : 1 pour activer REST_GZIP=1, 0 sinon
run_rest_bench() {
  local script="$1"
  local monfile="$2"
  local dur="$3"
  local gzip="${4:-0}"

  local rest_pid="" rest_bash_pid="" mon_pid=""
  cleanup() {
    [ -n "$rest_bash_pid" ] && kill "$rest_bash_pid" 2>/dev/null || true
    [ -n "$rest_pid" ] && kill "$rest_pid" 2>/dev/null || true
    [ -n "$mon_pid" ] && kill "$mon_pid" 2>/dev/null || true
    wait 2>/dev/null || true
  }
  trap cleanup EXIT INT TERM

  if [ "$gzip" = "1" ]; then
    export REST_GZIP=1
    log "REST_GZIP=1 (gzip activé)"
  else
    unset REST_GZIP
  fi
  export PORT="$REST_PORT"

  local pidfile="$RESULTS/rest.pid"
  rm -f "$pidfile"
  export PID_FILE="$pidfile"

  log "Démarrage REST → $RESULTS/rest-service.log"
  "$BIN_REST" > "$RESULTS/rest-service.log" 2> "$RESULTS/rest-service-err.log" &
  rest_bash_pid=$!
  unset PID_FILE

  for _ in 1 2 3 4 5 6 7 8 9 10; do
    [ -s "$pidfile" ] && break
    sleep 0.2
  done
  rest_pid="$(cat "$pidfile" 2>/dev/null || echo "")"
  if [ -z "$rest_pid" ]; then
    die "REST n'a pas écrit son PID dans $pidfile"
  fi

  for _ in 1 2 3 4 5 6 7 8 9 10; do
    if curl -fsS -o /dev/null "$REST_BASE/health"; then break; fi
    sleep 0.3
  done

  log "Démarrage monitor PID=$rest_pid → $RESULTS/$monfile (dur=$dur)"
  "$BIN_MONITOR" -pid "$rest_pid" -duration "$dur" -interval 1s -out "$RESULTS/$monfile" &
  mon_pid=$!

  log "Lancement k6 : $script"
  BASE_URL="$REST_BASE" k6 run "$script"
  local rc=$?

  cleanup
  trap - EXIT INT TERM
  return $rc
}

# Lance le service gRPC en background, le monitor sur son PID, puis le script ghz.
# $1 : chemin du script ghz à exécuter
# $2 : nom du fichier monitor CSV
# $3 : durée max du monitor
run_grpc_bench() {
  local ghz_script="$1"
  local monfile="$2"
  local dur="$3"

  local grpc_pid="" grpc_bash_pid="" mon_pid=""
  cleanup() {
    [ -n "$grpc_bash_pid" ] && kill "$grpc_bash_pid" 2>/dev/null || true
    [ -n "$grpc_pid" ] && kill "$grpc_pid" 2>/dev/null || true
    [ -n "$mon_pid" ] && kill "$mon_pid" 2>/dev/null || true
    wait 2>/dev/null || true
  }
  trap cleanup EXIT INT TERM

  export GRPC_PORT="$GRPC_PORT"

  local pidfile="$RESULTS/grpc.pid"
  rm -f "$pidfile"
  export PID_FILE="$pidfile"

  log "Démarrage gRPC → $RESULTS/grpc-service.log"
  "$BIN_GRPC" > "$RESULTS/grpc-service.log" 2> "$RESULTS/grpc-service-err.log" &
  grpc_bash_pid=$!
  unset PID_FILE

  for _ in 1 2 3 4 5 6 7 8 9 10; do
    [ -s "$pidfile" ] && break
    sleep 0.3
  done
  grpc_pid="$(cat "$pidfile" 2>/dev/null || echo "")"
  if [ -z "$grpc_pid" ]; then
    die "gRPC n'a pas écrit son PID dans $pidfile"
  fi
  sleep 0.5

  log "Démarrage monitor PID=$grpc_pid → $RESULTS/$monfile (dur=$dur)"
  "$BIN_MONITOR" -pid "$grpc_pid" -duration "$dur" -interval 1s -out "$RESULTS/$monfile" &
  mon_pid=$!

  log "Lancement ghz : $ghz_script"
  GRPC_ADDR="$GRPC_ADDR" bash "$ghz_script"
  local rc=$?

  cleanup
  trap - EXIT INT TERM
  return $rc
}

scenario="${1:-all}"

case "$scenario" in
  build) build_all ;;
  sysinfo) collect_sysinfo ;;
  payload) build_all; regen_payload ;;

  rest-a) build_all; run_rest_bench benchmark/scripts/k6-rest-a-read.js monitor-rest-a.csv 30s 0 ;;
  rest-b) build_all; run_rest_bench benchmark/scripts/k6-rest-b-write.js monitor-rest-b.csv 30s 0 ;;
  rest-c) build_all; run_rest_bench benchmark/scripts/k6-rest-c-ramp.js monitor-rest-c.csv 300s 0 ;;
  rest-gzip) build_all; run_rest_bench benchmark/scripts/k6-rest-a-read-gzip.js monitor-rest-a-gzip.csv 30s 1 ;;

  grpc-a) build_all; run_grpc_bench benchmark/scripts/grpc-a-read.sh  monitor-grpc-a.csv 60s ;;
  grpc-b) build_all; run_grpc_bench benchmark/scripts/grpc-b-write.sh monitor-grpc-b.csv 60s ;;
  grpc-c) build_all; run_grpc_bench benchmark/scripts/grpc-c-ramp.sh  monitor-grpc-c.csv 300s ;;

  all)
    build_all
    collect_sysinfo
    regen_payload
    run_rest_bench benchmark/scripts/k6-rest-a-read.js  monitor-rest-a.csv 30s 0 || die "rest-a"
    run_rest_bench benchmark/scripts/k6-rest-b-write.js monitor-rest-b.csv 30s 0 || die "rest-b"
    run_rest_bench benchmark/scripts/k6-rest-c-ramp.js  monitor-rest-c.csv 300s 0 || die "rest-c"
    run_grpc_bench benchmark/scripts/grpc-a-read.sh   monitor-grpc-a.csv 60s    || die "grpc-a"
    run_grpc_bench benchmark/scripts/grpc-b-write.sh  monitor-grpc-b.csv 60s    || die "grpc-b"
    run_grpc_bench benchmark/scripts/grpc-c-ramp.sh   monitor-grpc-c.csv 300s   || die "grpc-c"
    run_rest_bench benchmark/scripts/k6-rest-a-read-gzip.js monitor-rest-a-gzip.csv 30s 1 || die "rest-gzip"
    log "Tous les scénarios terminés. Résultats : $RESULTS/"
    ;;

  *) die "scénario inconnu : $scenario (voir l'en-tête du script)" ;;
esac
