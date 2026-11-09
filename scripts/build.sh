#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$root/build"
export CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH:-$(go env GOARCH)}"
export GOMAXPROCS="${GOMAXPROCS:-1}" GOMEMLIMIT="${GOMEMLIMIT:-350MiB}"
flags='-s -w'
if [[ -n "${RELEASE_PUBLIC_KEY:-}" ]]; then
  flags+=" -X github.com/backpack/backpack/internal/app.ReleasePublicKey=$RELEASE_PUBLIC_KEY"
fi
cd "$root/engines/backpack"
go build -p "${SOBER_BUILD_PARALLELISM:-1}" -trimpath -ldflags "$flags" -o "$root/build/sober-tunnel" .
