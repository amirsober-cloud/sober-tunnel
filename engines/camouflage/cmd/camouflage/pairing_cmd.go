package main

import (
	"flag"
	"fmt"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/config"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/pairing"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/sysutil"
)

func cmdPairing(args []string) {
	fs := flag.NewFlagSet("pairing", flag.ExitOnError)
	publicIP := fs.String("public-ip", "", "public exit IP embedded in the pairing token (auto-detected if empty)")
	_ = fs.Parse(args)
	cfg, err := config.LoadConfig()
	if err != nil {
		fail("configuration: %v", err)
	}
	if cfg.Role != "foreign" {
		fail("pairing export requires a foreign exit")
	}
	ip := *publicIP
	if ip == "" {
		ip, err = sysutil.GetPublicIPv4()
		if err != nil {
			fail("public IP: %v", err)
		}
	}
	if err = validIP(ip); err != nil {
		fail("public IP: %v", err)
	}
	eps := map[string]int{}
	for _, m := range cfg.Mimics {
		eps[m.Type] = m.Port
	}
	sni := firstServerName(cfg.Mimics)
	if sni == "" {
		sni = cfg.Domain
	}
	fmt.Println(pairing.Encode(pairing.Token{ExitIP: ip, AuthKey: cfg.AuthToken, Persona: cfg.Persona, SNI: sni, Endpoints: eps}))
}

func cmdHostSSH(args []string) {
	fs := flag.NewFlagSet("host-ssh", flag.ExitOnError)
	port := fs.Int("port", 0, "new OpenSSH port")
	_ = fs.Parse(args)
	if err := validPort(*port); err != nil {
		fail("port: %v", err)
	}
	if err := sysutil.ChangeSSHPort(fmt.Sprint(*port)); err != nil {
		fail("OpenSSH relocation: %v", err)
	}
	fmt.Printf("OpenSSH configured on port %d\n", *port)
}
