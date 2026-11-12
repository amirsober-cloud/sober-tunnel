#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 && $(uname -s) == Linux ]] || { echo 'Run as root on Linux.' >&2; exit 1; }
missing=()
for tool in curl git tar python3; do
  command -v "$tool" >/dev/null || missing+=("$tool")
done
if ((${#missing[@]})); then
  command -v apt-get >/dev/null || { echo "Install these prerequisites first: ${missing[*]} and ca-certificates" >&2; exit 1; }
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends "${missing[@]}" ca-certificates
fi
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
work="$(mktemp -d -t sober-tunnel-install.XXXXXXXX)"
trap 'rm -rf -- "$work"' EXIT
if [[ ! -f "$root/scripts/build.sh" || ! -f "$root/engines/backpack/go.mod" ]]; then
  git clone --depth 1 --branch main https://github.com/amirsober-cloud/sober-tunnel.git "$work/source"
  root="$work/source"
fi
# Prefer the published bundle: installing on a VPS must not require a Go SDK.
# Source compilation remains available with SOBER_INSTALL_MODE=source.
install_mode="${SOBER_INSTALL_MODE:-release}"
case "$install_mode" in
  release)
    case "$(uname -m)" in
      x86_64|amd64) arch=amd64 ;;
      aarch64|arm64) arch=arm64 ;;
      *) echo 'Release installation supports amd64 and arm64.' >&2; exit 1 ;;
    esac
    version="${SOBER_VERSION:-v0.2.1-rc.1}"
    [[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]] || { echo 'Invalid SOBER_VERSION.' >&2; exit 1; }
    archive="sober-tunnel_linux_${arch}.tar.gz"
    base="https://github.com/amirsober-cloud/sober-tunnel/releases/download/$version"
    echo "Downloading sober-tunnel $version for linux/$arch..."
    if ! curl --fail --silent --show-error --location --retry 3 "$base/$archive" -o "$work/$archive"; then
      echo "Release asset unavailable: $base/$archive" >&2
      echo 'This prerelease currently has an amd64 bundle. For source builds, use SOBER_INSTALL_MODE=source with a compatible Go toolchain.' >&2
      exit 1
    fi
    curl --fail --silent --show-error --location --retry 3 "$base/SHA256SUMS" -o "$work/SHA256SUMS"
    python3 - "$work/SHA256SUMS" "$work/$archive" "$archive" <<'PY'
import hashlib, pathlib, sys, tarfile
entries = [line.split() for line in pathlib.Path(sys.argv[1]).read_text().splitlines()]
expected = [fields[0] for fields in entries if len(fields) == 2 and fields[1].lstrip('*') == sys.argv[3]]
h = hashlib.sha256()
with open(sys.argv[2], 'rb') as f:
    for block in iter(lambda: f.read(1024 * 1024), b''):
        h.update(block)
if len(expected) != 1 or h.hexdigest() != expected[0]:
    raise SystemExit('Release archive checksum verification failed')
with tarfile.open(sys.argv[2], 'r:gz') as archive:
    members = {m.name: m for m in archive.getmembers()}
    for name in ('sober-tunnel', 'sober-tunnel-mimic'):
        if name not in members or not members[name].isfile():
            raise SystemExit('Release archive must contain both executable files')
PY
    mkdir -p "$root/build"
    tar -xzf "$work/$archive" -C "$root/build" sober-tunnel sober-tunnel-mimic
    chmod 0755 "$root/build/sober-tunnel" "$root/build/sober-tunnel-mimic"
    bash "$root/scripts/install.sh"
    if (($#)); then
      [[ ${1:-} == link && ${2:-} == apply ]] || { echo 'Supported install arguments: link apply ...' >&2; exit 1; }
      /usr/local/bin/sober-tunnel "$@"
    fi
    exit 0
    ;;
  source) ;;
  *) echo 'SOBER_INSTALL_MODE must be release or source.' >&2; exit 1 ;;
esac
required="$(awk '$1 == "go" { print $2; exit }' "$root/go.work")"
[[ $required =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Invalid required Go version.' >&2; exit 1; }
installed="$(GOTOOLCHAIN=local go version 2>/dev/null | awk '{sub(/^go/, "", $3); print $3}' || true)"
if [[ -z "$installed" || $(printf '%s\n' "$required" "$installed" | sort -V | head -n1) != "$required" ]]; then
  case "$(uname -m)" in
    x86_64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo 'Automatic Go installation supports amd64 and arm64; install Go manually.' >&2; exit 1 ;;
  esac
  archive="go${required}.linux-${arch}.tar.gz"
  curl --fail --silent --show-error --location --retry 3 "https://go.dev/dl/$archive" -o "$work/$archive"
  curl --fail --silent --show-error --location --retry 3 'https://go.dev/dl/?mode=json&include=all' -o "$work/go-releases.json"
  python3 - "$work/go-releases.json" "$work/$archive" "$archive" <<'PY'
import hashlib, json, sys
with open(sys.argv[1]) as f:
    releases = json.load(f)
expected = next((f['sha256'] for r in releases for f in r['files'] if f['filename'] == sys.argv[3]), None)
h = hashlib.sha256()
with open(sys.argv[2], 'rb') as f:
    for block in iter(lambda: f.read(1024 * 1024), b''):
        h.update(block)
if not expected or h.hexdigest() != expected:
    raise SystemExit('Official Go archive checksum verification failed')
PY
  toolchain="/opt/sober-tunnel/toolchains/go${required}-${arch}"
  install -d -m 0755 "$toolchain"
  tar -xzf "$work/$archive" -C "$toolchain"
  export PATH="$toolchain/go/bin:$PATH"
fi
export GOTOOLCHAIN=local
bash "$root/scripts/build.sh"
bash "$root/scripts/install.sh"
if (($#)); then
  [[ ${1:-} == link && ${2:-} == apply ]] || { echo 'Supported install arguments: link apply ...' >&2; exit 1; }
  /usr/local/bin/sober-tunnel "$@"
fi
