# Local recovery and validation — 2026-11-15

The interrupted working tree was preserved. The integrated product entrypoint
is `engines/backpack`; `cmd/client` and `cmd/server` are the earlier prototype.

Fixed prototype multiplexing: register pooled sessions, share a single
forwarder/read loop per connection, serialize frame writes in both directions,
keep stream reservations until forwarding ends, close failed/finished streams,
remove disconnected sessions, and stop probing live sockets by changing their
deadlines. Client destination validation and its example were updated.

Added a real TCP round-trip regression using 16 concurrent logical streams on
one pooled tunnel, checking payload isolation, reservation cleanup, removal of
closed sessions, and refused-destination EOF.

The Camouflage save regression now checks Unix mode 0600 only on platforms
that expose those bits. Serialization and temporary-file cleanup still run on
Windows, and Linux retains the original permission assertion.

Validated locally on Windows with Go 1.26.10:

- Root module tests and `go vet ./...`.
- Camouflage config, securestream and pool tests, plus BackPack mimicmgr tests.
- Cross-build of both Linux amd64 executables.
- Linux amd64 distribution archive with licenses and SHA256SUMS.

Race detection was unavailable because this Go environment has cgo disabled.
Linux systemd/network runtime checks were not rerun: neither WSL nor Docker
is installed on this desktop. Earlier Linux evidence is recorded separately;
the local cross-build does not establish live-route or production readiness.

Output: `build/sober-tunnel_linux_amd64.tar.gz` and `build/SHA256SUMS`.
No deployment or publication was performed.
