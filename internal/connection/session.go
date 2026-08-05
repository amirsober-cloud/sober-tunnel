package connection

import (
	"net"
	"sync"
)

type Session struct {
	Conn net.Conn

	Streams int

	MaxStreams int

	mu sync.Mutex
}

func NewSession(conn net.Conn, maxStreams int) *Session {
	return &Session{
		Conn:       conn,
		MaxStreams: maxStreams,
	}
}

func (s *Session) CanOpenStream() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Streams < s.MaxStreams
}

func (s *Session) AcquireStream() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Streams >= s.MaxStreams {
		return false
	}

	s.Streams++
	return true
}

func (s *Session) ReleaseStream() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Streams > 0 {
		s.Streams--
	}
}

func (s *Session) ActiveStreams() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Streams
}
