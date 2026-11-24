// Package firewall best-effort opens a TCP listen port on whichever firewall the
// host is running, so users are not silently blocked when the tunnel binds a
// non-default port. It is distribution-agnostic (ufw, firewalld, iptables, and
// the iptables-nft shim that fronts nftables on most modern systems) and every
// operation is idempotent, so it is safe to run on every daemon start.
//
// It is intended to be invoked from a fully-privileged context (e.g. a systemd
// ExecStartPre=+ step), NOT from inside the hardened daemon sandbox, because
// editing the firewall requires CAP_NET_ADMIN and filesystem write access that
// the sandboxed daemon deliberately drops.
package firewall

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// EnsurePortOpen opens tcp/<port> on the active firewall. It returns the backend
// that handled it ("ufw", "firewalld", "iptables", or "none" when no restrictive
// firewall was detected) and an error only if a detected backend failed.
func EnsurePortOpen(port int) (string, error) {
	switch {
	case ufwActive():
		return "ufw", runQuiet("ufw", "allow", fmt.Sprintf("%d/tcp", port))

	case firewalldActive():
		// Apply to both stores directly so other services' runtime rules
		// survive opening this port.
		arg := fmt.Sprintf("--add-port=%d/tcp", port)
		errPermanent := runQuiet("firewall-cmd", "--permanent", arg)
		errRuntime := runQuiet("firewall-cmd", arg)
		return "firewalld", errors.Join(errPermanent, errRuntime)

	default:
		return ensureIptablesPortOpen(port)
	}
}

func ensureIptablesPortOpen(port int) (string, error) {
	// Try both families independently. A failed read cannot establish that
	// the host has no restrictive firewall.
	backend := "none"
	var failures []error
	for _, bin := range []string{"iptables", "ip6tables"} {
		restrictive, err := iptablesRestrictive(bin)
		if err != nil {
			failures = append(failures, err)
		}
		if restrictive {
			backend = "iptables"
			if err := ensureIptablesAccept(bin, port); err != nil {
				failures = append(failures, err)
			}
		}
	}
	if len(failures) > 0 {
		return "iptables", errors.Join(failures...)
	}
	return backend, nil
}

func ufwActive() bool {
	if !hasCmd("ufw") {
		return false
	}
	out, err := exec.Command("ufw", "status").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "Status: active")
}

func firewalldActive() bool {
	if !hasCmd("firewall-cmd") {
		return false
	}
	// `--state` exits 0 and prints "running" only when firewalld is up.
	out, err := exec.Command("firewall-cmd", "--state").CombinedOutput()
	return err == nil && strings.Contains(string(out), "running")
}

// iptablesRestrictive reports whether the given backend (iptables for IPv4 or
// ip6tables for IPv6) is present and its INPUT chain would drop unsolicited
// traffic (default policy DROP/REJECT), i.e. a new port actually needs an
// explicit ACCEPT rule.
func iptablesRestrictive(bin string) (bool, error) {
	if !hasCmd(bin) {
		return false, nil
	}
	out, err := exec.Command(bin, "-w", "5", "-S", "INPUT").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s INPUT inspection: %w: %s", bin, err, strings.TrimSpace(string(out)))
	}
	return parseInputRestrictive(string(out)), nil
}

// parseInputRestrictive inspects `iptables -S INPUT` output for a restrictive
// default policy. Kept separate for unit testing.
func parseInputRestrictive(rules string) bool {
	for _, line := range strings.Split(rules, "\n") {
		line = strings.TrimSpace(line)
		if line == "-P INPUT DROP" || line == "-P INPUT REJECT" {
			return true
		}
	}
	return false
}

// ensureIptablesAccept adds an ACCEPT rule for tcp/<port> to INPUT on the given
// backend (iptables/ip6tables) unless one is already present (checked with -C for
// idempotency).
func ensureIptablesAccept(bin string, port int) error {
	rule := []string{"INPUT", "-p", "tcp", "--dport", strconv.Itoa(port), "-j", "ACCEPT"}
	check := append([]string{"-w", "5", "-C"}, rule...)
	if err := exec.Command(bin, check...).Run(); err == nil {
		return nil // already allowed
	} else {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return fmt.Errorf("%s %s: %w", bin, strings.Join(check, " "), err)
		}
	}
	return runQuiet(bin, append([]string{"-w", "5", "-I"}, rule...)...)
}

func hasCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runQuiet(name string, args ...string) error {
	if err := exec.Command(name, args...).Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
