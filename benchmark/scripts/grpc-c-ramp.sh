#!/usr/bin/env bash
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
    -n 5000 -c "$c" \
    --format=json -o "$ROOT/benchmark/results/ghz-grpc-c-${c}.json" \
    "$ADDR"
done
