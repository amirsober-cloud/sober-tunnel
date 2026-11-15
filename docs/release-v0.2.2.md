# v0.2.2 — Fix updates from menu 8

The previous release advertised updates in the terminal but had no release key,
so selecting menu 8 refused installation before downloading anything.

- The combined binary now embeds sober-tunnel's own Ed25519 public key.
- Release checksums are signed using the repository's encrypted Actions secret.
- Publication fails if that secret does not match the embedded public key.
- Online updates retain mandatory signature/checksum verification, snapshots,
  health checks and rollback. Both executables remain together.
- Added the missing root VERSION fallback and refreshed the native version file.
- Missing-key errors include a recovery command and do not suggest that changing
  the network relay can fix the signing configuration.

One-time upgrade from v0.2.1 or older (as root on Linux):

```sh
SOBER_VERSION=v0.2.2 bash <(curl -fsSL https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh)
```

Then close the old menu and run `sudo sober-tunnel` again. Subsequent releases
can be installed from **8 → Check For Updates**. An already-running older binary
cannot acquire the new signing key without this initial upgrade.

Includes Linux amd64 executables, licenses, SHA256SUMS and SHA256SUMS.sig.
The release workflow runs Linux build checks, root tests, multiplexing race
checks, native integration tests, and signing/tampering/key-mismatch regressions.
Live geographically separate server and systemd checks are not repeated here.
