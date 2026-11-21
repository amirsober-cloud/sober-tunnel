# v2 integration and TCP continuation

This work is on `work/v2-security-resume` until the complete Linux validation
and signed release pass. Do not install a development branch on a live pair.

## SoberLinkPack data plane

The upstream BackPack v2.0.0 engine is integrated with this distribution's
management paths, signed updater and independent Camouflage executable.
The plain reverse transports now prove both peers with fresh challenges and
bind backend addresses to the authenticated session. Persisted peer capability
state prevents silently returning to the old handshake after a successful new
handshake. Plain transports and GRE still do not encrypt application traffic.
Use an encrypted transport when confidentiality is required.

PCK uses the new upstream paqet engine and the pinned kcp/smux forks. Upgrade
both PCK endpoints together; an old PCK endpoint is not wire compatible.
GRE and IPv6 GRE are layer-3 options, not encryption or a guarantee against DPI.
Source builds verify libpcap 1.11.0 against a pinned SHA-256 and link it statically.
The archive includes libpcap and fork licenses. Runtime hosts need no libpcap.

## Camouflage continuation

After upgrading BOTH ends, set `"resume": true` on the Iran `foreign_nodes`
entry, or use the Resume setting in the native setup/edit form. It is disabled
by default so existing peers and performance characteristics remain compatible.
It applies to TCP carried through SOCKS/TUN, not UDP.

Continuation attaches a new authenticated encrypted mux stream to the same
logical TCP connection. Sequence offsets, acknowledgements and retransmission
preserve byte order and suppress duplicate bytes. The foreign backend connection
is opened once. Unknown or expired stream IDs fail; they never recreate a
backend connection and replay the request. TCP half-close is preserved.

Each direction has a 128 KiB payload window, with 32 KiB maximum data frames.
The foreign registry holds at most 256 streams per configured authentication
identity (up to 64 MiB of payload buffering, plus allocation/protocol overhead).
Disconnected state expires after 60 seconds. It cannot survive a process restart,
a closed destination connection or an outage longer than the grace period.
The window bounds memory and can limit single-flow throughput on high-RTT paths;
measure on the real route before enabling it broadly. No speed or DPI guarantee
follows from a successful local recovery test.

## Validation

Local repeated tests cover a multi-megabyte echo across several lane losses,
backpressure, deadlines, half-close, invalid acknowledgements and unknown IDs.
A real loopback SSH mimic + encrypted stream + Yamux test drops the physical
session during transfer and checks exact payload and one backend open.
The CI workflow also runs the race detector, all engine packages, browser-code
checks and isolated PCK/GRE/IPv6 GRE network tests. Final CI outcomes must be
recorded before publishing a stable update.