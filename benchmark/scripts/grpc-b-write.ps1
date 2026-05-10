# Scénario B gRPC — nécessite : serveur gRPC + `ghz`
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
Set-Location $Root
$addr = if ($env:GRPC_ADDR) { $env:GRPC_ADDR } else { "localhost:9090" }
New-Item -ItemType Directory -Force -Path "benchmark/results" | Out-Null

$data = '{"sensor":{"name":"Bench-ghz-B","type":"PRESSURE","location":"Line-2","unit":"bar","status":"ACTIVE","last_value":1.02,"last_reading_at":"2026-01-15T10:00:00Z"}}'

& ghz --insecure `
  --proto="$Root/grpc-service/proto/sensor.proto" `
  --call=benchlab.sensor.v1.SensorService/CreateSensor `
  --data=$data `
  -n 500 -c 5 `
  --format=json -o "$Root/benchmark/results/ghz-grpc-b.json" `
  $addr
