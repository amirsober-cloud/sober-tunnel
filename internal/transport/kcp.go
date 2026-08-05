package transport

import (
	"net"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/xtaci/kcp-go/v5"
)

type KCPTransport struct {
	cfg config.KCPConfig
}

func NewKCP(cfg config.KCPConfig) *KCPTransport {

	if cfg.MTU == 0 {
		cfg.MTU = 1350
	}

	if cfg.NoDelay == 0 {
		cfg.NoDelay = 1
	}

	if cfg.Interval == 0 {
		cfg.Interval = 10
	}

	if cfg.Resend == 0 {
		cfg.Resend = 2
	}

	if cfg.SndWnd == 0 {
		cfg.SndWnd = 1024
	}

	if cfg.RcvWnd == 0 {
		cfg.RcvWnd = 1024
	}

	return &KCPTransport{
		cfg: cfg,
	}
}

func (k *KCPTransport) Connect(address string) (net.Conn, error) {

	conn, err := kcp.DialWithOptions(
		address,
		nil,
		10,
		3,
	)

	if err != nil {
		return nil, err
	}

	session := conn

	session.SetNoDelay(
		k.cfg.NoDelay,
		k.cfg.Interval,
		k.cfg.Resend,
		k.cfg.NC,
	)

	session.SetWindowSize(
		k.cfg.SndWnd,
		k.cfg.RcvWnd,
	)

	session.SetMtu(
		k.cfg.MTU,
	)

	session.SetACKNoDelay(true)

	session.SetStreamMode(true)

	session.SetWriteDelay(false)

	return session, nil
}

func (k *KCPTransport) Listen(address string) (net.Listener, error) {

	listener, err := kcp.ListenWithOptions(
		address,
		nil,
		10,
		3,
	)

	if err != nil {
		return nil, err
	}

	return listener, nil
}
func ConfigureKCP(conn *kcp.UDPSession, cfg config.KCPConfig) {

	conn.SetNoDelay(
		cfg.NoDelay,
		cfg.Interval,
		cfg.Resend,
		cfg.NC,
	)

	conn.SetWindowSize(
		cfg.SndWnd,
		cfg.RcvWnd,
	)

	conn.SetMtu(
		cfg.MTU,
	)

	conn.SetACKNoDelay(true)

	conn.SetStreamMode(true)

	conn.SetWriteDelay(false)
}
func (k *KCPTransport) Close() error {
	return nil
}

func (k *KCPTransport) Name() string {
	return "kcp"
}
