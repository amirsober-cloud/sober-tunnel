package mimicmgr

import (
	"os"
	"os/exec"
	"testing"
)

func TestJobLaunchLockAcrossProcesses(t *testing.T) {
	if dir := os.Getenv("SOBER_TEST_JOB_LOCK"); dir != "" {
		unlock, err := lockJobLaunch(dir)
		if err == nil {
			unlock()
			os.Exit(2)
		}
		os.Exit(0)
	}
	dir := t.TempDir()
	unlock, err := lockJobLaunch(dir)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestJobLaunchLockAcrossProcesses$")
	cmd.Env = append(os.Environ(), "SOBER_TEST_JOB_LOCK="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		unlock()
		t.Fatalf("another process bypassed the launch lock: %v %s", err, out)
	}
	unlock()
	again, err := lockJobLaunch(dir)
	if err != nil {
		t.Fatal("lock not released", err)
	}
	again()
}
