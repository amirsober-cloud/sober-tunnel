// Package resumestream preserves ordered TCP bytes across authenticated mux
// streams. It does not survive process restart and never replays a backend dial.
package resumestream

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

const Type byte = 7
const window = 128 << 10
const frameMax = 32 << 10
const grace = 60 * time.Second
const (
	dataFrame  byte = 1
	ackFrame   byte = 2
	finFrame   byte = 3
	resetFrame byte = 4
)

type OpenFunc func() (net.Conn, error)
type Conn struct {
	finDelivered                bool
	mu                          sync.Mutex
	writeMu                     sync.Mutex
	changed                     chan struct{}
	done                        chan struct{}
	closed                      bool
	tx, rx                      []byte
	base, end, received         uint64
	txFIN, rxFIN                bool
	readDeadline, writeDeadline time.Time
	wire                        net.Conn
	generation                  uint64
	id                          [16]byte
	opener                      OpenFunc
	recovering                  bool
	onClose                     func()
}

func newConn() *Conn    { return &Conn{changed: make(chan struct{}), done: make(chan struct{})} }
func (c *Conn) notify() { close(c.changed); c.changed = make(chan struct{}) }
func wait(ch <-chan struct{}, deadline time.Time) error {
	if deadline.IsZero() {
		<-ch
		return nil
	}
	delay := time.Until(deadline)
	if delay <= 0 {
		return os.ErrDeadlineExceeded
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ch:
		return nil
	case <-timer.C:
		return os.ErrDeadlineExceeded
	}
}
func (c *Conn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		c.mu.Lock()
		if len(c.rx) > 0 {
			n := copy(p, c.rx)
			c.rx = c.rx[n:]
			c.notify()
			c.mu.Unlock()
			return n, nil
		}
		if c.closed || c.rxFIN {
			c.mu.Unlock()
			return 0, io.EOF
		}
		ch, deadline := c.changed, c.readDeadline
		c.mu.Unlock()
		if err := wait(ch, deadline); err != nil {
			return 0, err
		}
	}
}
func (c *Conn) Write(p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		c.mu.Lock()
		if c.closed || c.txFIN {
			c.mu.Unlock()
			return total, net.ErrClosed
		}
		if len(c.tx) < window {
			n := len(p)
			if n > window-len(c.tx) {
				n = window - len(c.tx)
			}
			c.tx = append(c.tx, p[:n]...)
			c.end += uint64(n)
			total += n
			p = p[n:]
			c.notify()
			c.mu.Unlock()
			continue
		}
		ch, deadline := c.changed, c.writeDeadline
		c.mu.Unlock()
		if err := wait(ch, deadline); err != nil {
			return total, err
		}
	}
	return total, nil
}
func (c *Conn) CloseWrite() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return net.ErrClosed
	}
	c.txFIN = true
	c.notify()
	return nil
}
func (c *Conn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	wire := c.wire
	graceful := c.txFIN && c.rxFIN && c.finDelivered && len(c.tx) == 0
	c.wire = nil
	c.tx = nil
	c.rx = nil
	close(c.done)
	c.notify()
	callback := c.onClose
	c.mu.Unlock()
	if wire != nil {
		if !graceful && c.writeMu.TryLock() {
			_ = wire.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
			_ = writeFrame(wire, resetFrame, 0, nil)
			c.writeMu.Unlock()
		}
		wire.Close()
	}
	if callback != nil {
		callback()
	}
	return nil
}
func (c *Conn) LocalAddr() net.Addr  { return resumeAddr("local") }
func (c *Conn) RemoteAddr() net.Addr { return resumeAddr("peer") }

type resumeAddr string

func (a resumeAddr) Network() string { return "resumable-mux" }
func (a resumeAddr) String() string  { return string(a) }
func (c *Conn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDeadline = t
	c.writeDeadline = t
	c.notify()
	c.mu.Unlock()
	return nil
}
func (c *Conn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDeadline = t
	c.notify()
	c.mu.Unlock()
	return nil
}
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	c.writeDeadline = t
	c.notify()
	c.mu.Unlock()
	return nil
}
func writeAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, e := w.Write(b)
		if e != nil {
			return e
		}
		if n <= 0 || n > len(b) {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}
