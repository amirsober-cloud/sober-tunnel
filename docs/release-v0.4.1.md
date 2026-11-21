# v0.4.1 — Camouflage pool scheduling

Short sequential TCP connections used the first physical connection whenever
several connections had equal load. This concentrated traffic on one pipe even
with a warm pool. Equal-load choices now rotate, while lower-load connections
remain preferred and closed/draining connections remain excluded.

Stream admission now happens outside the pool lock. A stalled transport cannot
block pool maintenance while opening a stream. Pending admissions are reserved
so concurrent selection and draining account for a connection already being
opened. Existing streams stay on their physical connection; this change adds
no periodic restart, replay, protocol switch or new traffic probe.

Both cores report v0.4.1 and ship in the signed distribution. No new foreign-side
wire feature or setting is required for this scheduling fix. Instance configuration,
SSH endpoint selection, bandwidth limits, credentials and x-ui routing are retained.
The v0.4.0 PCK migration requirements still apply when upgrading older versions.

Regression tests reproduce the first-pipe concentration and pool-lock stall on
the original implementation and pass with the fix. The affected Linux suites and
race checks passed before release preparation. A temporary 12-request comparison
on the Iran server observed median/max 369/4679 ms before versus 329/566 ms with
the candidate. The installed candidate then completed 16 requests over roughly
five minutes with median 288 ms and maximum 573 ms, without crashes or automatic
restarts. These small samples do not prove that route jitter or loss is resolved.

Publication requires complete validation of the matching runtime source, signed
checksums, and verification of the published archive through the updater.
