# Integration status

Updated 2026-11-10 after testing on an Ubuntu 24.04 x86_64 VPS with Go 1.26.10.
The combined binary builds and the HTTPS panel and monitor are installed as
experimental systemd services. This is a tested integration preview, not a
claim that every upstream feature has been validated on a production route.

## Implemented

- Complete upstream source snapshots included under engines/, preserving module
  identities, source licenses, notices and attribution.
- One combined executable entrypoint (engines/backpack/main.go) linking both
  engine families through Hedioum's new public engine package.
- One panel, with the existing transport/management screens and a Camouflage
  page for isolated native JSON configurations. No second Hedioum panel starts.
- Local and SSH-fleet camouflage operations: list, read, validate/save,
  start/stop/restart/delete and bounded journal logs. The endpoint requires admin
  scope because it returns credentials and can enable gateway routing.
- Checked instance names, private atomic configuration writes, engine-side
  validation, rollback on a unit/reload failure, and stopped-before-edit policy.
- Per-instance certificate storage, systemd units, and shared backup inclusion.
  Restore recreates units and explicitly leaves camouflage activation to the
  operator. Restart-all includes active camouflage instances.
- Independent sober-tunnel binary/config/service paths and own UI name/icons.
- Linux build/install scripts and Windows cross-build script. Managed-server
  bootstrap installs engines only; it does not start another management panel.
- Online updater points at sober-tunnel and requires this distribution's own
  release signing key. Upstream keys cannot replace the combined executable.
- Public GitHub source installer bootstraps missing Debian/Ubuntu dependencies,
  verifies the official Go archive checksum, and builds the combined executable.
  Fresh panel configurations enable HTTPS. Standalone camouflage commands,
  services, container packaging and updater assets use `sober-tunnel-mimic`.
- v0.1.4 also builds/installs the native helper for the unified Tunnels workflow.
  Forms use its original setup/edit commands and flag validations, with Persona,
  all mimic ports, pairing, SOCKS/pool/bandwidth, TUN, DNS and gateway controls.
  Edits produce a validated private preview before an explicit save. Native
  probe/test/speed/IP checks run as bounded asynchronous jobs on local or
  SSH-managed servers. x-ui exports use the saved per-node SOCKS listener.
  Optional SSH/firewall/BBR actions require separate explicit confirmations.
- v0.1.5 integrates with x-ui's existing Xray on the selected Iran server. The
  optional check queries its version and validates both SOCKS outbound schemas;
  no x-ui database or active Xray configuration is changed. Native validation,
  zero-jitter persistence, diagnostic process locking and TUN test routing were
  corrected. Source install, archive update and rollback now keep both binaries
  together; standalone helper maintenance cannot remove half the installation.

## Initial Windows verification

Passed with the locally installed Go 1.26.5, with GOWORK=off:

```powershell
go test ./...
go test ./engines/backpack/internal/mimicmgr/manager.go ./engines/backpack/internal/mimicmgr/manager_test.go
go test ./engines/hedioum/config/config.go ./engines/hedioum/config/config_test.go ./engines/hedioum/config/sober_test.go
```

The explicit-file checks exercise the actual added source and tests without
changing either upstream module's required toolchain. Manager tests mock engine
validation/systemd. They do not establish that systemd works on Windows.

## Linux verification completed

- The Go 1.26.10 archive was verified against the official download checksum.
- Built the combined executable from the workspace with `go build -p 1
  -trimpath -o build/sober-tunnel ./engines/backpack`.
- Passed the integrated main-package test and the original sober-tunnel
  workspace tests under Go 1.26.10.
- Passed BackPack packages `internal/mimicmgr`, `internal/webui`,
  `internal/node`, `internal/app`, and `internal/control`.
- Passed `internal/tunnel/direct` and `internal/tunnel/l3` unit tests.
- Passed the end-to-end tests `TestEveryTransportEstablishesAndCarriesData`,
  `TestForwardedUDPWorksOnEveryTransport`, and `TestWrongTokenIsRejected`.
  The matrix covers TCP, TCP mux, stealth, WS, WSS, WS mux, WSS mux, KCP,
  QUIC and UDP. Tests ran inside a bounded Docker container, not on host ports.
- Passed Hedioum packages `config`, `internal/mimic`, `internal/securestream`,
  `internal/pool`, `internal/ingress`, `internal/egress`, `internal/tunproto`,
  and `internal/tundev`; its public `engine` package also compiled.
