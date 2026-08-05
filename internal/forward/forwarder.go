package forward

import (
	"net"

	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

type Forwarder struct {
	target string
}

func New(target string) *Forwarder {

	return &Forwarder{
		target: target,
	}
}

func (f *Forwarder) Handle(conn net.Conn) {

	targetConn, err := net.Dial(
		"tcp",
		f.target,
	)

	if err != nil {

		logger.Error.Printf(
			"Target connection failed: %v",
			err,
		)

		conn.Close()
		return
	}

	logger.Info.Printf(
		"Forwarding to %s",
		f.target,
	)

	Relay(
		conn,
		targetConn,
	)
}
