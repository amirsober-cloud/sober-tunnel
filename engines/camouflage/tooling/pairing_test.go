package tooling

import (
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/pairing"
	"testing"
)

func TestPairingDefaultsUsesForeignPorts(t *testing.T) {
	token := pairing.Encode(pairing.Token{ExitIP: "203.0.113.10", AuthKey: "0123456789abcdef0123456789abcdef", Persona: "devops", SNI: "exit.example", Endpoints: map[string]int{"ssh": 22, "tls": 243}})
	v, err := PairingDefaults(token)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"target-ip": "203.0.113.10", "target": "", "persona": "devops", "tls-servername": "exit.example", "ssh-port": "22", "tls-port": "243", "mimics": "ssh,tls"} {
		if v[key] != want {
			t.Errorf("%s = %q, want %q", key, v[key], want)
		}
	}
}

func TestPairingDefaultsLegacyDoesNotInventEndpoints(t *testing.T) {
	v, err := PairingDefaults("0123456789abcdef0123456789abcdef")
	if err != nil || v != nil {
		t.Fatalf("legacy defaults = %v, %v", v, err)
	}
}
