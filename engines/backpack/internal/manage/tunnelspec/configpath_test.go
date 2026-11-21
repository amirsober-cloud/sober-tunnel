package tunnelspec

import "testing"

func useSpecRoot(t *testing.T) string {
	t.Helper()
	old := configRoot
	configRoot = t.TempDir()
	t.Cleanup(func() { configRoot = old })
	return configRoot
}
