#!/usr/bin/env bash
# Scénario C — montée en charge gRPC (ghz).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
ADDR="${GRPC_ADDR:-localhost:9090}"
mkdir -p benchmark/results
for c in 10 25 50 75 100; do
  ID="$(go run ./benchmark/cmd/seedsensor "$ADDR")"
  ghz --insecure \
    --proto="$ROOT/grpc-service/proto/sensor.proto" \
    --call=benchlab.sensor.v1.SensorService/GetSensor \
    -d "{\"id\":\"${ID}\"}" \
    -n 5000 -c "$c" --connections="$c" \
    --format=json \
    "$ADDR" | tee "$ROOT/benchmark/results/ghz-grpc-c-${c}.json" | "$ROOT/bin/ghzsummary$(go env GOEXE)"
done
