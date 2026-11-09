#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
[[ $EUID -eq 0 ]] || { echo 'Run this installer as root.' >&2; exit 1; }
command -v systemctl >/dev/null
test -x "$root/build/sober-tunnel" || { echo 'Build sober-tunnel first: bash scripts/build.sh' >&2; exit 1; }
test -x "$root/build/sober-tunnel-mimic" || { echo 'Build the native tools first: bash scripts/build.sh' >&2; exit 1; }
# Install a previously built local artifact. Never run upstream installers.
install -d -m 0755 /opt/sober-tunnel /etc/sober-tunnel
install -d -m 0700 /etc/sober-tunnel/mimic
install -m 0755 "$root/build/sober-tunnel" /usr/local/bin/sober-tunnel.new
mv -f /usr/local/bin/sober-tunnel.new /usr/local/bin/sober-tunnel
install -m 0755 "$root/build/sober-tunnel-mimic" /usr/local/bin/sober-tunnel-mimic.new
mv -f /usr/local/bin/sober-tunnel-mimic.new /usr/local/bin/sober-tunnel-mimic
install -m 0644 "$root/NOTICE" "$root/README.md" /opt/sober-tunnel/
install -d -m 0755 /opt/sober-tunnel/licenses
install -m 0644 "$root/engines/backpack/LICENSE" /opt/sober-tunnel/licenses/BackPack-AGPL-3.0.txt
install -m 0644 "$root/engines/hedioum/LICENSE" /opt/sober-tunnel/licenses/Hedioum-GPL-3.0.txt
if [[ "${SOBER_ENGINE_ONLY:-0}" == 1 ]]; then
  echo 'Installed sober-tunnel engines. Management stays on the central panel.'
  exit 0
fi
if [[ ! -f /etc/sober-tunnel/webui.json ]]; then
  (umask 077; printf '%s\n' '{"https":true}' > /etc/sober-tunnel/webui.json)
fi
cat > /etc/systemd/system/sober-tunnel-webui.service <<'UNIT'
[Unit]
Description=sober-tunnel unified management panel
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sober-tunnel --webui
Restart=on-failure
RestartSec=3
LimitNOFILE=1048576
UMask=0077

[Install]
WantedBy=multi-user.target
UNIT
cat > /etc/systemd/system/sober-tunnel-monitor.service <<'UNIT'
[Unit]
Description=sober-tunnel monitoring
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sober-tunnel --monitor
Restart=on-failure
RestartSec=3
LimitNOFILE=1048576
UMask=0077

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable sober-tunnel-webui.service sober-tunnel-monitor.service
systemctl restart sober-tunnel-webui.service sober-tunnel-monitor.service
echo 'Installed sober-tunnel. Run sudo sober-tunnel to view the panel address and login credentials.'
