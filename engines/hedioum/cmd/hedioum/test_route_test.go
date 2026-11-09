package main

import (
	"os"
	"path/filepath"
	"strconv"
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
	if strings.Contains(s, "flush") || strings.Contains(s, "table 5199\n") || strings.Contains(s, "priority 5199\n") {
		t.Fatalf("modified an unrelated routing table: %s", s)
	}
	if strings.Count(s, "rule del") != 1 || strings.Count(s, "route del") != 1 {
		t.Fatalf("cleanup must remove only its own entries once: %s", s)
	}
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "rule add ") {
			parts := strings.Fields(line)
			priority, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil || priority >= 32766 || priority < 1 {
				t.Fatalf("test route must precede the Linux main table: %s", line)
			}
		}
	}
}
