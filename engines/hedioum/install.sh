#!/usr/bin/env bash
# Compatibility entrypoint: install the unified sober-tunnel product.
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
if [[ -f "$root/install.sh" && -f "$root/go.work" ]]; then
  exec bash "$root/install.sh" "$@"
fi
work="$(mktemp -d -t sober-tunnel-bootstrap.XXXXXXXX)"
trap 'rm -rf -- "$work"' EXIT
curl -fsSL https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh -o "$work/install.sh"
bash "$work/install.sh" "$@"