func writeFrame(w io.Writer, kind byte, offset uint64, payload []byte) error {
	var h [13]byte
	h[0] = kind
	binary.BigEndian.PutUint64(h[1:9], offset)
	binary.BigEndian.PutUint32(h[9:], uint32(len(payload)))
	if err := writeAll(w, h[:]); err != nil {
		return err
	}
	return writeAll(w, payload)
}
func (c *Conn) acknowledge(offset uint64) error {
	if offset < c.base || offset > c.end {
		return errors.New("resume: invalid acknowledgement")
	}
	c.tx = c.tx[int(offset-c.base):]
	c.base = offset
	c.notify()
	return nil
}
func (c *Conn) attach(wire net.Conn, ack uint64) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return net.ErrClosed
	}
	if err := c.acknowledge(ack); err != nil {
		c.mu.Unlock()
		return err
	}
	old := c.wire
	c.wire = wire
	c.finDelivered = false
	c.generation++
	c.recovering = false
	c.notify()
	c.mu.Unlock()
	if old != nil {
		old.Close()
	}
	go c.readWire(wire)
	go c.writeWire(wire, ack)
	return nil
}
func (c *Conn) detached(wire net.Conn) {
	wire.Close()
	c.mu.Lock()
	if c.closed || c.wire != wire {
		c.mu.Unlock()
		return
	}
	c.wire = nil
	c.generation++
	generation := c.generation
	c.notify()
	recoverNow := c.opener != nil && !c.recovering && !(c.txFIN && c.rxFIN)
	if recoverNow {
		c.recovering = true
	}
	c.mu.Unlock()
	if recoverNow {
		go c.reconnect()
		return
	}
	if c.opener == nil {
		go func() {
			timer := time.NewTimer(grace)
			defer timer.Stop()
			select {
			case <-c.done:
				return
			case <-timer.C:
			}
			c.mu.Lock()
			expired := !c.closed && c.wire == nil && c.generation == generation
			c.mu.Unlock()
			if expired {
				c.Close()
			}
		}()
	}
}
func (c *Conn) readWire(wire net.Conn) {
	defer c.detached(wire)
	for {
		var h [13]byte
		if _, err := io.ReadFull(wire, h[:]); err != nil {
			return
		}
		kind := h[0]
		offset := binary.BigEndian.Uint64(h[1:9])
		size := binary.BigEndian.Uint32(h[9:])
		if size > frameMax {
			c.Close()
			return
		}
		payload := make([]byte, int(size))
		if _, err := io.ReadFull(wire, payload); err != nil {
			return
		}
		c.mu.Lock()
		if c.closed || c.wire != wire {
			c.mu.Unlock()
			return
		}
		switch kind {
		case ackFrame:
			if size != 0 || c.acknowledge(offset) != nil {
				c.mu.Unlock()
				c.Close()
				return
			}
		case resetFrame:
			c.mu.Unlock()
			c.Close()
			return
		case finFrame:
			if size != 0 || offset != c.received {
				c.mu.Unlock()
				c.Close()
				return
			}
			c.rxFIN = true
			c.notify()
		case dataFrame:
			if c.rxFIN || offset > c.received {
				c.mu.Unlock()
				c.Close()
				return
			}
			skip := c.received - offset
			if skip >= uint64(len(payload)) {
				c.notify()
				c.mu.Unlock()
				continue
			}
			payload = payload[int(skip):]
			for len(c.rx)+len(payload) > window && !c.closed && c.wire == wire {
				ch := c.changed
				c.mu.Unlock()
				<-ch
				c.mu.Lock()
			}
			if c.closed || c.wire != wire {
				c.mu.Unlock()
				return
			}
			c.rx = append(c.rx, payload...)
			c.received += uint64(len(payload))
			c.notify()
		default:
			c.mu.Unlock()
			c.Close()
			return
		}
		c.mu.Unlock()
	}
}
func (c *Conn) writeWire(wire net.Conn, cursor uint64) {
	defer c.detached(wire)
	lastAck := ^uint64(0)
	finSent := false
	for {
		c.mu.Lock()
		if c.closed || c.wire != wire {
			c.mu.Unlock()
			return
		}
		if cursor < c.base {
			cursor = c.base
		}
		ack := c.received - uint64(len(c.rx))
		var payload []byte
		off := cursor
		if cursor < c.end {
			start := int(cursor - c.base)
			size := len(c.tx) - start
			if size > frameMax {
				size = frameMax
			}
			payload = append([]byte(nil), c.tx[start:start+size]...)
		}
		fin := c.txFIN && cursor == c.end && !finSent
		ch := c.changed
		c.mu.Unlock()
		if len(payload) == 0 && ack == lastAck && !fin {
			select {
			case <-ch:
				continue
			case <-c.done:
				return
			}
		}
		c.writeMu.Lock()
		_ = wire.SetWriteDeadline(time.Now().Add(8 * time.Second))
		var err error
		if ack != lastAck {
			err = writeFrame(wire, ackFrame, ack, nil)
			lastAck = ack
		}
		if err == nil && len(payload) > 0 {
			err = writeFrame(wire, dataFrame, off, payload)
			cursor += uint64(len(payload))
		}
		if err == nil && fin {
			err = writeFrame(wire, finFrame, cursor, nil)
			finSent = true
			if err == nil {
				c.mu.Lock()
				if c.wire == wire {
					c.finDelivered = true
					c.notify()
				}
				c.mu.Unlock()
			}
		}
		c.writeMu.Unlock()
		if err != nil {
			return
		}
	}
}
func (c *Conn) handshake(wire net.Conn, fresh bool) error {
	_ = wire.SetDeadline(time.Now().Add(8 * time.Second))
	defer wire.SetDeadline(time.Time{})
	var request [27]byte
	request[0] = Type
	request[1] = 1
	if fresh {
		request[2] = 1
	}
	copy(request[3:19], c.id[:])
	c.mu.Lock()
	binary.BigEndian.PutUint64(request[19:], c.received-uint64(len(c.rx)))
	c.mu.Unlock()
	if err := writeAll(wire, request[:]); err != nil {
		return err
	}
	var response [9]byte
	if _, err := io.ReadFull(wire, response[:]); err != nil {
		return err
	}
	if response[0] != 1 {
		return errors.New("resume: peer rejected stream (upgrade both ends or state expired)")
	}
	_ = wire.SetDeadline(time.Time{})
	return c.attach(wire, binary.BigEndian.Uint64(response[1:]))
}
func Dial(opener OpenFunc) (net.Conn, error) {
	c := newConn()
	if _, err := rand.Read(c.id[:]); err != nil {
		return nil, err
	}
	c.opener = opener
	wire, err := opener()
	if err != nil {
		return nil, err
	}
	if err = c.handshake(wire, true); err != nil {
		wire.Close()
		c.Close()
		return nil, err
	}
	return c, nil
}
func (c *Conn) reconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	for {
		select {
		case <-c.done:
			return
		case <-ctx.Done():
			c.Close()
			return
		case <-time.After(250 * time.Millisecond):
		}
		wire, err := c.opener()
		if err == nil {
			err = c.handshake(wire, false)
			if err == nil {
				return
			}
			wire.Close()
		}
	}
}

