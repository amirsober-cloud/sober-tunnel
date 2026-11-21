#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$root/build"
export CGO_ENABLED=1 GOOS=linux GOARCH="${GOARCH:-$(go env GOARCH)}"
export GOMAXPROCS="${GOMAXPROCS:-1}" GOMEMLIMIT="${GOMEMLIMIT:-350MiB}"
flags='-s -w'
if [[ -z "${RELEASE_PUBLIC_KEY:-}" && -f "$root/docs/release-public-key.txt" ]]; then
  RELEASE_PUBLIC_KEY="$(cat "$root/docs/release-public-key.txt")"
fi
if [[ -n "${RELEASE_PUBLIC_KEY:-}" ]]; then
  flags+=" -X github.com/backpack/backpack/internal/app.ReleasePublicKey=$RELEASE_PUBLIC_KEY"
fi
cd "$root/engines/backpack"
pcap_prefix="${LIBPCAP_PREFIX:-$HOME/.cache/sober-tunnel-libpcap}"
bash scripts/build-libpcap.sh "$pcap_prefix"
export CGO_CFLAGS="-I$pcap_prefix/include ${CGO_CFLAGS:-}"
export CGO_LDFLAGS="-L$pcap_prefix/lib -lpcap ${CGO_LDFLAGS:-}"
go build -tags netgo,osusergo -p "${SOBER_BUILD_PARALLELISM:-1}" -trimpath -ldflags "$flags -linkmode external -extldflags -static" -o "$root/build/sober-tunnel" .
export CGO_ENABLED=0
cd "$root/engines/camouflage"
go build -p "${SOBER_BUILD_PARALLELISM:-1}" -trimpath -ldflags '-s -w' -o "$root/build/sober-tunnel-mimic" ./cmd/camouflage
