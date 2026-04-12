set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export PATH="$(go env GOPATH)/bin:$PATH"
if ! command -v protoc >/dev/null 2>&1; then
  echo "protoc est requis (https://grpc.io/docs/protoc-installation/)." >&2
  exit 1
fi
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
protoc -I grpc-service/proto \
  --go_out=. --go_opt=module=github.com/ArthurDescourvieres/BenchLab \
  --go-grpc_out=. --go-grpc_opt=module=github.com/ArthurDescourvieres/BenchLab \
  grpc-service/proto/sensor.proto
echo "OK: grpc-service/gen/benchlab/sensor/v1/"
