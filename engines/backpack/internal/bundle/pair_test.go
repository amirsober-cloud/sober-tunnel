package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallPairAndRollbackOnSecondRenameFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			dir := t.TempDir()
			main := filepath.Join(dir, "main")
			helper := filepath.Join(dir, "helper")
			nextMain := filepath.Join(dir, "source-main")
			nextHelper := filepath.Join(dir, "source-helper")
			for path, data := range map[string]string{main: "old main", helper: "old helper", nextMain: "new main", nextHelper: "new helper"} {
				if err := os.WriteFile(path, []byte(data), 0755); err != nil {
					t.Fatal(err)
				}
			}
			rename := func(a, b string) error {
				if fail && b == main && strings.HasSuffix(a, "next-main") {
					return fmt.Errorf("injected disk error")
				}
				return os.Rename(a, b)
			}
			err := installPair(nextMain, nextHelper, main, helper, rename)
			if (err != nil) != fail {
				t.Fatalf("unexpected result %v", err)
			}
			want := "new "
			if fail {
				want = "old "
			}
			for path, suffix := range map[string]string{main: "main", helper: "helper"} {
				b, _ := os.ReadFile(path)
				if string(b) != want+suffix {
					t.Fatalf("mismatched bundle: %s", b)
				}
			}
			matches, _ := filepath.Glob(filepath.Join(dir, ".sober-bundle-*"))
			if len(matches) > 0 {
				t.Fatal("staging files left behind")
			}
		})
	}
}
func TestMissingHelperNeverChangesMain(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main")
	os.WriteFile(main, []byte("old"), 0755)
	if InstallPair(main, filepath.Join(dir, "missing"), main, filepath.Join(dir, "helper")) == nil {
		t.Fatal("accepted missing helper")
	}
	b, _ := os.ReadFile(main)
	if string(b) != "old" {
		t.Fatal("modified main before complete bundle was staged")
	}
}
