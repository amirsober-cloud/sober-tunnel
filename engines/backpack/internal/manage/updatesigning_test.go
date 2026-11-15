package manage

import (
	"errors"
	"strings"
	"testing"
)

func TestUnsignedUpdateHasActionableBootstrapError(t *testing.T) {
	restore := withReleaseKey(t, "")
	defer restore()
	err := ApplyUpdate(nil)
	if !errors.Is(err, ErrUpdateSigningUnavailable) {
		t.Fatalf("expected signing bootstrap error, got %v", err)
	}
	if !strings.Contains(err.Error(), "https://raw.githubusercontent.com/amirsober-cloud/sober-tunnel/main/install.sh") {
		t.Fatal("unsigned build gives no recovery command")
	}
}
