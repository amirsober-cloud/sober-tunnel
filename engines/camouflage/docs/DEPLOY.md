# Deploying Camouflage

The binary is fully self-sufficient: it installs its own systemd service, opens
the firewall, and writes its config. Installing via `install.sh` and running the
binary by hand are equivalent — the script only detects the architecture,
downloads the matching release, and runs the binary.

## Install order

> Install the **Foreign (egress)** node first — it prints the auth token the Iran
> hub needs.

## Option A — one-line bootstrap (GitHub reachable)

```bash
bash <(curl -s https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh)
```

This downloads the right binary, runs `sober-tunnel-mimic install`, and launches the
interactive wizard.

## Option B — manual (GitHub blocked)

Download the matching asset (`sober-tunnel-mimic` or `sober-tunnel-mimic-arm64`) on any
machine, copy it to the server, then:

```bash
chmod +x sober-tunnel-mimic
./sober-tunnel-mimic install      # self-copies to /usr/local/bin + installs the service
sober-tunnel-mimic                # interactive setup wizard
```

## Option C — fully non-interactive (automation / Ansible)

No wizard, no prompts.

**Foreign node:**
```bash
./sober-tunnel-mimic install
sober-tunnel-mimic setup-foreign --listen-port 22 --decoy-port 2022 --move-ssh
#   prints: Auth Token: <hex>
systemctl start sober-tunnel-mimic.service
```

**Iran hub:**
```bash
./sober-tunnel-mimic install
sober-tunnel-mimic setup-iran \
  --alias DE-01 --target <FOREIGN_IP>:22 --socks-port 40001 --token <hex> \
  --min 10 --max 20 --bw 8 --jitter 2
systemctl start sober-tunnel-mimic.service
sober-tunnel-mimic add-node --alias NL-02 --target <IP>:22 --socks-port 40002 --token <hex>
sober-tunnel-mimic remove-node --alias NL-02
```

## Ports

- `--listen-port` (default **22**): the public tunnel port. A non-22 port helps
  where outbound :22 is blocked, but weakens the SSH mimic — a warning is shown.
- `--decoy-port` (default **2022**): where the real `sshd` is relocated; the
  tunnel routes unauthenticated probes (and real admin SSH) there.
- `--egress-mode ipv4|ipv6|dual` and `--egress-bind-ip IP` control how the foreign
  node reaches the internet (default `ipv4`, no IPv6 identity leak).

## Updating

```bash
sober-tunnel-mimic update                       # download the latest release + rollback-safe swap
sober-tunnel-mimic update --file ./sober-tunnel-mimic   # install a locally-provided binary (GitHub blocked)
```

## Logs & debugging

The daemon logs to journald (structured, color-free):

```bash
journalctl -u sober-tunnel-mimic.service -f
journalctl -u sober-tunnel-mimic.service -p warning     # warnings and errors only
```

Turn on verbose diagnostics (e.g. to see why a connection dropped):

```bash
systemctl edit sober-tunnel-mimic.service   # add:  [Service]\nEnvironment=CAMOUFLAGE_LOG_LEVEL=debug
systemctl restart sober-tunnel-mimic.service
```

## Management

```bash
sober-tunnel-mimic version           # version + build (commit/date)
sober-tunnel-mimic uninstall --yes   # stop, disable, and remove everything
```
