#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$root/scripts/build.sh"
# Check the actual executables before creating or signing a release archive.
expected="$(tr -d '\r\n' < "$root/VERSION")"
main_version="$("$root/build/sober-tunnel" -v | head -n 1)"
native_version="$("$root/build/sober-tunnel-mimic" version)"
[[ "$main_version" == "$expected" ]] || { echo "Main binary version mismatch: $main_version" >&2; exit 1; }
[[ "$native_version" == "sober-tunnel-mimic $expected "* ]] || { echo "Native binary version mismatch: $native_version" >&2; exit 1; }
arch="${GOARCH:-$(go env GOARCH)}"
archive="$root/build/sober-tunnel_linux_${arch}.tar.gz"
pcap_prefix="${LIBPCAP_PREFIX:-$HOME/.cache/sober-tunnel-libpcap}"
cp "$pcap_prefix/LICENSE.libpcap" "$root/build/LICENSE.libpcap"
tar -czf "$archive" -C "$root/build" sober-tunnel sober-tunnel-mimic LICENSE.libpcap -C "$root" LICENSE NOTICE engines/backpack/LICENSE engines/camouflage/LICENSE engines/backpack/internal/thirdparty/kcp-go/LICENSE engines/backpack/internal/thirdparty/smux/LICENSE
(cd "$root/build"; sha256sum sober-tunnel_linux_*.tar.gz > SHA256SUMS)
echo "Created $archive with both binaries, licenses and SHA256SUMS."
