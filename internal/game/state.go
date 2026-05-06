package game

// Player holds mutable per-player state.
type Player struct {
	ID    string
	X     float64
	Y     float64
	Color string
	Name  string
}

// GameState is the full authoritative state for one room at one tick.
// It is immutable by convention — Tick() returns a new value.
type GameState struct {
	Tick    uint64
	Players map[string]*Player
}
