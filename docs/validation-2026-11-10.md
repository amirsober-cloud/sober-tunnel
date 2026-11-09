# Validation of sober-tunnel v0.1.5

The intended topology uses **x-ui's own Xray on Iran**. Xray's SOCKS outbound
connects to the local Iran hub, usually `127.0.0.1:PORT`; that hub sends traffic
through camouflage to the foreign exit. The installer supplies sober-tunnel's
two binaries and uses the existing Xray installation for its optional core check.

## Changes verified

- Native configuration save is atomic/private. Explicit bandwidth jitter `0`
  stays zero through save/load, while omitted legacy values retain their default.
- Validation rejects wildcard SOCKS listener conflicts, malformed bind/interface
  names, missing endpoints, invalid gateway CIDRs and DNS enabled without TUN.
- The engine validates saved configuration again before start/restart. Preflight
  checks actual listeners even when systemd says a service is active.
- Diagnostic launches use a shared kernel lock across panel and node RPC
  processes. Detached workers survive completion of the short-lived node RPC.
- The TUN connection-test rule uses a priority before Linux's main table and
  removes only its own temporary route/rule.
- The new Xray check detects the core already supplied by a normal x-ui host
  installation, queries its version and tests both exported SOCKS JSON shapes.
  It does not read the x-ui database or change its active configuration/process.
- Release bundles contain both executables. The installer stages both before
  replacing either and restores the previous helper if the main replacement
  fails. Pre-update snapshots also contain the helper; older snapshots remain
  readable. The full uninstaller handles the owned native units and helper.

## Runtime evidence

Tests ran in bounded Docker containers on an Ubuntu 24.04 x86_64 VPS with Go
1.26.10. The panel, native helpers and engine processes were real. A small
systemctl adapter started/stopped those processes inside the disposable test
container; the installed VPS panel uses real systemd.

| Check | Observed result |
| --- | --- |
| Original Hedioum package suite | Passed, including config, CLI, pool, mimics, securestream, DNS, TUN helpers and pairing |
| BackPack package suite | Passed across the baseline and corrected-package reruns; generated config reference refreshed and legacy-backup private modes repaired |
| Management/live setup-link suite | Passed in 773 seconds; its 24 concurrent tunnel pairs needed a higher container PID allowance than the initial 256 limit |
| Paired executable installation | Passed missing-helper, version-mismatch and injected main-rename failure cases; rejected bundles left the installed pair intact |
| Native helper maintenance | Standalone update/uninstall refused inside a combined installation; owned-unit cleanup tests preserved unrelated service names |
| Full transport end-to-end suite | Passed in 457 seconds with a 20-minute package deadline; the original three-minute deadline was too short |
| Native setup/pairing/edit | Passed; invalid edits preserved saved config and previews caused no host service changes |
| Asynchronous Probe/speed/node RPC | Passed, including worker survival after RPC exit |
| Cross-version wire compatibility | All 20 transport/direction cases passed between v0.1.4 and v0.1.5 |
| Local SOCKS to camouflage exit | HTTPS to example.com returned 200 |
| Real kernel TUN | Interface created inside container; source-routed HTTPS returned 200 |
| DNS forwarder | Answer received through the TUN-address DNS forwarder using remote DNS-over-TCP through SOCKS |
| Route cleanup | Container policy rules matched the pre-test snapshot after cleanup |
| Xray `servers[]` SOCKS export | Config validated; actual Xray HTTP inbound → SOCKS → camouflage → HTTPS returned 200 |
| Xray `address`/`port` SOCKS export | Same validation and actual HTTP 200 result |
| Chromium panel workflow | Tunnels engine selection, native forms, pairing, x-ui JSON and asynchronous results passed with zero runtime exceptions |
| Public source installation | Fresh Ubuntu 24.04 container cloned public `main`, bootstrapped prerequisites/verified Go and installed both v0.1.5 executables, units and private HTTPS config |
| Experimental VPS deployment | Both installed v0.1.5 binaries matched the tested build checksums; real systemd services, HTTPS login, authenticated APIs and new panel assets passed |

The isolated test core was the checksum-verified official Xray **v26.11.9**.
Archive SHA-256:

```text
b651f7e1d9966ee2c1b1ee1bf8b52c72771824d3065dce96bbc8bcc19ed95f82
```

It stayed in a private test-tools directory mounted read-only into the disposable
container. The existing VPS Xray and Docker workloads were not replaced.

The final management rerun used one CPU, a 640-MiB memory limit, 896-MiB
memory-plus-swap limit and 1,024 PIDs. The upstream live-link matrix manages its
own 24-pair concurrency; Go's `-parallel 1` alone does not reduce that matrix.
The package deadline was 20 minutes. Optional environment-gated host/network
tests remain distinct from the normal package runs.

The fresh installer container used a systemctl stub; the experimental VPS
separately verified real panel/monitor startup. Its existing unrelated Docker
workload remained running. The source/code commit tested and deployed was
`8c7446c`; this report's final update only adds the installation evidence.

## Reproducing integration checks

Build both binaries with `bash scripts/build.sh`. Use a disposable Linux
container containing Python 3 and curl; copy the two binaries into its
`/usr/local/bin`, mount the source read-only, and run
`python3 /work/scripts/check-native-panel.py`. The script refuses ordinary host
execution and creates its configuration/services only inside that container.

Optional test environment variables:

- `SOBER_TEST_XRAY=/test-tools/xray`: also validate/run the existing-core and
  Xray outbound chain tests. `scripts/fetch-test-xray.py OUTPUT_DIRECTORY`
  downloads/checks the official test binary; mount that directory read-only.
- `SOBER_TEST_TUN=1`: additionally test a real kernel TUN and DNS forwarder.
  Give **only the disposable container** NET_ADMIN, `/dev/net/tun` and iproute2.
  Use its own network namespace; do not use host networking or privileged mode.
- `SOBER_BROWSER_CHECK=1`: add Node.js and Chromium to exercise the real panel.
- `SOBER_SCREENSHOT=/artifacts/panel.png`: optional screenshot of the cleared
  form after the browser checks, with no pairing/auth token shown.

## Practical limits

Both engine endpoints in these checks were on one VPS in an isolated network.
This proves the software path, not reachability or performance on a real
Iran-to-foreign route. No Iran host credentials were available. Transparent
LAN gateway behavior, RouterOS, path MTU on a real route, WAN throughput and
prolonged soak tests remain separate checks. Some optional benchmark/previous-
binary tests require their documented environment variables.

Sources: [Xray configuration test command](https://xtls.github.io/en/document/command),
[official SOCKS config implementation](https://github.com/XTLS/Xray-core/blob/main/infra/conf/socks.go),
[official test-core release](https://github.com/XTLS/Xray-core/releases/tag/v26.11.9).
