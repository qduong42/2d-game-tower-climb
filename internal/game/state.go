package game

import "github.com/qduong42/2d-game-tower-climb/internal/schema"

const (
	MaxPlayers     = 3
	NumPlatforms   = 4             // 0=ground, 1, 2, 3=top
	MidMaxPlatform = NumPlatforms - 2 // ClimberIndex 0 (MID) caps here
)

// Player holds mutable per-player state.
type Player struct {
	ID           string
	Color        string
	Name         string
	Role         schema.Role
	ClimberIndex int          // 0 or 1 for climbers; -1 for base operator
	Platform     int          // 0–(NumPlatforms-1); base operator always stays at 0
	HasTool      bool
	PrevKeys     schema.InputKeys // used for rising-edge detection
}

// GameState is the full authoritative state for one room at one tick.
type GameState struct {
	Tick    uint64
	Phase   schema.Phase
	Players map[string]*Player
}
