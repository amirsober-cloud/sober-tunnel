package health

import (
	"sync"
	"time"
)

type Checker struct {
	mu        sync.RWMutex
	lastSeen  time.Time
	connected bool
}

func New() *Checker {
	return &Checker{
		lastSeen:  time.Now(),
		connected: false,
	}
}

func (c *Checker) SetConnected(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = v
	c.lastSeen = time.Now()
}

func (c *Checker) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected
}

func (c *Checker) LastSeen() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lastSeen
}
