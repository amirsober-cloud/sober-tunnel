#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$root/scripts/build.sh"
arch="${GOARCH:-$(go env GOARCH)}"
archive="$root/build/sober-tunnel_linux_${arch}.tar.gz"
tar -czf "$archive" -C "$root/build" sober-tunnel sober-tunnel-mimic -C "$root" LICENSE NOTICE engines/backpack/LICENSE engines/hedioum/LICENSE
(cd "$root/build"; sha256sum sober-tunnel_linux_*.tar.gz > SHA256SUMS)
echo "Created $archive with both binaries, licenses and SHA256SUMS."
