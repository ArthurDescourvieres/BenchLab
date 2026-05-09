#!/usr/bin/env bash
# Collecte automatique des conditions de test (machine, OS, versions des outils).
# Sortie : benchmark/results/system-info.json + benchmark/results/system-info.txt
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
mkdir -p benchmark/results

# Helpers
get_cmd_version() {
  local name="$1"
  shift
  if command -v "$name" >/dev/null 2>&1; then
    "$name" "$@" 2>&1 | head -n 1 | tr -d '\r'
  else
    echo "not installed"
  fi
}

# Hostname
HOSTNAME_VAL="$(hostname 2>/dev/null || echo unknown)"

# OS / kernel
if [ "$(uname -s 2>/dev/null)" = "Darwin" ]; then
  OS_VAL="$(sw_vers -productName 2>/dev/null) $(sw_vers -productVersion 2>/dev/null)"
elif [ -r /etc/os-release ]; then
  # shellcheck disable=SC1091
  . /etc/os-release
  OS_VAL="${PRETTY_NAME:-$(uname -sr)}"
else
  OS_VAL="$(uname -sr 2>/dev/null || echo unknown)"
fi
ARCH_VAL="$(uname -m 2>/dev/null || echo unknown)"

# CPU model + logical cores
CPU_MODEL="unknown"
CPU_CORES="unknown"
if [ -r /proc/cpuinfo ]; then
  CPU_MODEL="$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2- | sed 's/^ *//' | tr -d '\r' || echo unknown)"
  CPU_CORES="$(grep -c '^processor' /proc/cpuinfo || echo unknown)"
elif command -v sysctl >/dev/null 2>&1; then
  CPU_MODEL="$(sysctl -n machdep.cpu.brand_string 2>/dev/null || echo unknown)"
  CPU_CORES="$(sysctl -n hw.logicalcpu 2>/dev/null || echo unknown)"
fi

# RAM total (en Mo)
RAM_MB="unknown"
if [ -r /proc/meminfo ]; then
  KB="$(grep -m1 MemTotal /proc/meminfo | awk '{print $2}')"
  if [ -n "${KB:-}" ]; then
    RAM_MB="$((KB / 1024))"
  fi
elif command -v sysctl >/dev/null 2>&1; then
  BYTES="$(sysctl -n hw.memsize 2>/dev/null || echo "")"
  if [ -n "${BYTES:-}" ]; then
    RAM_MB="$((BYTES / 1024 / 1024))"
  fi
fi

GO_VERSION="$(get_cmd_version go version)"
K6_VERSION="$(get_cmd_version k6 version)"
GHZ_VERSION="$(get_cmd_version ghz --version)"
PROTOC_VERSION="$(get_cmd_version protoc --version)"

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

JSON_FILE="benchmark/results/system-info.json"
TXT_FILE="benchmark/results/system-info.txt"

cat > "$JSON_FILE" <<EOF
{
  "collected_at": "$TIMESTAMP",
  "hostname": "$HOSTNAME_VAL",
  "os": "$OS_VAL",
  "arch": "$ARCH_VAL",
  "cpu_model": "$CPU_MODEL",
  "cpu_logical_cores": "$CPU_CORES",
  "ram_mb": "$RAM_MB",
  "tools": {
    "go": "$GO_VERSION",
    "k6": "$K6_VERSION",
    "ghz": "$GHZ_VERSION",
    "protoc": "$PROTOC_VERSION"
  }
}
EOF

cat > "$TXT_FILE" <<EOF
BenchLab — Conditions de test (collecte automatique)
Date (UTC) : $TIMESTAMP

[Machine]
Hostname           : $HOSTNAME_VAL
OS                 : $OS_VAL
Architecture       : $ARCH_VAL
CPU (modèle)       : $CPU_MODEL
CPU (cœurs logiques): $CPU_CORES
RAM totale         : ${RAM_MB} Mo

[Outils]
Go     : $GO_VERSION
k6     : $K6_VERSION
ghz    : $GHZ_VERSION
protoc : $PROTOC_VERSION
EOF

echo "Wrote $JSON_FILE"
echo "Wrote $TXT_FILE"
