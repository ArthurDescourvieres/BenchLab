#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
ADDR="${GRPC_ADDR:-localhost:9090}"
mkdir -p benchmark/results
ID="$(go run ./benchmark/cmd/seedsensor "$ADDR")"
ghz --insecure \
  --proto="$ROOT/grpc-service/proto/sensor.proto" \
  --call=benchlab.sensor.v1.SensorService/GetSensor \
  -d "{\"id\":\"${ID}\"}" \
  -n 1000 -c 10 --connections=10 \
  --format=json -o "$ROOT/benchmark/results/ghz-grpc-a.json" \
  "$ADDR"