// Registry must be scoped to one authenticated peer identity by its caller.
// New connections are capped at 256: at most 64 MiB of payload buffers.
type Registry struct {
	mu      sync.Mutex
	streams map[[16]byte]*Conn
}

func NewRegistry() *Registry { return &Registry{streams: map[[16]byte]*Conn{}} }
func (r *Registry) Accept(wire net.Conn, serve func(net.Conn)) error {
	_ = wire.SetDeadline(time.Now().Add(8 * time.Second))
	defer wire.SetDeadline(time.Time{})
	var h [26]byte
	if _, err := io.ReadFull(wire, h[:]); err != nil {
		return err
	}
	if h[0] != 1 || h[1] > 1 {
		return errors.New("unsupported resume header")
	}
	fresh := h[1] == 1
	var id [16]byte
	copy(id[:], h[2:18])
	ack := binary.BigEndian.Uint64(h[18:])
	r.mu.Lock()
	c := r.streams[id]
	if fresh {
		if c != nil || len(r.streams) >= 256 {
			r.mu.Unlock()
			_ = writeAll(wire, make([]byte, 9))
			return errors.New("resume capacity or duplicate id")
		}
		c = newConn()
		c.id = id
		r.streams[id] = c
		c.onClose = func() {
			r.mu.Lock()
			if r.streams[id] == c {
				delete(r.streams, id)
			}
			r.mu.Unlock()
		}
	}
	r.mu.Unlock()
	if c == nil {
		_ = writeAll(wire, make([]byte, 9))
		return errors.New("resume state expired")
	}
	c.mu.Lock()
	rx := c.received - uint64(len(c.rx))
	valid := !c.closed && ack >= c.base && ack <= c.end
	c.mu.Unlock()
	if !valid {
		if fresh {
			c.Close()
		}
		return errors.New("invalid resume offset")
	}
	var response [9]byte
	response[0] = 1
	binary.BigEndian.PutUint64(response[1:], rx)
	if err := writeAll(wire, response[:]); err != nil {
		if fresh {
			c.Close()
		}
		return err
	}
	_ = wire.SetDeadline(time.Time{})
	if err := c.attach(wire, ack); err != nil {
		if fresh {
			c.Close()
		}
		return err
	}
	if fresh {
		go func() { defer c.Close(); serve(c) }()
	}
	return nil
}
