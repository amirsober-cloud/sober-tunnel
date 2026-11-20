package mimicmgr

import (
	"context"
	"os"
	"testing"
)

func TestTrafficIsIsolatedAndDoesNotBecomeAnInstance(t *testing.T) {
	m := testManager(t)
	if err := m.Save(context.Background(), "exit", []byte(`{"role":"foreign"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Traffic("../exit"); err == nil {
		t.Fatal("unsafe name accepted")
	}
	if _, err := m.Traffic("missing"); err == nil {
		t.Fatal("missing instance accepted")
	}
	expected := `{"sent":123,"received":456,"timestamp":1000,"started":1}`
	if err := os.WriteFile(m.configPath("exit")+".traffic.json", []byte(expected), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := m.Traffic("exit")
	if err != nil || string(data) != expected {
		t.Fatalf("%s %v", data, err)
	}
	items, err := m.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("%v %v", items, err)
	}
}

func TestTrafficSummaryIncludesStoppedCoreAndDailyUsage(t *testing.T) {
	m := testManager(t)
	if err := m.Save(context.Background(), "exit", []byte(`{"role":"foreign"}`)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m.configPath("exit")+".traffic.json", []byte(`{"sent":123,"received":456,"timestamp":1,"days":{"2026-11-21":579}}`), 0600); err != nil {
		t.Fatal(err)
	}
	s := m.TrafficSummary()
	if s.Sent != 123 || s.Received != 456 || s.Total != 1 || s.Running != 0 || s.Days["2026-11-21"] != 579 {
		t.Fatalf("wrong summary: %+v", s)
	}
}
