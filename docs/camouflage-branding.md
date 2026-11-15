# Camouflage branding

The integrated engine is named **Camouflage** throughout the panel, terminal
menus, native help, options and distribution documentation. Its source lives in
`engines/camouflage`, with native commands in `cmd/camouflage`, and its Go module
is `github.com/amirsober-cloud/sober-tunnel/engines/camouflage`.

Build/package scripts, the workspace and both engine imports use those paths.
The distributed helper remains `sober-tunnel-mimic`; installed instance paths
and unit names stay within the existing sober-tunnel namespace.

Original contributor names and upstream links remain in NOTICE and the source
provenance table. Four historical cryptographic derivation strings also remain
unchanged so authentication and deterministic personas retain compatibility
with existing peers. They are protocol constants, not presentation labels.

The structured log override is now `CAMOUFLAGE_LOG_LEVEL`. Newly generated TUN
interface names use the `camouflage` prefix. Explicit interface names in saved
configuration files are retained.
