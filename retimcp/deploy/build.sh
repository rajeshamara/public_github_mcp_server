#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"
command -v go >/dev/null
mkdir -p bin
go test ./retimcp/...
go build -buildvcs=false -o bin/retimcp-gateway ./retimcp/cmd/gateway
