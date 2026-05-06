package game

import "github.com/qduong42/2d-game-tower-climb/internal/schema"

const (
	Speed  = 200.0 // pixels per second
	WorldW = 800.0
	WorldH = 600.0
)

// Tick advances game state by dt seconds given the latest inputs.
// It is a pure function — no side effects.
func Tick(state GameState, inputs map[string]schema.InputPayload, dt float64) GameState {
	next := GameState{
		Tick:    state.Tick + 1,
		Players: make(map[string]*Player, len(state.Players)),
	}
	for id, p := range state.Players {
		np := *p
		if inp, ok := inputs[id]; ok {
			if inp.Keys.Left {
				np.X -= Speed * dt
			}
			if inp.Keys.Right {
				np.X += Speed * dt
			}
			if inp.Keys.Up {
				np.Y -= Speed * dt
			}
			if inp.Keys.Down {
				np.Y += Speed * dt
			}
		}
		np.X = clamp(np.X, 0, WorldW)
		np.Y = clamp(np.Y, 0, WorldH)
		next.Players[id] = &np
	}
	return next
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
