package room

import (
	"sync"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
)

// Manager maintains the active rooms indexed by room code.
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

// GetOrCreate returns the existing room for code, or starts a new public one.
// To create a private room use GetOrCreateWithPrivacy.
func (m *Manager) GetOrCreate(code string) *Room {
	return m.GetOrCreateWithPrivacy(code, false)
}

// GetOrCreateWithPrivacy returns the existing room for code, or starts a new
// one with the supplied privacy setting.  The isPrivate flag is only applied
// when the room is first created; existing rooms are returned as-is.
func (m *Manager) GetOrCreateWithPrivacy(code string, isPrivate bool) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rooms[code]; ok {
		return r
	}
	r := newRoom(code, isPrivate)
	m.rooms[code] = r
	go r.run()
	return r
}

// RoomInfo is a lightweight snapshot used by the room browser endpoint.
type RoomInfo struct {
	Code    string `json:"code"`
	Players int    `json:"players"`
}

// ListPublicOpen returns info about all rooms that are public and not yet full.
func (m *Manager) ListPublicOpen() []RoomInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]RoomInfo, 0, len(m.rooms))
	for _, r := range m.rooms {
		if r.IsPrivate() {
			continue
		}
		count := r.PlayerCount()
		if count >= game.MaxPlayers {
			continue
		}
		out = append(out, RoomInfo{Code: r.Code(), Players: count})
	}
	return out
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
