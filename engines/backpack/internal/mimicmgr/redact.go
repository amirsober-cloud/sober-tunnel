package mimicmgr

import (
	"encoding/json"
	"strings"
)

// RedactConfig is for display only; the original configuration is saved intact.
func RedactConfig(data []byte) []byte {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return []byte("[invalid configuration]")
	}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				name := strings.ToLower(k)
				if strings.Contains(name, "token") || strings.Contains(name, "password") || strings.Contains(name, "secret") || strings.Contains(name, "private_key") {
					x[k] = "[hidden]"
				} else {
					walk(child)
				}
			}
		case []any:
			for _, child := range x {
				walk(child)
			}
		}
	}
	walk(value)
	out, _ := json.MarshalIndent(value, "", "  ")
	return out
}
