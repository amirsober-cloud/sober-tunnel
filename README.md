# sober-tunnel

A unified management panel and executable for multiple tunnel engines.

- Reverse/direct port forwarding, TCP, WebSocket/TLS, KCP, QUIC, UDP,
  transport fallback, and a separate layer-3 engine from BackPack.
- Protocol camouflage, authenticated encrypted connections, dynamic pools,
  draining, SOCKS5 TCP/UDP, TUN, DNS and gateway mode from Hedioum.
- One web panel: the existing transport/management screens plus **Camouflage**
  for local or SSH-managed remote instances. The complete native JSON remains
  editable, including advanced fields. Both ends must run this combined build
  to manage camouflage through the fleet API.

The integrated entrypoint lives in `engines/backpack`; `engines/hedioum/engine`
exposes the second core to the same binary. Each tunnel runs in its own process.
The original `cmd/client` and `cmd/server` are retained as the earlier prototype;
they are not the entrypoint for the integrated product.

## نصب از GitHub

روی سرور Linux دارای systemd، با کاربر root اجرا کنید:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh)
```

نصب‌کننده کد را از همین مخزن می‌گیرد، پیش‌نیازها و نسخهٔ مناسب Go را آماده
می‌کند و پنل و مانیتور را نصب می‌کند. نصب خودکار پیش‌نیازها روی Debian/Ubuntu
پشتیبانی می‌شود؛ در توزیع‌های دیگر ابتدا `git`، `curl`، `tar`، `python3` و
`ca-certificates` را نصب کنید. ساخت اولیه ممکن است چند دقیقه طول بکشد.
بعد از نصب، دستور `sober-tunnel` آدرس پنل و مشخصات ورود را نشان می‌دهد.
HTTPS اولیه گواهی خودامضا دارد. این نسخه آزمایشی است.

موتور camouflage هم در همین پنل مدیریت می‌شود. نام ابزار مستقل آن در کد منبع
`sober-tunnel-mimic` است؛ نصب اصلی از باینری یکپارچهٔ `sober-tunnel` استفاده می‌کند.

## Build and install

Requires Go 1.26.10 or newer. Production management requires Linux + systemd.

```sh
bash scripts/build.sh
sudo bash scripts/install.sh
sudo sober-tunnel
```

On Windows, cross-build the Linux binary:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/build.ps1
```

Installation is an explicit step on the target Linux host; no upstream installer
is invoked. The CLI shows the generated panel login and secret URL. Normal
tunnels use the existing Tunnels screen. The Camouflage screen supports server
selection, JSON templates, validation/save, start/stop/restart/delete, and logs.
Configure the foreign exit first; copy its token into the Iran hub's node
configuration and set the actual foreign endpoint. A running camouflage instance
must be stopped before its config can be edited.

```sh
sober-tunnel --check-mimic --mimic-config /path/to/instance.json
sober-tunnel --mimic-config /path/to/instance.json
sober-tunnel -c /path/to/tunnel.toml
sober-tunnel --webui
```

Configs live under `/etc/sober-tunnel`; camouflage JSON and per-instance TLS
caches live under its `mimic/` directory. Backups include that directory.
Restoring a backup registers camouflage services but leaves starting them as an
explicit panel action. New installations do not migrate or alter existing
BackPack/Hedioum deployments, SSH settings, or firewall rules automatically.

Source builds have no online release signing key. The online updater targets
this repository and refuses installation until a sober-tunnel key is embedded
using `RELEASE_PUBLIC_KEY` (shell) or `-ReleasePublicKey` (PowerShell). Offline
installation from your own build remains available. A release archive must
contain `sober-tunnel`, named `sober-tunnel_linux_<arch>.tar.gz`.

## Validation status

The combined Linux build, selected backend and transport tests, 32 JavaScript
tests, and a real camouflage HTTPS request have passed. An experimental HTTPS
panel is running on the test VPS. See [integration-status.md](docs/integration-status.md)
for the exact matrix and remaining validation; all upstream modes have not yet
been exercised on a real route.

## Attribution and licenses

Based on BackPack by Amin Mohammadi (AminMGMT)
https://github.com/AminMGMT/BackPack

Includes [Hedioum-Pool-Tunnel](https://github.com/hedioum/Hedioum-Pool-Tunnel).
Their source licenses/notices are preserved. See [NOTICE](NOTICE) and
[engine provenance](engines/README.md). The original MIT license does not
relicense the incorporated GPL/AGPL components.
