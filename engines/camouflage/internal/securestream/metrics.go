package securestream

import (
	"encoding/json"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var sentBytes, receivedBytes atomic.Uint64

type trafficSnapshot struct {
	Sent      uint64            `json:"sent"`
	Received  uint64            `json:"received"`
	Started   int64             `json:"started"`
	Timestamp int64             `json:"timestamp"`
	Days      map[string]uint64 `json:"days"`
}

// StartMetrics preserves totals across restarts; daily figures begin at upgrade.
func StartMetrics(path string) {
	var old trafficSnapshot
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &old)
	}
	if old.Days == nil {
		old.Days = map[string]uint64{}
	}
	baseSent, baseRecv := old.Sent, old.Received
	previous := uint64(0)
	publish := func() {
		sent, recv := sentBytes.Load(), receivedBytes.Load()
		total := sent + recv
		day := time.Now().UTC().Format("2006-01-02")
		old.Days[day] += total - previous
		previous = total
		cutoff := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
		for key := range old.Days {
			if key < cutoff {
				delete(old.Days, key)
			}
		}
		old.Sent = baseSent + sent
		old.Received = baseRecv + recv
		old.Timestamp = time.Now().UnixMilli()
		data, err := json.Marshal(old)
		if err != nil {
			return
		}
		tmp := path + ".tmp"
		if os.WriteFile(tmp, data, 0600) == nil {
			_ = os.Rename(tmp, path)
		}
	}
	old.Started = time.Now().UnixNano()
	publish()
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			publish()
		}
	}()
}

// MeterAuthenticatedConn accounts for plaintext on mimics whose TLS layer
// supplies encryption instead of SecureConn. Apply only after authentication;
// unauthenticated probes and decoy replies must not count as tunnel traffic.
func MeterAuthenticatedConn(conn net.Conn) net.Conn {
	return &meteredConn{Conn: conn}
}

type meteredConn struct{ net.Conn }

func (c *meteredConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	receivedBytes.Add(uint64(n))
	return n, err
}
func (c *meteredConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	sentBytes.Add(uint64(n))
	return n, err
}
