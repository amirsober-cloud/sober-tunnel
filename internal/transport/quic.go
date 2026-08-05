package transport

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/quic-go/quic-go"
)

type QUICTransport struct {
	cfg config.QUICConfig

	listener *quic.Listener

	mu sync.Mutex
}

func NewQUIC(cfg config.QUICConfig) *QUICTransport {

	if cfg.MaxIdleTimeout == 0 {
		cfg.MaxIdleTimeout = 60
	}

	return &QUICTransport{
		cfg: cfg,
	}
}

func (q *QUICTransport) Connect(address string) (net.Conn, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,

		NextProtos: []string{
			"sober-tunnel",
		},
	}

	session, err := quic.DialAddr(
		context.Background(),
		address,
		tlsConfig,
		q.config(),
	)

	if err != nil {
		return nil, err
	}

	stream, err := session.OpenStreamSync(
		context.Background(),
	)

	if err != nil {
		session.CloseWithError(
			0,
			"stream open failed",
		)

		return nil, err
	}

	return NewQUICConn(
		stream,
		session.RemoteAddr(),
	), nil
}

func (q *QUICTransport) Listen(address string) (net.Listener, error) {

	tlsConfig := generateTLSConfig()

	listener, err := quic.ListenAddr(
		address,
		tlsConfig,
		q.config(),
	)

	if err != nil {
		return nil, err
	}

	q.listener = listener

	return &quicListener{
		listener: listener,
	}, nil
}

func (q *QUICTransport) config() *quic.Config {

	cfg := &quic.Config{

		MaxIdleTimeout: time.Duration(
			q.cfg.MaxIdleTimeout,
		) * time.Second,

		EnableDatagrams: false,
	}

	if q.cfg.KeepAlive {

		cfg.KeepAlivePeriod = 20 * time.Second

	}

	return cfg
}

func (q *QUICTransport) Close() error {

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.listener != nil {
		return q.listener.Close()
	}

	return nil
}
func (q *QUICTransport) Name() string {

	return "quic"
}

type quicListener struct {
	listener *quic.Listener
}

func (l *quicListener) Accept() (net.Conn, error) {

	session, err := l.listener.Accept(
		context.Background(),
	)

	if err != nil {
		return nil, err
	}

	stream, err := session.AcceptStream(
		context.Background(),
	)

	if err != nil {

		session.CloseWithError(
			0,
			"stream accept failed",
		)

		return nil, err
	}

	return NewQUICConn(
		stream,
		session.RemoteAddr(),
	), nil
}

func (l *quicListener) Close() error {

	return l.listener.Close()
}

func (l *quicListener) Addr() net.Addr {

	return l.listener.Addr()
}

func generateTLSConfig() *tls.Config {

	cert, err := generateCertificate()

	if err != nil {
		panic(err)
	}

	return &tls.Config{

		Certificates: []tls.Certificate{
			cert,
		},

		NextProtos: []string{
			"sober-tunnel",
		},
	}
}
