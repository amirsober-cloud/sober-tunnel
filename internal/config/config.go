package config

type Config struct {
	Role       string `json:"role"`
	ListenAddr string `json:"listen_addr"`

	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`

	MinConnections int `json:"min_connections"`
	MaxConnections int `json:"max_connections"`
}
