#!/usr/bin/env bash
# Source installer used by local operators and the managed-server bootstrap.
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if [[ -f "$root/scripts/build.sh" && -f "$root/engines/backpack/go.mod" ]]; then
  bash "$root/scripts/build.sh"
  exec bash "$root/scripts/install.sh"
fi
command -v git >/dev/null || { echo 'Install git before running the source installer.' >&2; exit 1; }
command -v go >/dev/null || { echo 'Install Go 1.26.10 or newer before running the source installer.' >&2; exit 1; }
work="$(mktemp -d -t sober-tunnel-install.XXXXXXXX)"
trap 'rm -rf -- "$work"' EXIT
git clone --depth 1 https://github.com/amirsober-cloud/sober-tunnel.git "$work/source"
bash "$work/source/scripts/build.sh"
bash "$work/source/scripts/install.sh"
