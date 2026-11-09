#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$root/build"
export CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH:-amd64}"
flags='-s -w'
if [[ -n "${RELEASE_PUBLIC_KEY:-}" ]]; then
  flags+=" -X github.com/backpack/backpack/internal/app.ReleasePublicKey=$RELEASE_PUBLIC_KEY"
fi
cd "$root/engines/backpack"
go build -trimpath -ldflags "$flags" -o "$root/build/sober-tunnel" .
