# Scénario C gRPC — plusieurs paliers de concurrence
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
Set-Location $Root
$addr = if ($env:GRPC_ADDR) { $env:GRPC_ADDR } else { "localhost:9090" }
New-Item -ItemType Directory -Force -Path "benchmark/results" | Out-Null

foreach ($c in 10, 25, 50, 75, 100) {
  $out = & go run ./benchmark/cmd/seedsensor $addr 2>&1
  if ($LASTEXITCODE -ne 0) { throw "seedsensor: $out" }
  $id = ($out | Out-String).Trim()
  if (-not $id) { throw "seedsensor a échoué (serveur gRPC sur $addr ?)" }
  & ghz --insecure `
    --proto="$Root/grpc-service/proto/sensor.proto" `
    --call=benchlab.sensor.v1.SensorService/GetSensor `
    --data="{`"id`":`"$id`"}" `
    -n 5000 -c $c `
    --format=json -o "$Root/benchmark/results/ghz-grpc-c-$c.json" `
    $addr
}
