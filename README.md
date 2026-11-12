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

موتور camouflage از بخش **Tunnels → Add tunnel → Camouflage / Hedioum** ساخته
و مدیریت می‌شود. فرم‌های اصلی، Persona، توکن pairing، تست اتصال و سرعت و
تنظیمات آمادهٔ x-ui در پنل قرار دارند. ابزار native هم با نام
`sober-tunnel-mimic` نصب می‌شود؛ اجرای تونل‌ها از باینری یکپارچهٔ `sober-tunnel` است.
راهنمای اتصال از SOCKS محلی `127.0.0.1` در [راهنمای x-ui](docs/x-ui.md) آمده است.

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
selection, native setup/edit forms, pairing, diagnostics, preflight checks,
x-ui exports, advanced JSON, validation/save, start/stop/restart/delete, and logs.
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
contain both `sober-tunnel` and `sober-tunnel-mimic`, named
`sober-tunnel_linux_<arch>.tar.gz`. Build it with `bash scripts/package.sh`.
Updates and pre-update snapshots keep the two executables together. When
upgrading v0.1.4 or earlier, run the source installer once: that older updater
only knew about the main binary.

## Validation status

The Linux build, native package suite, transport end-to-end suite, JavaScript
checks and real Chromium workflow passed. Isolated runtime checks also carried
HTTPS through Xray → local SOCKS → camouflage, and through a real kernel TUN
with its DNS forwarder. See the [v0.1.5 validation report](docs/validation-2026-11-10.md)
and [integration status](docs/integration-status.md) for the exact evidence and
remaining live-route checks. No actual Iran server was available for those checks.

## Terminal camouflage management

Run `sudo sober-tunnel` and choose **11) Camouflage**, or **Manage → Camouflage / Hedioum**.
Set up the foreign exit first, export its pairing token, then set up the Iran hub.
The terminal menu uses the same named instances and systemd services as the web
panel. It exposes the native setup/edit flags (including mimic/persona settings),
node management, JSON import/view, start/stop/restart/delete, logs, pairing and
native probe/connection/speed/IP diagnostics. Preview validates settings before
saving; saving does not start a tunnel. Stop an instance before saving edits.
JSON previews and pairing output contain credentials; use a private terminal.
## Release installer

The one-line installer defaults to the published `v0.2.1-rc.1` bundle, verifies
its SHA256 checksum and installs both executables without downloading Go.
This prerelease currently provides the Linux amd64 archive only. `SOBER_VERSION`
selects another published tag. Use `SOBER_INSTALL_MODE=source` to explicitly
compile the checkout; source builds require the Go version specified in go.work.