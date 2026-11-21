package egress

import (
	"crypto/sha256"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/resumestream"
	"sync"
)

var resumePeers = struct {
	sync.Mutex
	m map[[32]byte]*resumestream.Registry
}{m: map[[32]byte]*resumestream.Registry{}}

func resumeRegistry(token string) *resumestream.Registry {
	key := sha256.Sum256([]byte(token))
	resumePeers.Lock()
	defer resumePeers.Unlock()
	if registry := resumePeers.m[key]; registry != nil {
		return registry
	}
	registry := resumestream.NewRegistry()
	resumePeers.m[key] = registry
	return registry
}
