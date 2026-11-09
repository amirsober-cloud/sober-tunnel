# Integration status

Updated 2026-11-09 after testing on an Ubuntu 24.04 x86_64 VPS with Go 1.26.10.
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

- The complete upstream test suites and prolonged stress/throughput tests have
  not been run. Passing the selected matrix is not a performance guarantee.
- Kernel TUN/gateway routing, MTU on real routes, raw-socket carriers, and
  geographically separate two-node operation still need live validation.
- Signed binary releases, online updates and a fleet-wide upgrade rollout still
  need an actual sober-tunnel release and its own signing key. The repository
  and source installer are public; no binary release was published. Source
  installation does not establish a signed upgrade channel.
- Advanced Hedioum settings are currently exposed through the complete JSON
  editor, rather than separate guided forms for every individual field. Pool
  events are available through its journal; they are not yet merged into the
  existing transport metrics/alerts charts.

## Recommended Linux validation

```sh
go test ./...
(cd engines/hedioum && go test ./...)
(cd engines/backpack && go mod tidy && go test ./...)
bash scripts/build.sh
```

Use a disposable Linux host for root/systemd-dependent tests. Then run two-node
TCP/UDP, wrong-token/replay, reconnect/draining, reverse/direct, fallback, TUN,
and gateway checks. The source integration is not yet a production-ready
release until these build and runtime checks pass.
