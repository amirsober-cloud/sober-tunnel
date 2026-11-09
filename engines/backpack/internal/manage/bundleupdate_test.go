package manage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundleArchiveRequiresBothExecutablesBeforeReplacingEither(t *testing.T) {
	for _, complete := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing helper", true: "complete"}[complete], func(t *testing.T) {
			dir := t.TempDir()
			archive := filepath.Join(dir, "release.tar.gz")
			f, err := os.Create(archive)
			if err != nil {
				t.Fatal(err)
			}
			gz := gzip.NewWriter(f)
			tw := tar.NewWriter(gz)
			members := map[string]string{"sober-tunnel": "#!/bin/sh\nprintf 'v9.8.7\\n'\n"}
			if complete {
				members["sober-tunnel-mimic"] = "#!/bin/sh\nprintf 'sober-tunnel-mimic v9.8.7 (test)\\n'\n"
			}
			for name, data := range members {
				if err = tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(data))}); err != nil {
					t.Fatal(err)
				}
				if _, err = tw.Write([]byte(data)); err != nil {
					t.Fatal(err)
				}
			}
			tw.Close()
			gz.Close()
			f.Close()
			main := filepath.Join(dir, "installed-main")
			helper := filepath.Join(dir, "installed-helper")
			os.WriteFile(main, []byte("old main"), 0755)
			os.WriteFile(helper, []byte("old helper"), 0755)
			err = installBundleArchive(archive, main, helper)
			if (err == nil) != complete {
				t.Fatalf("unexpected archive result: %v", err)
			}
			for path, name := range map[string]string{main: "main", helper: "helper"} {
				b, _ := os.ReadFile(path)
				want := "old " + name
				if complete {
					key := "sober-tunnel"
					if name == "helper" {
						key += "-mimic"
					}
					want = members[key]
				}
				if string(b) != want {
					t.Fatalf("inconsistent installed pair: %s", b)
				}
			}
		})
	}
}

func TestBundleVersionMismatchDoesNotReplaceInstalledPair(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "release.tar.gz")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, version := range map[string]string{"sober-tunnel": "v9.8.7", "sober-tunnel-mimic": "sober-tunnel-mimic v9.8.6"} {
		data := fmt.Sprintf("#!/bin/sh\nprintf '%s\\n'\n", version)
		if err = tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		tw.Write([]byte(data))
	}
	tw.Close()
	gz.Close()
	f.Close()
	main := filepath.Join(dir, "main")
	helper := filepath.Join(dir, "helper")
	os.WriteFile(main, []byte("old main"), 0755)
	os.WriteFile(helper, []byte("old helper"), 0755)
	if err = installBundleArchive(archive, main, helper); err == nil || !strings.Contains(err.Error(), "same version") {
		t.Fatalf("did not reject mismatched versions: %v", err)
	}
	for path, want := range map[string]string{main: "old main", helper: "old helper"} {
		b, _ := os.ReadFile(path)
		if string(b) != want {
			t.Fatal("modified the installed pair before checking versions")
		}
	}
}
