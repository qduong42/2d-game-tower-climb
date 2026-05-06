package game

import "github.com/qduong42/2d-game-tower-climb/internal/schema"

// Tick advances game state by dt seconds given the latest inputs.
// It is a pure function — no side effects.
func Tick(state GameState, inputs map[string]schema.InputPayload, dt float64) GameState {
	next := GameState{
		Tick:    state.Tick + 1,
		Phase:   state.Phase,
		Players: make(map[string]*Player, len(state.Players)),
	}
	for id, p := range state.Players {
		cp := *p
		next.Players[id] = &cp
	}

	if state.Phase != schema.PhasePlaying {
		return next
	}

	for id, inp := range inputs {
		p, ok := next.Players[id]
		if !ok {
			continue
		}
		prev := p.PrevKeys

		if p.Role == schema.RoleClimber {
			// Rising edge only — one platform step per keypress
			if inp.Keys.Up && !prev.Up && p.Platform < NumPlatforms-1 {
				p.Platform++
			}
			if inp.Keys.Down && !prev.Down && p.Platform > 0 {
				p.Platform--
			}
		}

		// Pass tool to any other player at the same platform (rising edge)
		if inp.Keys.Space && !prev.Space && p.HasTool {
			for otherID, other := range next.Players {
				if otherID == id {
					continue
				}
				if other.Platform == p.Platform {
					p.HasTool = false
					other.HasTool = true
					break
				}
			}
		}

		p.PrevKeys = inp.Keys
	}

	// Win: climber reaches top platform with tool
	for _, p := range next.Players {
		if p.Role == schema.RoleClimber && p.Platform == NumPlatforms-1 && p.HasTool {
			next.Phase = schema.PhaseWon
			break
		}
	}

	return next
}
