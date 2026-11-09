package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMimicValidationModeNeverStartsNetworking(t *testing.T) {
	file := filepath.Join(t.TempDir(), "exit.json")
	if err := os.WriteFile(file, []byte(`{"role":"foreign","auth_token":"0123456789abcdef","mimics":[{"type":"tls","port":8443}],"http_decoy_port":-1}`), 0600); err != nil {
		t.Fatal(err)
	}
	handled, err := mimicCommand([]string{"--check-mimic", "--mimic-config", file})
	if !handled || err != nil {
		t.Fatalf("check mode: handled=%v err=%v", handled, err)
	}
	if handled, err := mimicCommand([]string{"--check-mimic"}); !handled || err == nil {
		t.Fatal("check accepted missing config")
	}
	if handled, _ := mimicCommand([]string{"--webui"}); handled {
		t.Fatal("intercepted panel mode")
	}
	if handled, _ := mimicCommand([]string{"-c", "tunnel.toml"}); handled {
		t.Fatal("intercepted existing tunnel mode")
	}
}
