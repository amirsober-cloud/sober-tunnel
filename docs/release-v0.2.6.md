# v0.2.6 - Start newly created tunnels automatically

New Camouflage instances now run systemctl enable --now after successful validation and save, from both the terminal menu and panel. They also start after reboot. Editing an existing stopped instance preserves its stopped state. Startup errors explicitly report that settings were saved and remain available for recovery through Start or Logs.

Existing running tunnels and x-ui settings are unchanged. Existing stopped instances are not automatically started by this update; use Start once for those instances.

Regression checks cover one-time startup, preserving stopped edits and retaining configuration after startup failure. Signed Linux amd64 bundle; both executables report v0.2.6.
