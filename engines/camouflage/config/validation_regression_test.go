package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeSaveReplacesPrivatelyAndPreservesZeroJitter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instance.json")
	t.Setenv("SOBER_MIMIC_CONFIG", path)
	if err := os.WriteFile(path, []byte(`{"role":"old"}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := &AppConfig{Role: "iran", ForeignNodes: []ForeignNode{{Alias: "a", BandwidthJitterMbps: 0}}}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("config credentials are not private")
	}
	b, _ := os.ReadFile(path)
	again, err := parseConfig(b)
	if err != nil || again.ForeignNodes[0].BandwidthJitterMbps != 0 {
		t.Fatal("zero jitter lost on native save")
	}
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".sober-config-*"))
	if len(files) > 0 {
		t.Fatal("temporary config not cleaned")
	}
}

func TestExplicitZeroJitterSurvivesSerialization(t *testing.T) {
	for _, value := range []string{``, `,"bandwidth_jitter_mbps":0`, `,"bandwidth_jitter_mbps":4`} {
		cfg, err := parseConfig([]byte(`{"role":"iran","foreign_nodes":[{"alias":"a"` + value + `}]}`))
		if err != nil {
			t.Fatal(err)
		}
		want := 2
		if value != "" {
			want = 0
		}
		if value == `,"bandwidth_jitter_mbps":4` {
			want = 4
		}
		if cfg.ForeignNodes[0].BandwidthJitterMbps != want {
			t.Fatalf("wrong jitter for %s", value)
		}
		b, _ := json.Marshal(cfg)
		again, err := parseConfig(b)
		if err != nil || again.ForeignNodes[0].BandwidthJitterMbps != want {
			t.Fatal("jitter changed after save/reload")
		}
	}
}
func TestRejectListenerAndRoutingConflicts(t *testing.T) {
	base := `{"role":"iran","foreign_nodes":[{"alias":"a","auth_token":"0123456789abcdef","local_socks_port":1081,"target_ip":"203.0.113.1"}]}`
	var cases []AppConfig
	cfg, _ := parseConfig([]byte(base))
	n := cfg.ForeignNodes[0]
	other := n
	other.Alias = "b"
	other.SocksBind = "0.0.0.0"
	cfg.ForeignNodes = append(cfg.ForeignNodes, other)
	cases = append(cases, *cfg)
	cfg, _ = parseConfig([]byte(base))
	cfg.ForeignNodes[0].DNSEnabled = true
	cases = append(cases, *cfg)
	cfg, _ = parseConfig([]byte(base))
	cfg.ForeignNodes[0].TunEnabled = true
	cfg.ForeignNodes[0].TunName = "../invalid"
	cfg.ForeignNodes[0].TunAddr = "10.200.0.1/24"
	cases = append(cases, *cfg)
	cfg, _ = parseConfig([]byte(base))
	cfg.ForeignNodes[0].SocksBind = "bad host"
	cases = append(cases, *cfg)
	cfg, _ = parseConfig([]byte(base))
	cfg.ForeignNodes[0].Endpoints = nil
	cases = append(cases, *cfg)
	for i, c := range cases {
		if c.Validate() == nil {
			t.Fatalf("accepted invalid case %d", i)
		}
	}
}
