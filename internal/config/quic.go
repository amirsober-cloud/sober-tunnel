package config

type QUICConfig struct {
	MaxIdleTimeout int `json:"max_idle_timeout"`

	KeepAlive bool `json:"keep_alive"`
}
