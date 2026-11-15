package mux

import (
	"net"
	"sync"

	"github.com/amirsober-cloud/sober-tunnel/internal/protocol"
)

type Session struct {
	conn net.Conn

	mu      sync.RWMutex
	writeMu sync.Mutex

	streams map[uint32]*Stream

	nextID uint32

	closed bool
}

func NewSession(
	conn net.Conn,
) *Session {

	return &Session{
		conn: conn,

		streams: make(map[uint32]*Stream),

		nextID: 1,
	}
}

func (s *Session) NextStreamID() uint32 {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nextID == 0 {
		s.nextID = 1
	}

	id := s.nextID

	s.nextID++

	return id
}

func (s *Session) AddStream(
	st *Stream,
) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[st.ID]; exists || s.closed {
		return false
	}

	s.streams[st.ID] = st

	return true
}

func (s *Session) WriteFrame(frame *protocol.Frame) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return protocol.WriteFrame(s.conn, frame)
}

func (s *Session) RemoveStream(
	id uint32,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(
		s.streams,
		id,
	)
}

func (s *Session) GetStream(
	id uint32,
) *Stream {

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.streams[id]
}

func (s *Session) Close() {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true

	for _, st := range s.streams {

		st.Close()
	}

	s.streams = make(map[uint32]*Stream)

	if s.conn != nil {
		s.conn.Close()
	}
}
