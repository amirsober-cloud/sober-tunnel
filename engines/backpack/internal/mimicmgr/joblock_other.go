//go:build !linux

package mimicmgr

// Production service and diagnostic management requires Linux.
func lockJobLaunch(string) (func(), error) { return func() {}, nil }
