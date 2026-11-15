# v0.2.1-rc.2 — Camouflage branding

Camouflage is now the engine name in the web panel, terminal menus, native help
and documentation. The source module and command directory also use Camouflage
under this repository's namespace. Original source attribution and cryptographic
derivation constants are retained for provenance and peer compatibility.

- Includes both `sober-tunnel` and `sober-tunnel-mimic` for Linux amd64.
- The one-line installer selects this release and verifies its SHA256 checksum.
- Adds Windows release packaging and corrects platform-specific test handling.

Validation: both Linux executables built on GitHub Actions. Root module tests,
camouflage manager (including native setup/pairing/editing), securestream, pool,
persona and logging tests passed on Linux. Corresponding portable tests also
passed on Windows; terminal, web panel, native CLI and TUN test executables
compiled for Linux. Live systemd and real two-server route validation were not
repeated for this release.

This is an unsigned release candidate. The signed online updater remains
disabled. Install or update manually on both Linux servers as root:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh)
```

Existing saved instance names, config paths and sober-tunnel service names are
preserved. The log override is now `CAMOUFLAGE_LOG_LEVEL`; newly generated TUN
interface names use the `camouflage` prefix.
