# v0.4.0 — SoberLinkPack v2 and Camouflage continuation

SoberLinkPack now includes the upstream v2 data plane with mutual challenge/
response authentication and authenticated backend addresses on plain reverse
transports, plus persistent downgrade protection. PCK uses the new paqet-based
engine and its congestion/window fixes. Independent Camouflage supervision,
traffic charts, existing SSH endpoint choices and signed updates are retained.

Camouflage adds optional TCP stream continuation across a lost encrypted mux
session. It preserves byte order and half-close without re-opening the foreign
backend. It is disabled by default; after upgrading both ends, enable Resume on
the Iran node if wanted. State expires after 60 seconds disconnected and is not
preserved across process restarts. UDP is outside this feature. Bounded windows
can limit single-flow throughput on high-latency paths; test on the actual route.

The update also fixes a WebSocket mux startup race that could lose the first raw
frame when it arrived with HTTP upgrade headers, and an Overview render error
introduced during v2 integration. Overview and the Camouflage screen retain the
sent/received charts and totals. Authenticated TLS mimics now feed the same counters as SSH.

**Migration:** upgrade BOTH PCK endpoints together: the old and new PCK wires
are incompatible. Migrate legacy layer-3 UDP/QUIC configs before updating; their reverse transport variants remain supported. Keep the existing SSH-only configuration on routes where it
has proved more stable. GRE and IPv6 GRE do not encrypt traffic. New engine
features do not imply DPI resistance or measured speed gains on an Iran route.

Both Linux amd64 executables ship together in the signed archive. libpcap is
pinned, checksum-verified and linked statically; runtime servers do not need a
compiler or libpcap package. Source builds require Go 1.26.11, a C compiler, flex
and bison. The distribution includes the additional dependency licenses.

See [migration and continuation details](v2-integration.md) and the
[validation report](validation-v0.4.0.md) for exact test coverage.