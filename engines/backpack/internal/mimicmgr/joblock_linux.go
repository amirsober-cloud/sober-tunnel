package mimicmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// RPC processes and the panel share this kernel lock; a crashed caller cannot
// leave a stale lock that blocks all future diagnostics.
func lockJobLaunch(dir string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(dir, ".launch.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("a diagnostic is being launched on this server; retry shortly")
	}
	return func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN); _ = f.Close() }, nil
}
