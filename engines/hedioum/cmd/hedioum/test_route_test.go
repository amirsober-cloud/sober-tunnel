package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTunDiagnosticNeverFlushesExistingRoutingTables(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	script := "#!/bin/sh\necho \"$*\" >> \"$ROUTE_TEST_LOG\"\nexit 0\n"
	if err := os.WriteFile(filepath.Join(dir, "ip"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("ROUTE_TEST_LOG", log)
	cleanup, ok := ensureTunTestRoute("10.200.1.1", "testtun0")
	if !ok {
		t.Fatal("could not create test route")
	}
	cleanup()
	cleanup()
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "flush") || strings.Contains(s, "5199") {
		t.Fatalf("modified an unrelated routing table: %s", s)
	}
	if strings.Count(s, "rule del") != 1 || strings.Count(s, "route del") != 1 {
		t.Fatalf("cleanup must remove only its own entries once: %s", s)
	}
}
