package manage

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/backpack/backpack/internal/app"
)

// Opt-in check against published artifacts; never installs or restarts services.
func TestPublishedReleaseSignatureAndArchive(t *testing.T) {
	tag := os.Getenv("RELEASE_VERIFICATION_TAG")
	if tag == "" {
		t.Skip("set RELEASE_VERIFICATION_TAG for the published-release check")
	}
	key := os.Getenv("RELEASE_PUBLIC_KEY")
	if key == "" {
		t.Fatal("published-release verification requires the pinned public key")
	}
	restore := withReleaseKey(t, key)
	defer restore()
	asset := app.AssetName()
	sums, want, err := fetchChecksums(tag, asset)
	if err != nil || want == "" {
		t.Fatalf("published checksums: %q %v", want, err)
	}
	if err := verifyChecksumSignature(tag, sums); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Get(repoURL() + "/releases/download/" + tag + "/" + asset)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("archive download: %s", resp.Status)
	}
	path := filepath.Join(t.TempDir(), asset)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_, copyErr := io.Copy(f, io.LimitReader(resp.Body, 256<<20))
	closeErr := f.Close()
	if copyErr != nil {
		t.Fatal(copyErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if err := verifyChecksum(path, want); err != nil {
		t.Fatal(err)
	}
}
