package connection

import (
	"net"
	"sync"

	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

type Pool struct {
	mu sync.Mutex

	conns []net.Conn

	sessions []*Session
}

func (p *Pool) AddSession(s *Session) {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.sessions = append(p.sessions, s)
}

func (p *Pool) Sessions() []*Session {

	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]*Session, len(p.sessions))
	copy(out, p.sessions)

	return out
}

func (p *Pool) FreeSession() *Session {

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, s := range p.sessions {
		if s.CanOpenStream() {
			return s
		}
	}

	return nil
}
func NewPool() *Pool {
	return &Pool{
		conns: make([]net.Conn, 0),
	}
}

func (p *Pool) Add(conn net.Conn) {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.conns = append(
		p.conns,
		conn,
	)

	logger.Info.Printf(
		"Connection added. Total: %d",
		len(p.conns),
	)
}

func (p *Pool) Size() int {

	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.conns)
}

func (p *Pool) Remove(conn net.Conn) {

	p.mu.Lock()
	defer p.mu.Unlock()

	for i, c := range p.conns {

		if c == conn {

			conn.Close()

			p.conns = append(
				p.conns[:i],
				p.conns[i+1:]...,
			)

			logger.Info.Printf(
				"Connection removed. Total: %d",
				len(p.conns),
			)

			return
		}
	}
}
func (p *Pool) RemoveLast() {

	p.mu.Lock()
	defer p.mu.Unlock()

	size := len(p.conns)

	if size == 0 {
		return
	}

	conn := p.conns[size-1]

	conn.Close()

	p.conns = p.conns[:size-1]

	logger.Info.Printf(
		"Connection removed. Total: %d",
		len(p.conns),
	)
}
func (p *Pool) Connections() []net.Conn {

	p.mu.Lock()
	defer p.mu.Unlock()

	result := make(
		[]net.Conn,
		len(p.conns),
	)

	copy(
		result,
		p.conns,
	)

	return result
}

func (p *Pool) Close() {

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.conns {
		conn.Close()
	}

	p.conns = nil

	logger.Info.Println(
		"Connection pool closed",
	)
}
