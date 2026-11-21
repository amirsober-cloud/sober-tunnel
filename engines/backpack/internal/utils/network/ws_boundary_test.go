package network

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRawMuxBytesRemainAfterCoalescedHTTPUpgrade(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	wrapped := newUpgradeBoundaryConn(a)
	wrapped.SetReadDeadline(time.Now().Add(time.Second))
	go func() {
		_, _ = io.WriteString(b, "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\nfirst-smux-frame")
	}()
	// This reader represents Gorilla's private HTTP response reader, which
	// NetConn bypasses. The bytes following the upgrade must stay in wrapped.
	private := bufio.NewReader(wrapped)
	if _, err := http.ReadResponse(private, &http.Request{Method: "GET"}); err != nil {
		t.Fatal(err)
	}
	if private.Buffered() != 0 {
		t.Fatal("HTTP parser stole raw mux bytes")
	}
	got := make([]byte, len("first-smux-frame"))
	if _, err := io.ReadFull(wrapped, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "first-smux-frame" {
		t.Fatalf("first frame corrupted: %q", got)
	}
}
