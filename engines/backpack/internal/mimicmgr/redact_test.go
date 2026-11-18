package mimicmgr

import (
	"strings"
	"testing"
)

func TestPreviewRedactsNestedCredentials(t *testing.T) {
	in := []byte(`{"auth_token":"foreign-secret","nodes":[{"password":"ssh-secret","target":"exit:22"}]}`)
	out := string(RedactConfig(in))
	if strings.Contains(out, "foreign-secret") || strings.Contains(out, "ssh-secret") || !strings.Contains(out, "exit:22") {
		t.Fatal("credential redaction failed")
	}
	if !strings.Contains(string(in), "foreign-secret") {
		t.Fatal("original modified")
	}
}
