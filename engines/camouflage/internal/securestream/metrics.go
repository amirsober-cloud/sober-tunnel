package securestream

import (
	"encoding/json"
	"os"
	"sync/atomic"
	"time"
)

var sentBytes, receivedBytes atomic.Uint64

// StartMetrics publishes per-process plaintext tunnel counters, never host traffic.
func StartMetrics(path string) {
	started := time.Now().UnixNano()
	publish := func() {
		data, err := json.Marshal(map[string]any{"sent": sentBytes.Load(), "received": receivedBytes.Load(), "started": started, "timestamp": time.Now().UnixMilli()})
		if err != nil {
			return
		}
		tmp := path + ".tmp"
		if os.WriteFile(tmp, data, 0600) == nil {
			_ = os.Rename(tmp, path)
		}
	}
	publish()
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			publish()
		}
	}()
}
