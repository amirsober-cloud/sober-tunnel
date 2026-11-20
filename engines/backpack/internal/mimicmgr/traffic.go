package mimicmgr

import (
	"encoding/json"
	"os"
)

// Traffic contains only counters published by this instance's independent core.
func (m *Manager) Traffic(name string) (json.RawMessage, error) {
	if err := CheckName(name); err != nil {
		return nil, err
	}
	if _, err := os.Stat(m.configPath(name)); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(m.configPath(name) + ".traffic.json")
	if os.IsNotExist(err) {
		return json.RawMessage(`{"sent":0,"received":0,"timestamp":0,"started":0}`), nil
	}
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, os.ErrInvalid
	}
	return json.RawMessage(data), nil
}
