package webui

import (
	"encoding/json"
	"github.com/backpack/backpack/internal/mimicmgr"
	"github.com/backpack/backpack/internal/node"
	"net/http"
)

var camouflage = mimicmgr.Default

func (s *server) handleMimic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if remote := r.URL.Query().Get("node"); remote != "" {
		s.handleRemoteMimic(w, r, remote)
		return
	}
	switch r.Method {
	case http.MethodGet:
		name := r.URL.Query().Get("name")
		if name == "" {
			items, err := camouflage.List(r.Context())
			if err != nil {
				http.Error(w, "could not list camouflage instances", 500)
				return
			}
			writeJSON(w, items)
			return
		}
		if mimicmgr.CheckName(name) != nil {
			http.Error(w, "invalid instance name", 400)
			return
		}
		if r.URL.Query().Get("logs") == "1" {
			logs, err := camouflage.Logs(r.Context(), name)
			if err != nil {
				http.Error(w, "could not read instance logs", 500)
				return
			}
			writeJSON(w, map[string]string{"logs": logs})
			return
		}
		cfg, err := camouflage.Read(name)
		if err != nil {
			http.Error(w, "instance not found", 404)
			return
		}
		writeJSON(w, map[string]any{"name": name, "config": cfg})
	case http.MethodPost:
		var req struct {
			Name   string          `json:"name"`
			Action string          `json:"action"`
			Config json.RawMessage `json:"config"`
		}
		if err := decodeJSON(w, r, &req); err != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		var err error
		if req.Action == "save" {
			err = camouflage.Save(r.Context(), req.Name, req.Config)
		} else {
			err = camouflage.Action(r.Context(), req.Name, req.Action)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", 405)
	}
}

func (s *server) handleRemoteMimic(w http.ResponseWriter, r *http.Request, remote string) {
	var req node.MimicRequest
	switch r.Method {
	case http.MethodGet:
		req.Name = r.URL.Query().Get("name")
		req.Action = "list"
		if req.Name != "" {
			req.Action = "read"
			if r.URL.Query().Get("logs") == "1" {
				req.Action = "logs"
			}
		}
	case http.MethodPost:
		if err := decodeJSON(w, r, &req); err != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		if req.Action != "save" && req.Action != "start" && req.Action != "stop" && req.Action != "restart" && req.Action != "delete" {
			http.Error(w, "unknown action", 400)
			return
		}
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", 405)
		return
	}
	if req.Action != "list" {
		if err := mimicmgr.CheckName(req.Name); err != nil {
			http.Error(w, "invalid instance name", 400)
			return
		}
	}
	run := s.nodes.Runner()
	if run == nil {
		http.Error(w, "fleet unavailable", 503)
		return
	}
	var result json.RawMessage
	if err := run.Call(remote, node.OpMimic, req, &result); err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	writeJSON(w, result)
}
