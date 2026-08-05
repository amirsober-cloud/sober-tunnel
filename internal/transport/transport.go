package transport

import "net"

type Transport interface {
	Connect(address string) (net.Conn, error)

	Listen(address string) (net.Listener, error)

	Close() error

	Name() string
}
