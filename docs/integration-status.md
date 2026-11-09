# Integration status

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

## Verification performed on this Windows workspace

Passed with the locally installed Go 1.26.5, with GOWORK=off:

```powershell
go test ./...
go test ./engines/backpack/internal/mimicmgr/manager.go ./engines/backpack/internal/mimicmgr/manager_test.go
go test ./engines/hedioum/config/config.go ./engines/hedioum/config/config_test.go ./engines/hedioum/config/sober_test.go
```

The explicit-file checks exercise the actual added source and tests without
changing either upstream module's required toolchain. Manager tests mock engine
validation/systemd. They do not establish that systemd works on Windows.

## Still unverified / release blockers

- The combined binary and complete upstream suites have NOT passed a build in
  this environment. Go 1.26.10 is required by the integrated modules. Automatic
  toolchain retrieval stalled on goproxy.cn; proxy.golang.org and goproxy.io zip
  requests returned 403; direct go.dev/dl and dl.google.com archive checks
  returned 404. Do not treat subset test success as a successful combined build.
- This host has no installed WSL distribution or Docker. Actual Linux systemd,
  TUN/gateway/MTU behavior and two-node network tests remain required.
- The panel's new view and full upstream JavaScript tests need a browser/Node
  validation pass. Source is wired into the existing embedded asset system.
- Combined dependency resolution and go.sum/go.work.sum updates need to finish
  on a host with access to the required toolchain/modules.
- Publish/install/update/fleet rollout need an actual sober-tunnel release and
  its own signing key. No release was published and no server was deployed.
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
