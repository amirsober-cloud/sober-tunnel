# x-ui / 3x-ui through the local camouflage bridge

x-ui and its existing Xray run on the **Iran server**. sober-tunnel supplies
the local Iran SOCKS5 bridge and the tunnel to the foreign exit. Select the Iran
server in the sober-tunnel panel when generating or checking x-ui settings.

Create a foreign exit in **Tunnels → Add tunnel → Camouflage / Hedioum**.
Select a Persona or explicit mimics and set ports that are free on that server.
Use a real domain for certificate-based camouflage when available. Build the
preview, save, run preflight checks and start it. Export its pairing token using
the pairing operation with the exit's public IP. Treat that token as a password.

On the Iran server, create an Iran hub. Paste the pairing token, enter a node
alias and an unused SOCKS port (for example 40001). Leave SOCKS bound to
127.0.0.1 for an x-ui process on the same host. The token supplies the actual
foreign IP, protocol ports, persona and authentication key. Save and start the
hub, then run Probe and Full connection test. Speed tests open their own tunnel
connection; they are not a measurement of x-ui client throughput.

In the hub's **Connect x-ui / 3x-ui** section, generate settings and choose the
foreign node. The exporter uses its saved SOCKS port and emits valid JSON for
one of two schemas: `settings.servers[]` (the 3x-ui schema) or the current Xray
`settings.address` / `settings.port` schema. Choose the format your installed
core supports. Merge this fragment into x-ui's Xray configuration; it is not a
replacement for the entire configuration. Replace the inbound-tag placeholder
with the tags of the inbounds whose traffic should use this tunnel. Keep
domestic/direct and other intentional exceptions ahead of that routing rule.

Use **Check x-ui's existing Xray on this Iran server** to detect a normal host
installation and run `xray run -test` on both exported SOCKS settings formats.
This reads the existing core's version and validates private temporary fragments;
it leaves the panel database, existing configuration and active Xray process
alone. A container or custom installation may require running the same validation
inside the container where that core is installed. Passing this configuration
check establishes schema compatibility; use the connection test for live egress.

The path is:

```text
Client → x-ui/Xray inbound on Iran → SOCKS outbound 127.0.0.1:40001
       → sober-tunnel Iran hub → camouflage tunnel → foreign exit → Internet
```

Use `AsIs` for routing and send destination hostnames to SOCKS for resolution
at the foreign exit. Configure Xray DNS to use the proxy as appropriate;
resolving names locally first can still leak DNS outside this path. The tunnel
cannot undo a DNS lookup already performed by Xray.

`127.0.0.1` names the current network namespace. A Docker bridge container
cannot reach a host listener through its own loopback address. Run x-ui/Xray
and the hub in the same namespace, or deliberately configure a reachable
private listener and network access. Do not expose an unauthenticated SOCKS
listener to the public Internet.

The native helper is installed as `sober-tunnel-mimic`; the managed instances
still run the combined `sober-tunnel` executable. The panel never needs a second
management panel. Native diagnostic jobs have a five-minute deadline, bounded
output and one active job per server. The panel shows the raw individual
checks; completion alone does not mean every endpoint passed.

Sources: [original camouflage and DNS behavior](https://github.com/hedioum/Hedioum-Pool-Tunnel/blob/main/docs/MIMICS.md),
[3x-ui SOCKS outbound schema](https://github.com/MHSanaei/3x-ui/blob/main/frontend/src/schemas/protocols/outbound/socks.ts),
[current Xray SOCKS outbound](https://xtls.github.io/en/config/outbounds/socks.html).
