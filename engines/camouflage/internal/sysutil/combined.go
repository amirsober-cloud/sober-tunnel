package sysutil

import "os"

// CombinedInstalled prevents standalone maintenance from leaving half of the
// unified installation behind. Panel tools use the paired installer instead.
func CombinedInstalled() bool {
	info, err := os.Stat("/usr/local/bin/sober-tunnel")
	return err == nil && !info.IsDir()
}
