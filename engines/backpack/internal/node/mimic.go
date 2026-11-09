package node

import (
	"context"
	"encoding/json"
	"github.com/backpack/backpack/internal/mimicmgr"
)

type MimicRequest struct {
	Name   string          `json:"name"`
	Action string          `json:"action"`
	Config json.RawMessage `json:"config,omitempty"`
}

func executeMimic(req Request) Response {
	var r MimicRequest
	if err := json.Unmarshal(req.Body, &r); err != nil {
		return failf("invalid camouflage request")
	}
	m := mimicmgr.Default
	ctx := context.Background()
	switch r.Action {
	case "list":
		items, err := m.List(ctx)
		if err != nil {
			return failf("%v", err)
		}
		return okBody(items)
	case "read":
		cfg, err := m.Read(r.Name)
		if err != nil {
			return failf("%v", err)
		}
		return okBody(map[string]any{"name": r.Name, "config": cfg})
	case "logs":
		logs, err := m.Logs(ctx, r.Name)
		if err != nil {
			return failf("%v", err)
		}
		return okBody(map[string]string{"logs": logs})
	case "save":
		if err := m.Save(ctx, r.Name, r.Config); err != nil {
			return failf("%v", err)
		}
	case "start", "stop", "restart", "delete":
		if err := m.Action(ctx, r.Name, r.Action); err != nil {
			return failf("%v", err)
		}
	default:
		return failf("unknown camouflage operation")
	}
	return okBody(map[string]bool{"ok": true})
}
