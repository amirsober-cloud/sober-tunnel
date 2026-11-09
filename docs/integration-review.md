# Hedioum / BackPack integration review

Reviewed on 2026-11-09. This is a source review and proposed design; neither
upstream project was built or run during this review.

## Source snapshots

- Hedioum: `b0f237f45adff84492a894cc19097e51121c72c8`
  https://github.com/hedioum/Hedioum-Pool-Tunnel/tree/b0f237f45adff84492a894cc19097e51121c72c8
- BackPack: `9d58f8449ac04705b2b47f4869dc9f6db0e034af`
  https://github.com/AminMGMT/BackPack/tree/9d58f8449ac04705b2b47f4869dc9f6db0e034af

## Findings

Hedioum's direct path is Iran hub -> foreign egress. Its mimic handshake
returns an authenticated encrypted net.Conn, then Yamux carries streams.
The pool selects the active connection with the fewest streams; bandwidth
measurements drive scaling and randomized rate limits. Connections drain
before retirement. TCP and UDP have separate physical connection pools.
UDP datagrams are framed inside streams over the masked TCP link; this is
not a native UDP or QUIC inter-node carrier.

BackPack's reverse path has the foreign node dial Iran, while users connect
to forwarded ports in Iran. It also has separate direct stream and layer-3
engines. Its TCP mux path uses smux and a control channel, pool admission
nonces, and mux-version negotiation. Management includes CLI, web UI,
monitoring, and per-tunnel systemd processes.

Both projects contain implementation and tests for these areas, but source
presence does not establish throughput, resilience, or effectiveness against DPI.

## Integration boundaries

1. Yamux, smux, and sober-tunnel's custom frames are different wire protocols.
   They cannot be interchanged or nested without an explicit design. Choose
   one mux for a new combined protocol, with adapters only for deliberate
   compatibility modes.
2. Hedioum's ServerMimic.Accept and ClientMimic.Dial return net.Conn, making
   the authenticated connection a useful integration boundary. Keep the
   mimic's authentication, replay checks, and decoy handling together.
3. Hedioum's pool is coupled to *yamux.Session. Port its policies behind a
   session interface; do not copy it unchanged onto a smux connection.
4. Hedioum's egress rejects loopback/private destinations. BackPack commonly
   forwards to a service on 127.0.0.1. Use explicit per-peer target mappings
   or an allowlist for port forwarding, while retaining public-only egress
   for general proxy mode. Do not globally remove the destination checks.
5. Socket direction and stream-opening direction must be separate concepts.
   In reverse mode, the foreign dialer can accept application streams from
   Iran; being the socket dialer does not make it the traffic origin.
6. Current BackPack transport fallback rotates whole transports; its docs
   require the same ordered chain on both ends. This does not establish
   seamless migration of existing streams between carriers.
7. Keep stream carriers and packet carriers separate. Layer-3 traffic needs
   explicit MTU, replay, packet-loss, and routing handling. Do not force its
   carriers through a reliable stream abstraction.
8. Both projects use internal packages and different module paths. A direct
   external import of those packages is not supported by Go's internal rules.
   Vendoring/porting or upstream extraction into public packages is needed.
9. Hedioum requests Go 1.26.10; BackPack requests Go 1.26.9. The local project
   currently declares 1.26.5. Dependencies and toolchain need a coordinated
   build check; do not mechanically concatenate the go.mod files.

## Proposed implementation

For the combined feature set, use BackPack's engine/management boundaries as
the starting design and add Hedioum's authenticated camouflage layer and pool
policies. Preserve sober-tunnel as the product name and integration workspace.
The user has not yet selected whether the full management UI is in scope.

The stream pipeline should be:

    forwarded port / SOCKS
      -> target policy
      -> session pool (health, selection, capacity, draining)
      -> one mux per physical connection
      -> authenticated transport (mimic, TLS, or another chosen carrier)
      -> socket

Implement one direct TCP mapping first, then reverse mode. Add scaling and
draining, UDP relay, carrier fallback, and management integration in that
order. Keep the layer-3 engine a separate milestone.

For a new stream protocol, smux v2 is a candidate because BackPack already
uses smux and v2 provides per-stream flow control. This is a proposed choice,
not a claim of compatibility with either unmodified upstream binary.

## Acceptance checks

- End-to-end TCP request and response through the actual entrypoints.
- Concurrent streams on one physical connection, including a slow reader.
- Reverse mode with the foreign node initiating the physical connection.
- Backend refusal, EOF, half-close policy, and tunnel failure cleanup.
- Wrong token, replay, handshake timeout, and unauthorized destination.
- No new streams on draining sessions; active streams finish within policy.
- Replacement of failed sessions without reviving dead ones.
- UDP datagram boundaries, idle cleanup, and bounded congestion behavior.
- Carrier fallback for new connections, with existing-stream behavior stated.
- Linux integration checks for routing/TUN/systemd; local Windows tests alone
  cannot validate those paths.

## Source notices

Hedioum supplies GPL-3.0; BackPack supplies AGPL-3.0 plus NOTICE and trademark
terms. Before incorporating upstream code, preserve the source licenses,
copyright notices, and applicable attribution. BackPack's NOTICE specifies:

    Based on BackPack by Amin Mohammadi (AminMGMT)
    https://github.com/AminMGMT/BackPack

Its trademark policy requires modified distributions to use their own name
and branding. Review the combined distribution's license obligations before
publication. This review does not change the local project's LICENSE.

## Files inspected

Hedioum: go.mod, LICENSE, internal/mimic/mimic.go,
internal/securestream/securestream.go, internal/pool/session.go,
internal/pool/manager.go, internal/ingress/dial.go,
internal/egress/server.go, internal/tunproto/tunproto.go.

BackPack: go.mod, LICENSE, NOTICE, TRADEMARK.md, docs/architecture.md,
docs/transport-fallback.md, internal/client/transport/tcpmux.go,
internal/server/transport/tcpmux.go, internal/utils/network/mux.go,
internal/tunnel/direct/session.go, internal/tunnel/direct/handshake.go.
