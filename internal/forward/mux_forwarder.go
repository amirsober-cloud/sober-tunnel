package forward

import (
	"github.com/amirsober-cloud/sober-tunnel/internal/protocol"
	"net"
	"strconv"
	"sync"
	"time"
)

type MuxForwarder struct {
	conn          net.Conn
	mu            sync.Mutex
	writeMu       sync.Mutex
	nextStream    uint32
	streams       map[uint32]net.Conn
	lastActive    map[uint32]time.Time
	maxStreams    int
	cleanInterval time.Duration
	done          chan struct{}
	closeOnce     sync.Once
}

func NewMuxForwarder(conn net.Conn) *MuxForwarder {
	return &MuxForwarder{conn: conn, nextStream: 1, streams: make(map[uint32]net.Conn), lastActive: make(map[uint32]time.Time), maxStreams: 1024, cleanInterval: time.Minute, done: make(chan struct{})}
}
func (m *MuxForwarder) writeFrame(frame *protocol.Frame) error {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	return protocol.WriteFrame(m.conn, frame)
}
func (m *MuxForwarder) Close() {
	m.closeOnce.Do(func() {
		close(m.done)
		m.conn.Close()
		m.mu.Lock()
		defer m.mu.Unlock()
		for id, conn := range m.streams {
			conn.Close()
			delete(m.streams, id)
			delete(m.lastActive, id)
		}
	})
}
func (m *MuxForwarder) removeStream(id uint32) {
	m.mu.Lock()
	conn := m.streams[id]
	delete(m.streams, id)
	delete(m.lastActive, id)
	m.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
}
func (m *MuxForwarder) ReadLoop() {
	defer m.Close()
	for {
		frame, err := protocol.ReadFrame(m.conn)
		if err != nil {
			return
		}
		if frame.Command == protocol.CmdClose {
			m.removeStream(frame.StreamID)
			continue
		}
		if frame.Command != protocol.CmdData {
			continue
		}
		m.mu.Lock()
		conn := m.streams[frame.StreamID]
		if conn != nil {
			m.lastActive[frame.StreamID] = time.Now()
		}
		m.mu.Unlock()
		if conn != nil {
			if _, err := conn.Write(frame.Payload); err != nil {
				m.removeStream(frame.StreamID)
			}
		}
	}
}
func (m *MuxForwarder) HandleLocal(conn net.Conn, target string) {
	defer conn.Close()
	host, port, err := splitTarget(target)
	if err != nil {
		return
	}
	m.mu.Lock()
	select {
	case <-m.done:
		m.mu.Unlock()
		return
	default:
	}
	if len(m.streams) >= m.maxStreams {
		m.mu.Unlock()
		return
	}
	id := m.nextStream
	for id == 0 || m.streams[id] != nil {
		id++
	}
	m.nextStream = id + 1
	m.streams[id] = conn
	m.lastActive[id] = time.Now()
	m.mu.Unlock()
	defer m.removeStream(id)
	if err := m.writeFrame(&protocol.Frame{Version: protocol.Version1, Command: protocol.CmdOpen, StreamID: id, Address: host, Port: uint16(port)}); err != nil {
		m.Close()
		return
	}
	defer m.writeFrame(&protocol.Frame{Version: protocol.Version1, Command: protocol.CmdClose, StreamID: id})
	m.pipeLocalToMux(conn, id)
}
func (m *MuxForwarder) pipeLocalToMux(conn net.Conn, id uint32) {
	buf := make([]byte, 32*1024)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			m.mu.Lock()
			if m.streams[id] != nil {
				m.lastActive[id] = time.Now()
			}
			m.mu.Unlock()
			if writeErr := m.writeFrame(&protocol.Frame{Version: protocol.Version1, Command: protocol.CmdData, StreamID: id, Payload: buf[:n]}); writeErr != nil {
				m.Close()
				return
			}
		}
		if err != nil {
			return
		}
	}
}
func (m *MuxForwarder) StartCleaner(timeout time.Duration) {
	ticker := time.NewTicker(m.cleanInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-m.done:
				return
			case now := <-ticker.C:
				m.mu.Lock()
				for id, last := range m.lastActive {
					if now.Sub(last) > timeout {
						if conn := m.streams[id]; conn != nil {
							conn.Close()
						}
						delete(m.streams, id)
						delete(m.lastActive, id)
					}
				}
				m.mu.Unlock()
			}
		}
	}()
}
func splitTarget(target string) (string, int, error) {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		return "", 0, err
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	if host == "" || p < 1 || p > 65535 {
		return "", 0, &net.AddrError{Err: "invalid target", Addr: target}
	}
	return host, p, nil
}
