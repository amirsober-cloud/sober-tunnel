package mux

import (
	"net"
	"sync"
)

type Stream struct {
	ID uint32

	Session *Session

	Conn net.Conn

	mu sync.Mutex

	closed bool
}

func (s *Stream) Close() {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true

	if s.Conn != nil {
		s.Conn.Close()
	}
}