- Both sides of an actual camouflage tunnel ran using the combined executable
  in an isolated container. An HTTPS request through its SOCKS5 -> TLS mimic ->
  authenticated egress path returned HTTP 200 from example.com.
- All 32 Node tests in `internal/webui/paneltest/*.test.js` passed.
- v0.1.3 passed the camouflage CLI/config/sysutil/egress and panel/app/manager
  package tests. Its installed HTTPS panel accepted login and authenticated
  stats/camouflage API requests after restarting both systemd services.
- v0.1.4 passed native CLI/config and manager/node/panel package tests, including
  real native setup → pairing export → hub setup, token rotation, rejection of
  invalid pools without changing live config, occupied-port detection, x-ui
  export, and route-cleanup isolation. An original edit bug for a disabled HTTP
  decoy was fixed. All 32 existing JavaScript tests passed.
- In a disposable container, the actual panel API prepared/saved/started both
  engines, completed asynchronous Probe and download/upload speed jobs, exported
  x-ui settings, and sent an HTTPS request through the local SOCKS bridge and
  camouflage tunnel to example.com (HTTP 200). The systemctl adapter in this
  container started real engine processes; it did not emulate tunnel traffic.
  The v0.1.4 VPS panel and monitor also passed HTTPS login before the v0.1.5 update.
- Chromium exercised the real panel in the disposable container: Tunnels → Add
  tunnel → engine selection, foreign/hub form previews, pairing round-trip,
  parseable x-ui export and asynchronous Probe output all passed with zero
  runtime exceptions. A navigation teardown issue found by this check was fixed.
  A diagnostic also completed after its short-lived node RPC process exited.
  These checks are reproducible with `scripts/check-native-panel.py` and
  `scripts/check-native-browser.mjs` in a disposable container. The local Windows
  browser could not reach the VPS over its SSH forward; no conclusion about
  every external browser route is implied by the container check.
- The public one-line installer completed in a fresh Ubuntu 24.04 container,
  cloning anonymously, installing missing prerequisites and verified Go,
  building v0.1.3, and writing both units and a mode-0600 HTTPS configuration.
  That container used a systemctl stub; actual service startup was checked
  separately on the experimental VPS.
- The installed panel passed HTTPS login, admin-only camouflage API checks,
  asset serving, and invalid-instance-name rejection. Both panel and monitor
  services were active after installation.
- A separate headless Edge profile logged into the real panel through a local
  SSH forward, opened Camouflage, and observed no page error. Direct access
  from the test desktop was unstable, so this does not prove browser access on
  every external route. The panel certificate is currently self-signed.
- Fixed the branding script changing Go identifiers, preserved upstream doc
  verification stamps, documented the new route, and pinned source/script LF
  line endings with `.gitattributes` after cross-platform test failures.

## Remaining validation and product work

- The generated reference was refreshed and legacy backup permissions corrected.
  The long transport and management suites passed with suitable deadlines.
  The live-link matrix needed a higher container PID allowance for its 24
  concurrent tunnel pairs. Optional benchmarks and prolonged WAN soak/throughput
  tests are distinct from these checks; see the dated validation report.
- A real kernel TUN, DNS forwarding and source-routed HTTPS passed inside an
  isolated container. Transparent LAN gateway routing, RouterOS, MTU on real
  routes, raw-socket carriers, and geographically separate two-node operation
  still need live validation. No Iran host was available during this work.
- Signed binary releases, online updates and a fleet-wide upgrade rollout still
  need an actual sober-tunnel release and its own signing key. The repository
  and source installer are public; no binary release was published. Source
  installation does not establish a signed upgrade channel.
- The advanced JSON editor remains available alongside the native setup/edit
  forms. Pool events are available through the journal; they are not yet merged
  into the existing transport metrics/alerts charts.

## Recommended Linux validation

```sh
go test ./...
(cd engines/hedioum && go test -p 1 -parallel 1 -timeout 20m ./...)
(cd engines/backpack && go test -p 1 -parallel 1 -timeout 20m ./...)
bash scripts/build.sh
```

Use a disposable Linux host for root/systemd-dependent tests. Then run two-node
TCP/UDP, wrong-token/replay, reconnect/draining, reverse/direct, fallback, TUN,
and gateway checks. The source integration is not yet a production-ready
release until these build and runtime checks pass.
