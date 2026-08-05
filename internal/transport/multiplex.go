package transport

import "net"

type MultiplexTransport interface {
	Transport

	OpenStream() (net.Conn, error)

	AcceptStream() (net.Conn, error)

	SupportsMultiplex() bool
}
