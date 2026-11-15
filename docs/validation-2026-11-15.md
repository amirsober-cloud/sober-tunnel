# Local release validation — 2026-11-15

Version: `v0.2.1-rc.1`. Host: Windows amd64, Go 1.26.10.

## Completed

- Root module `go test ./...` passed with the existing prototype changes.
- `scripts/build.ps1` built both Linux amd64 executables successfully.
- `scripts/package.ps1` built `build/sober-tunnel_linux_amd64.tar.gz`, including
  both executables, LICENSE, NOTICE and both engine licenses. Archive contents
  were listed successfully and `build/SHA256SUMS` was generated.
- The camouflage instance-manager suite passed after explicitly skipping its
  Linux executable integration test on non-Linux hosts.
- Camouflage securestream and pool package tests passed; its engine package compiled.
- Linux test executables compiled for the integrated entrypoint, web panel,
  node, control and terminal menu packages. Compilation is not execution.
- Fixed the atomic-write test's reader shutdown so an early writer error always
  signals the reader to stop instead of leaving the test waiting indefinitely.

## Limitations

The Windows host has neither WSL installed nor Docker available. Linux service,
network and native pairing integration tests were not executed in this run.
Running the panel packages directly on Windows fails on Linux socket APIs.
Windows app/config tests also fail Unix file-permission assertions; the initial
app run timed out in the reader shutdown path described above.
After the shutdown fix, the targeted test terminates promptly and reports the
Windows rename error (`Access is denied`) rather than timing out. This is still
a failed Windows test, not a passing Linux runtime check.

The version remains a release candidate. No production deployment, public release,
signed updater, two-server route test or publication was performed. Existing
uncommitted prototype changes and local files were preserved.

Rebuild the archive on Windows with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/package.ps1 -Architecture amd64
```

For Linux execution checks, follow the commands in `integration-status.md` on a
disposable Linux host. The installer explicitly sets both extracted executables
to mode 0755, including archives created on Windows.
