# v0.4.0 validation — 2026-11-21

## Engine and management checks

[Run 44899809766](https://github.com/amirsober-cloud/sober-tunnel/actions/runs/44899809766)
passed the complete Linux suites for both engines, native manager/panel contracts,
JavaScript syntax/form checks, continuation race tests and 20 repeated runs of
25 simultaneous WebSocket mux connections. It includes the HTTP upgrade boundary
fix. The earlier pre-fix full run reproduced WebSocket mux failures; a small
standalone test now also forces HTTP headers and the first raw frame to arrive
in one write, ensuring the HTTP parser cannot consume that frame.

The Windows checks passed root package tests, Camouflage resume/ingress/egress/
pool/config tests and vet for Camouflage plus root. These do not substitute for
the Linux checks. The host has no WSL and the main engine uses Linux socket APIs.

## Raw carriers and compatibility

[Run 44899987400](https://github.com/amirsober-cloud/sober-tunnel/actions/runs/44899987400)
built both static executables. Its isolated network jobs passed:

- Direct PCK: a byte-identical 5 MB transfer, UDP datagrams up to 30,000 bytes,
  and half-close followed by a complete reply.
- IPv4 GRE and IPv6 GRE: no packet loss in the fixture ping and identical 5 MB
  transfers. GRE provides no payload encryption.
- Reverse PCK: identical 2 MB TCP transfer and UDP up to 8,000 bytes; both TCP
  and UDP returned after each interface was disabled for six seconds and
  restored, without restarting the engine processes.
- Mixed v0.3.2/current peers in both directions on the reverse transports, and
  persistent refusal of the legacy handshake after proving a new peer.

Both downloaded candidate ELFs report Go 1.26.11. ELF program headers contain
neither a dynamic interpreter nor a dynamic-link segment. The main engine links
verified libpcap 1.11.0 statically; Camouflage is built with CGO disabled.

## Continuation scope

Tests cover repeated lane loss, flow-control backpressure, deadlines, duplicate
suppression, invalid ACKs, unknown IDs and reply replay after FIN. The real
loopback SSH mimic + authenticated encryption + Yamux test preserves exact
payload across a dropped physical session and opens the backend once.
Continuation remains opt-in, bounded to 128 KiB per direction and 256 foreign
streams per configured identity, with a 60-second disconnected grace period.
It does not persist across process restart, preserve UDP or establish throughput
or DPI resistance on the Iran route.

## Browser workflow

[Run 44902004543](https://github.com/amirsober-cloud/sober-tunnel/actions/runs/44902004543)
passed the real Chromium workflow in a disposable Debian container. It checked
login, Overview rendering including Camouflage statistics, engine selection,
native preview/pairing round-trip, x-ui export, asynchronous diagnostics and
zero page runtime exceptions. The Python fixture also started the independent
Camouflage executable, enabled continuation on the Iran node and successfully
carried HTTPS through its local SOCKS listener. No production config was changed.

The first browser run found an undefined showcase variable that interrupted
Overview rendering after the merge. The overview now keeps that upstream content
disabled and renders the traffic fields correctly; the browser check explicitly
waits for those fields so a similar startup exception cannot be missed.

## TLS traffic accounting

[Run 44903158551](https://github.com/amirsober-cloud/sober-tunnel/actions/runs/44903158551)
passed the native browser fixture with Resume enabled and explicitly verified
nonzero authenticated TLS sent/received counters in the combined Overview totals.
TLS authentication/decoy paths are counted only after authentication succeeds;
SSH retains its existing AEAD counters, avoiding a second count.
The affected securestream, mimic, ingress and egress suites also passed five
local repeats. The release workflow runs the security/mimic suites with races.

## Publication gate

The signed-release workflow requires the complete
[final validation run 44903531320](https://github.com/amirsober-cloud/sober-tunnel/actions/runs/44903531320)
to succeed and checks that runtime source/build inputs match that tested commit
before signing or publishing. Documentation and workflow changes may follow it;
the capture-builder change is comments only. Release packaging verifies binary
versions and includes both executables and licenses. The workflow then checks
the published archive and Ed25519 signature through the actual updater verifier.
No production service or configuration is modified by these validation jobs.