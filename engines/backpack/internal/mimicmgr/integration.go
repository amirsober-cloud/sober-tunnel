package mimicmgr

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hedioum/Hedioum-Pool-Tunnel/config"
	"github.com/hedioum/Hedioum-Pool-Tunnel/tooling"
)

type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

// Preflight checks availability without modifying SSH, firewall or routing.
// A successful bind check is a snapshot; the service still owns final binding.
func (m *Manager) Preflight(ctx context.Context, name string) ([]Check, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	if err := CheckName(name); err != nil {
		return nil, err
	}
	cfg, err := config.LoadFile(m.configPath(name))
	if err != nil {
		return nil, err
	}
	checks := []Check{{Name: "Configuration", OK: true, Detail: "Native configuration validation passed"}}
	active, _ := m.command(ctx, "systemctl", "is-active", m.unit(name))
	running := strings.TrimSpace(string(active)) == "active"
	checks = append(checks, Check{Name: "Native tools", OK: fileExists(m.nativeBinary()), Detail: m.nativeBinary()})
	listen := func(label, host string, port int) {
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		if running {
			dialHost := host
			if host == "" || host == "0.0.0.0" || host == "localhost" {
				dialHost = "127.0.0.1"
			}
			if host == "::" {
				dialHost = "::1"
			}
			d := net.Dialer{Timeout: 2 * time.Second}
			conn, e := d.DialContext(ctx, "tcp", net.JoinHostPort(dialHost, strconv.Itoa(port)))
			if e != nil {
				checks = append(checks, Check{Name: label, Detail: addr + " service active but listener unreachable: " + e.Error()})
				return
			}
			conn.Close()
			checks = append(checks, Check{Name: label, OK: true, Detail: addr + " TCP listener reachable; run Probe to authenticate the tunnel"})
			return
		}
		lc := net.ListenConfig{}
		c, e := lc.Listen(ctx, "tcp", addr)
		if e != nil {
			checks = append(checks, Check{Name: label, Detail: addr + ": " + e.Error()})
			return
		}
		c.Close()
		checks = append(checks, Check{Name: label, OK: true, Detail: addr + " available now"})
	}
	if cfg.Role == "foreign" {
		for _, ep := range cfg.Mimics {
			listen(ep.Type, "0.0.0.0", ep.Port)
		}
		if cfg.HTTPDecoyPort > 0 {
			listen("HTTP decoy", "0.0.0.0", cfg.HTTPDecoyPort)
		}
		if cfg.Domain != "" {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			ips, e := net.DefaultResolver.LookupHost(ctx, cfg.Domain)
			checks = append(checks, Check{Name: "Certificate domain DNS", OK: e == nil && len(ips) > 0, Detail: fmt.Sprintf("%s: %v (must point at this exit)", cfg.Domain, ips)})
		}
	} else {
		for _, n := range cfg.ForeignNodes {
			listen(n.Alias+" SOCKS5", n.SocksBind, n.LocalSocksPort)
			if n.TunEnabled || n.GatewayEnabled {
				checks = append(checks, Check{Name: n.Alias + " TUN device", OK: fileExists("/dev/net/tun"), Detail: "Requires /dev/net/tun and CAP_NET_ADMIN"})
			}
		}
	}
	return checks, nil
}
func fileExists(path string) bool { _, e := os.Stat(path); return e == nil }

func (m *Manager) HostAction(ctx context.Context, r ToolRequest) (any, error) {
	if err := CheckName(r.Name); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cfg, err := config.LoadFile(m.configPath(r.Name))
	if err != nil {
		return nil, err
	}
	var args []string
	confirm := ""
	switch r.Command {
	case "move-ssh":
		if cfg.Role != "foreign" {
			return nil, fmt.Errorf("SSH decoy belongs to a foreign exit")
		}
		confirm = "MOVE SSH"
		args = []string{"host-ssh", "--port=" + strconv.Itoa(cfg.DecoyPort)}
	case "open-firewall":
		if cfg.Role != "foreign" {
			return nil, fmt.Errorf("firewall action belongs to a foreign exit")
		}
		confirm = "OPEN PORTS"
		args = []string{"--open-firewall"}
	case "bbr":
		confirm = "ENABLE BBR"
		args = []string{"host-bbr"}
	default:
		return nil, fmt.Errorf("unknown host action")
	}
	if r.Confirm != confirm {
		return nil, fmt.Errorf("type %s to confirm this host operation", confirm)
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, m.nativeBinary(), args...)
	cmd.Env = append(os.Environ(), "SOBER_MIMIC_CONFIG="+m.configPath(r.Name), "NO_COLOR=1", "TERM=dumb")
	var output bytes.Buffer
	w := &tooling.CappedWriter{W: &output, Remaining: 64 << 10}
	cmd.Stdout = w
	cmd.Stderr = w
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("host operation failed: %s (%w)", output.String(), err)
	}
	return map[string]string{"output": output.String()}, nil
}

func (m *Manager) Integration(name string) (any, error) {
	if err := CheckName(name); err != nil {
		return nil, err
	}
	cfg, err := config.LoadFile(m.configPath(name))
	if err != nil {
		return nil, err
	}
	if cfg.Role != "iran" {
		return nil, fmt.Errorf("x-ui connects to the Iran hub's local SOCKS5 listener")
	}
	result := []any{}
	for _, n := range cfg.ForeignNodes {
		host := n.SocksBind
		if host == "" || host == "0.0.0.0" || host == "localhost" {
			host = "127.0.0.1"
		}
		if host == "::" {
			host = "::1"
		}
		tag := "sober-" + name + "-" + n.Alias
		server := map[string]any{"address": host, "port": n.LocalSocksPort}
		out := func(settings any) any { return map[string]any{"tag": tag, "protocol": "socks", "settings": settings} }
		rule := map[string]any{"type": "field", "inboundTag": []string{"REPLACE_WITH_YOUR_INBOUND_TAG"}, "network": "tcp,udp", "outboundTag": tag}
		result = append(result, map[string]any{"alias": n.Alias, "address": host, "port": n.LocalSocksPort, "legacy_outbound": out(map[string]any{"servers": []any{server}}), "modern_outbound": out(server), "routing_rule": rule, "domain_strategy": "AsIs"})
	}
	return result, nil
}
