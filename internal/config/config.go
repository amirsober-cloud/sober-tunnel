package config

type Config struct {
	Role string `json:"role"`

	LocalAddr string `json:"local_addr"`
	LocalPort int    `json:"local_port"`

	ListenAddr string `json:"listen_addr"`
	ListenPort int    `json:"listen_port"`

	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`

	TargetHost string `json:"target_host"`
	TargetPort int    `json:"target_port"`

	Transport string `json:"transport"`
	LogLevel  string `json:"log_level"`

	MinConnections int `json:"min_connections"`
	MaxConnections int `json:"max_connections"`

	KCP  KCPConfig  `json:"kcp"`
	QUIC QUICConfig `json:"quic"`
}
