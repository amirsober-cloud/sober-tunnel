package transport

import (
	"fmt"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
)

func NewTransport(
	name string,
	kcpConfig config.KCPConfig,
	quicConfig config.QUICConfig,
) (Transport, error) {

	switch name {

	case "tcp":
		return NewTCP(), nil

	case "kcp":
		return NewKCP(kcpConfig), nil

	case "quic":
		return NewQUIC(quicConfig), nil

	default:
		return nil, fmt.Errorf(
			"unknown transport: %s",
			name,
		)
	}
}
