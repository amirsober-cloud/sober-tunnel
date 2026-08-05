package forward

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/protocol"
)

type MuxForwarder struct {
	conn net.Conn

	mu sync.Mutex

	nextStream uint32

	streams map[uint32]net.Conn

	lastActive map[uint32]time.Time

	maxStreams    int
	cleanInterval time.Duration
}

func NewMuxForwarder(
	conn net.Conn,
) *MuxForwarder {

	return &MuxForwarder{
		conn:          conn,
		nextStream:    1,
		streams:       make(map[uint32]net.Conn),
		lastActive:    make(map[uint32]time.Time),
		maxStreams:    1024,
		cleanInterval: time.Minute,
	}
}

func (m *MuxForwarder) ReadLoop() {

	for {

		frame, err := protocol.ReadFrame(
			m.conn,
		)

		if err != nil {
			return
		}

		if frame.Command != protocol.CmdData {
			continue
		}

		m.mu.Lock()

		conn := m.streams[frame.StreamID]

		m.mu.Unlock()

		if conn == nil {
			continue
		}

		_, err = conn.Write(
			frame.Payload,
		)

		if err != nil {

			conn.Close()

			m.mu.Lock()

			delete(
				m.streams,
				frame.StreamID,
			)

			m.mu.Unlock()
		}
	}
}

func (m *MuxForwarder) nextID() uint32 {

	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.nextStream

	m.nextStream++

	return id
}

func (m *MuxForwarder) HandleLocal(
	conn net.Conn,
	target string,
) {

	host, port, err := splitTarget(target)

	if err != nil {
		conn.Close()
		return
	}
	m.mu.Lock()

	if len(m.streams) >= m.maxStreams {
		m.mu.Unlock()

		conn.Close()

		return
	}

	m.mu.Unlock()
	streamID := m.nextID()

	m.mu.Lock()

	m.streams[streamID] = conn
	m.lastActive[streamID] = time.Now()

	m.mu.Unlock()

	open := &protocol.Frame{

		Version: protocol.Version1,

		Command: protocol.CmdOpen,

		StreamID: streamID,

		Address: host,

		Port: uint16(port),
	}

	err = protocol.WriteFrame(
		m.conn,
		open,
	)

	if err != nil {

		m.mu.Lock()

		delete(
			m.streams,
			streamID,
		)

		m.mu.Unlock()

		conn.Close()

		return
	}

	fmt.Println("OPEN WRITE OK")

	go m.pipeLocalToMux(
		conn,
		streamID,
	)
}

func (m *MuxForwarder) pipeLocalToMux(
	conn net.Conn,
	streamID uint32,
) {

	buf := make(
		[]byte,
		32*1024,
	)

	for {

		n, err := conn.Read(buf)
		m.mu.Lock()
		m.lastActive[streamID] = time.Now()
		m.mu.Unlock()

		if err != nil {

			m.mu.Lock()

			delete(
				m.streams,
				streamID,
			)

			m.mu.Unlock()

			return
		}

		frame := &protocol.Frame{

			Version: protocol.Version1,

			Command: protocol.CmdData,

			StreamID: streamID,

			Payload: append(
				[]byte(nil),
				buf[:n]...,
			),
		}

		err = protocol.WriteFrame(
			m.conn,
			frame,
		)

		if err != nil {

			return
		}
	}
}
func (m *MuxForwarder) StartCleaner(timeout time.Duration) {

	ticker := time.NewTicker(m.cleanInterval)

	go func() {

		for range ticker.C {

			now := time.Now()

			m.mu.Lock()

			for id, last := range m.lastActive {

				if now.Sub(last) > timeout {

					if conn := m.streams[id]; conn != nil {
						conn.Close()
					}

					delete(
						m.streams,
						id,
					)

					delete(
						m.lastActive,
						id,
					)
				}
			}

			m.mu.Unlock()
		}

	}()

}
func splitTarget(target string) (string, int, error) {

	host, port, err := net.SplitHostPort(
		target,
	)

	if err != nil {
		return "", 0, err
	}

	var p int

	_, err = fmt.Sscanf(
		port,
		"%d",
		&p,
	)

	if err != nil {
		return "", 0, err
	}

	return host, p, nil
}
