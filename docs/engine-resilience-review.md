# Engine resilience and security review — 2026-11-21

## Implemented in v0.2.14: Camouflage runtime path health

Existing endpoint selection had dial failure cooldown, but successful admission
reset failure history and established-session failures never fed it. A path that
passes its first ping and later blackholes could therefore be chosen repeatedly.

The existing encrypted Yamux session is checked after randomized 8–13 second
intervals during probation, then 20–45 seconds after two successes. Each round
trip is bounded to eight seconds; failed checks close the unusable session and
record endpoint failure. Two failures trigger the existing 60 second cooldown,
with exponential backoff capped at ten minutes. Cooling endpoints are skipped
while eligible alternatives exist; if all endpoints are cooling, attempts remain
possible for emergency recovery. Two successful runtime checks clear failure
history. No external website is used and no user data is replayed. Active streams
on healthy sessions are not moved. A closed failed session cannot preserve its
in-flight streams: applications must reconnect.

Yamux connection writes and stream admission are bounded to eight seconds.
Changing the budget for very slow deployments may be needed; these defaults
must be evaluated on real routes. Heartbeat success is transport liveness, not
proof of sustained throughput, DNS health or successful backend connections.
This update does not restore any endpoints previously removed by the operator.

## Findings and next candidates for BOTH engines

1. Measure goodput, p50/p95 request latency, retransmissions, memory and recovery
   time under simultaneous traffic. Compare the same route and workload before
   changing pool counts, buffers, FEC or caps. No universal speed preset follows
   from an SSH handshake or a successful small HTTP request.
2. Camouflage: extend passive feedback to stream-open failures and classify
   local cancellation separately from transport failures. Add visibility of
   endpoint probation/cooldown to the panel before adding a separate controller.
3. BackPack: integrate upstream v2 mutual challenge/response and authenticated
   backend addresses, with persistent downgrade protection and compatibility
   tests, before expanding plaintext carrier use. This migration is implemented on the v2 validation branch; see [integration details](v2-integration.md).
4. Both: offer explicit certificate/name validation or SPKI pinning with a
   rotation workflow. Preserve and test existing token/channel-binding proofs;
   a self-signed certificate is not itself an identity. Do not adopt Rathole's
   insecure-verification patch as a security enhancement. WSS simple_auth needs
   distinct documentation because it sends a reusable token to the TLS terminator.
5. Both: prefer TLS 1.3 for newly designed modes and evaluate compatibility
   before changing existing TLS 1.2 fallback. Never lower security automatically
   to recover connectivity. ECH needs actual client/server/DNS/CDN support;
   it is not an arbitrary toggle for a bare-IP tunnel.
6. Both: audit the full observable flow, not just ClientHello. uTLS shapes
   ClientHello; ALPN, certificates, application behavior, traffic volume and
   connection lifetime still matter. Avoid random unsupported extensions,
   unconditional padding or obsolete browser profiles as claimed DPI fixes.
7. BackPack PCK: upstream v2 congestion pacing and receive-window fixes are a
   stronger candidate than manually enlarging socket buffers. They require the
   new engine, forked dependencies, libpcap build and coordinated migration.
8. Independent binaries: continue separate systemd supervision; use restricted
   file permissions and least privileges compatible with TUN/raw sockets.
   Any capability reduction needs Linux runtime checks, not a blanket unit edit.

## Source review

Rathole ZIP supplies manager code and pinned upstream patches, not the complete
Rathole core. Its adaptive thresholds/cooldown are useful design evidence; no
Rathole code was copied into this change.

Primary references consulted:
- https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier
- https://github.com/hashicorp/yamux/blob/master/mux.go
- https://github.com/hashicorp/yamux/blob/master/session.go
- https://github.com/refraction-networking/utls/blob/master/README.md
- https://www.rfc-editor.org/rfc/rfc9325.html
- https://www.rfc-editor.org/rfc/rfc9852.html
- https://www.rfc-editor.org/rfc/rfc9849.html

No change here establishes resistance to DPI, or proves that observed packet loss
was censorship. Real-route results must distinguish observation from inference.