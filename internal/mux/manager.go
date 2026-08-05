package mux

import "sync"

type Manager struct {
	mu sync.RWMutex

	sessions map[string]*Session
}

func NewManager() *Manager {

	return &Manager{
		sessions: make(map[string]*Session),
	}
}

func (m *Manager) AddSession(
	id string,
	session *Session,
) {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[id] = session
}

func (m *Manager) RemoveSession(
	id string,
) {

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(
		m.sessions,
		id,
	)
}

func (m *Manager) GetSession(
	id string,
) *Session {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.sessions[id]
}

func (m *Manager) Count() int {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sessions)
}
