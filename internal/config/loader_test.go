package config

import "testing"

func TestLoadExampleConfig(t *testing.T) {

	cfg, err := Load("config.example.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Role != "client" {
		t.Fatalf("expected role client, got %s", cfg.Role)
	}

	if cfg.Transport != "tcp" {
		t.Fatalf("expected transport tcp, got %s", cfg.Transport)
	}
}
