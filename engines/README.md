# Integrated engines

Both source trees are included so a checkout contains the complete product.
Their Go module paths remain intact for internal imports and provenance.

| Source | Pinned commit | Original license |
|---|---|---|
| [BackPack](https://github.com/AminMGMT/BackPack) | 9d58f8449ac04705b2b47f4869dc9f6db0e034af | AGPL-3.0 + NOTICE |
| [Hedioum](https://github.com/hedioum/Hedioum-Pool-Tunnel) | b0f237f45adff84492a894cc19097e51121c72c8 | GPL-3.0 |

Local changes expose Hedioum as a public Go engine package and compile it into
the main sober-tunnel binary. BackPack provides the existing transport engines
and management panel; a new admin-only Camouflage section manages Hedioum
instances. There is one binary and one panel, with one process per instance.
Wire protocols remain separate and unchanged; this is a shared control plane,
not an invented claim of interoperability between their mux protocols.

The integration adds isolated config/certificate paths, configuration checks,
service management, a configuration editor, own product identity and build
scripts. Upstream installers and release workflows are reference material;
use the scripts at the repository root to build/install this distribution.
Do not run upstream install.sh to install the combined product.

The product's GitHub release target is amirsober-cloud/sober-tunnel. Online
updates require its own signing key stamped at build time. Upstream release
keys and binaries must never overwrite the combined product.
