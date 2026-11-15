# v0.2.1

Final release of the unified sober-tunnel management panel and Camouflage integration.

- One panel and terminal interface for both tunnel engine families.
- Camouflage setup, pairing, configuration, diagnostics, x-ui exports and lifecycle management.
- Both Linux amd64 executables and distribution licenses in one archive.
- Installer defaults to v0.2.1 and verifies the archive's SHA256 checksum.
- Fixes in the earlier client/server prototype for pooled sessions, concurrent
  multiplexed transfers, response handling and stream cleanup.

The release workflow builds both executables on Linux and checks the root module,
race detection for multiplexing, Camouflage configuration/securestream/pool/persona/logging,
and the integrated Camouflage manager before publishing artifacts.

Install or update manually on each Linux server as root:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh)
```

The archive supports Linux amd64. HTTPS initially uses a self-signed certificate.
This release is not cryptographically signed; the signed online updater remains
disabled. Live systemd, gateway and geographically separate two-server route checks
were not repeated for this release. Prior Linux runtime evidence and remaining
validation limits are recorded in docs/integration-status.md.
