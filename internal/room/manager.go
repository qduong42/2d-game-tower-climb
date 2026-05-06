package room

import "sync"

// Manager maintains the active rooms indexed by room code.
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

// GetOrCreate returns the existing room for code, or starts a new one.
func (m *Manager) GetOrCreate(code string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rooms[code]; ok {
		return r
	}
	r := newRoom(code)
	m.rooms[code] = r
	go r.run()
	return r
}

// Remove stops and removes a room. Safe to call when room is already gone.
func (m *Manager) Remove(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rooms[code]; ok {
		r.stop()
		delete(m.rooms, code)
	}
}
