//go:build !linux

package mimicmgr

import "os/exec"

func detachDiagnostic(cmd *exec.Cmd) {}
