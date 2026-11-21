# v0.3.2 — SoberLinkPack naming and bounded Camouflage fallback

The managed BackPack-derived engine is presented as SoberLinkPack, under the
Sober Tunnel product. Camouflage remains an independent executable. Existing
commands, service names, paths, TOML keys, protocol identifiers and module paths
remain compatible; original license attribution remains in NOTICE/license files.

Camouflage now starts another eligible endpoint after 1.2 seconds rather than
waiting for a blackholed primary's entire handshake budget. No more than four
attempts compete. A completed authenticated mux round trip is required to win;
losing attempts are cancelled and their sockets or sessions are closed. Genuine
failures update endpoint health; cancellation of a losing race does not. Sessions
closed before two sustained readiness checks now also count against path health.

Bandwidth shaping now accounts for every byte even when an application passes a
buffer larger than the limiter burst. Cancelled limiter waits no longer fall
through into a socket write. Short writes and partial byte counts are preserved.
This fixes correctness and pacing; it is not a claim of increased peak speed.

Regression coverage: blocked-primary fallback latency, loser cancellation,
empty/cancelled races, post-handshake blackholes, recovery hysteresis, oversized
write accounting and cancellation. Both Linux executables build; CI runs ingress
and pool race tests as well as the release's existing checks and signature test.
Real-route smoke check: the published v0.3.2 bundle was downloaded on Iran and
checked against GitHub's SHA256 archive digest. The native binary ran from an
isolated temporary directory with SOCKS port 18089, TUN/DNS/gateway disabled.
All six HTTPS requests through the existing SSH endpoint returned HTTP 200,
taking 0.140–0.175 seconds. The temporary process and files were cleaned up.
Production services/configuration were not changed. This is a smoke check, not
a sustained load benchmark. The SSH-only route is not expanded by updating.

Not included: upstream BackPack v2/PCK migration, new cryptographic handshakes,
stream resumption, automatic MTU changes or restored mixed endpoints. These need
independent compatibility, corruption/loss and two-server validation. See
engine-resilience-review.md for the research and remaining design work.