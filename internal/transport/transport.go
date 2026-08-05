package transport

import "net"

type Transport interface {
	Connect(address string) (net.Conn, error)
	Close() error
	Name() string
}
