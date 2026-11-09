package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsolatedConfigAndValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instance.json")
	data := []byte(`{"role":"foreign","auth_token":"0123456789abcdef","mimics":[{"type":"tls","port":8443}],"http_decoy_port":-1}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DecoyPort != 2022 || cfg.EgressIPMode != "ipv4" {
		t.Fatal("upstream defaults lost")
	}
	t.Setenv("SOBER_MIMIC_CONFIG", path)
	if getConfigPath() != path {
		t.Fatal("instance ignored isolated path")
	}
	if _, err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
}

func TestRejectInvalidMimicConfiguration(t *testing.T) {
	for _, data := range []string{
		`{"role":"invalid"}`,
		`{"role":"foreign","auth_token":"short"}`,
		`{"role":"foreign","auth_token":"0123456789abcdef","mimics":[{"type":"typo","port":443}]}`,
		`{"role":"foreign","auth_token":"0123456789abcdef","mimics":[{"type":"tls","port":70000}]}`,
		`{"role":"foreign","auth_token":"0123456789abcdef","mimics":[{"type":"tls","port":443},{"type":"ssh","port":443}]}`,
		`{"role":"iran","foreign_nodes":[]}`,
		`{"role":"iran","foreign_nodes":[{"alias":"exit","auth_token":"0123456789abcdef","local_socks_port":1081,"endpoints":[{"target":"missing-port","mimic":"tls"}]}]}`,
	} {
		cfg, err := parseConfig([]byte(data))
		if err == nil {
			err = cfg.Validate()
		}
		if err == nil {
			t.Errorf("accepted invalid config: %s", data)
		}
	}
}
