package mimicmgr

import (
	"encoding/json"
	"time"
)

type TrafficSummary struct {
	Sent     uint64            `json:"sent"`
	Received uint64            `json:"received"`
	Total    int               `json:"total"`
	Running  int               `json:"running"`
	Days     map[string]uint64 `json:"days"`
}

func (m *Manager) TrafficSummary() TrafficSummary {
	result := TrafficSummary{Days: map[string]uint64{}}
	for _, name := range m.Names() {
		result.Total++
		data, err := m.Traffic(name)
		if err != nil {
			continue
		}
		var snap struct {
			Sent, Received uint64
			Timestamp      int64
			Days           map[string]uint64
		}
		if json.Unmarshal(data, &snap) != nil {
			continue
		}
		result.Sent += snap.Sent
		result.Received += snap.Received
		if delta := time.Now().UnixMilli() - snap.Timestamp; delta >= 0 && delta < 5000 {
			result.Running++
		}
		for day, value := range snap.Days {
			result.Days[day] += value
		}
	}
	return result
}
