#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
ADDR="${GRPC_ADDR:-localhost:9090}"
mkdir -p benchmark/results
DATA='{"sensor":{"name":"Bench-ghz-B","type":"PRESSURE","location":"Line-2","unit":"bar","status":"ACTIVE","last_value":1.02,"last_reading_at":"2026-01-15T10:00:00Z"}}'
ghz --insecure \
  --proto="$ROOT/grpc-service/proto/sensor.proto" \
  --call=benchlab.sensor.v1.SensorService/CreateSensor \
  -d "$DATA" \
  -n 500 -c 5 --connections=5 \
  --format=json \
  "$ADDR" | tee "$ROOT/benchmark/results/ghz-grpc-b.json" | "$ROOT/bin/ghzsummary$(go env GOEXE)"
