package connection

import (
	"net"
	"sync"

	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

type Pool struct {
	mu    sync.Mutex
	conns []net.Conn
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
