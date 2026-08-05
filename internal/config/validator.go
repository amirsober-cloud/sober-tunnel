package config

import (
	"errors"
	"fmt"
)

func (c *Config) Validate() error {

	if c.Role != "client" && c.Role != "server" {
		return errors.New("invalid role")
	}

	if c.ListenAddr == "" {
		return errors.New("listen_addr is required")
	}

	if c.RemoteHost == "" {
		return errors.New("remote_host is required")
	}

	if c.RemotePort < 1 || c.RemotePort > 65535 {
		return fmt.Errorf("invalid remote_port: %d", c.RemotePort)
	}

	switch c.Transport {
	case "tcp":
		// OK
	default:
		return fmt.Errorf("unsupported transport: %s", c.Transport)
	}

	if c.MinConnections < 1 {
		return errors.New("min_connections must be greater than zero")
	}

	if c.MaxConnections < c.MinConnections {
		return errors.New("max_connections must be >= min_connections")
	}

	return nil
}
