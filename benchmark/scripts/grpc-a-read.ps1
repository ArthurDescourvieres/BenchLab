# Scénario A gRPC — nécessite : serveur `go run ./grpc-service` + `ghz` dans le PATH
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
Set-Location $Root
$addr = if ($env:GRPC_ADDR) { $env:GRPC_ADDR } else { "localhost:9090" }
New-Item -ItemType Directory -Force -Path "benchmark/results" | Out-Null

$out = & go run ./benchmark/cmd/seedsensor $addr 2>&1
if ($LASTEXITCODE -ne 0) { throw "seedsensor: $out" }
$id = ($out | Out-String).Trim()
if (-not $id) { throw "seedsensor n'a renvoyé aucun id (serveur gRPC démarré sur $addr ?)" }

& ghz --insecure `
  --proto="$Root/grpc-service/proto/sensor.proto" `
  --call=benchlab.sensor.v1.SensorService/GetSensor `
  --data="{`"id`":`"$id`"}" `
  -n 1000 -c 10 `
  --format=json -o "$Root/benchmark/results/ghz-grpc-a.json" `
  $addr
