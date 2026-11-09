package mimicmgr

import (
	"os/exec"
	"syscall"
)

// Keep a remote diagnostic alive after the short-lived node RPC/SSH exits.
func detachDiagnostic(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
