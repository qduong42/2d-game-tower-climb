package game

import "github.com/qduong42/2d-game-tower-climb/internal/schema"

const (
	MaxPlayers     = 3
	NumPlatforms   = 7                // 0=ground…3=handoff…6=TOP summit
	MidMaxPlatform = NumPlatforms / 2 // MID caps at platform 3; TOP spawns here and climbs to 6
)

// Player holds mutable per-player state.
type Player struct {
	ID           string
	Color        string
	Name         string
	Role         schema.Role
	ClimberIndex int              // 0 or 1 for climbers; -1 for base operator
	Platform     int              // 0–(NumPlatforms-1); base operator always stays at 0
	Tool         schema.ToolType  // tool currently carried ("" = none)
	HeldTools    []schema.ToolType // BASE inventory (nil for climbers)
	SelectedIdx  int              // BASE: index into HeldTools
	PrevKeys     schema.InputKeys // used for rising-edge detection
}

// GameState is the full authoritative state for one room at one tick.
type GameState struct {
	Tick         uint64
	Phase        schema.Phase
	Players      map[string]*Player
	RequiredTool schema.ToolType
}
