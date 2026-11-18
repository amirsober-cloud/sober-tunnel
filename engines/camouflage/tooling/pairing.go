package tooling

import (
	"sort"
	"strconv"
	"strings"

	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/pairing"
)

// PairingDefaults returns the actual foreign settings for the Iran setup form.
// Legacy secrets contain no endpoint metadata and leave manual settings intact.
func PairingDefaults(token string) (map[string]string, error) {
	t, modern, err := pairing.Decode(strings.TrimSpace(token))
	if err != nil || !modern {
		return nil, err
	}
	values := map[string]string{"target-ip": t.ExitIP, "target": "", "persona": t.Persona, "tls-servername": t.SNI}
	names := make([]string, 0, len(t.Endpoints))
	for name, port := range t.Endpoints {
		names = append(names, name)
		values[name+"-port"] = strconv.Itoa(port)
	}
	sort.Strings(names)
	values["mimics"] = strings.Join(names, ",")
	return values, nil
}
