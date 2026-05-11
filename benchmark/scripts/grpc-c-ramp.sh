#!/usr/bin/env bash
# Scénario C — montée en charge gRPC (ghz).
#
# ghz ne supporte pas nativement un ramp continu (montée progressive du nombre
# de connexions à l'intérieur d'une même exécution, cf. k6-grpc-c-ramp.js qui
# utilise des "stages"). On émule la montée par une série de runs indépendants
# à concurrence fixe : 10, 25, 50, 75, 100 connexions. Chaque run tape 5000
# requêtes à concurrence constante, et chaque résultat est écrit dans son
# propre fichier ghz-grpc-c-<c>.json. Les bornes (10 et 100) couvrent le même
# domaine que la consigne ; les paliers intermédiaires donnent en plus une
# vue par seuil de charge utile pour le rapport.
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
