# Independent Camouflage core and traffic chart

Camouflage units execute `sober-tunnel-mimic --foreground --config <instance.json>`.
The management binary no longer imports the Camouflage engine. Legacy
`--mimic-config` execution delegates to the sibling native executable.

Start and Restart refresh the unit and reload systemd. Restart enables the unit
for subsequent boots. Units wait for network-online.target, restart after process
exit and disable systemd start throttling. Stop still disables boot activation.
The source installer migrates saved units and restarts only active instances.
For an existing installation updated another way, Start or Restart each desired
instance once to refresh its unit and boot activation.

The Camouflage screen plots sent and received plaintext securestream bytes per
second, sampled every two seconds, for the selected local or managed instance.
Totals reset with the core process. These are tunnel payload counters, including
internal securestream messages, not host interface usage or billed wire bytes.
Metrics are written atomically with mode 0600 beside the configuration; stale
samples are shown as unavailable and process changes reset rate calculation.
Managed servers require the updated build for the traffic RPC operation.

Validation: native mimicmgr and securestream tests passed; both Linux binaries
built; webui and node tests cross-compiled for Linux; git diff --check passed.
No Linux runtime, real reboot, or browser execution was available in this session.
The separate BackPack v2 migration remains pending.