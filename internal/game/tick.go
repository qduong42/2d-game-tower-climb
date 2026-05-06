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
			// MID climber (index 0) is capped at MidMaxPlatform; TOP climber (index 1) goes to the top
			maxPlatform := NumPlatforms - 1
			if p.ClimberIndex == 0 {
				maxPlatform = MidMaxPlatform
			}
			// Rising edge only — one platform step per keypress
			if inp.Keys.Up && !prev.Up && p.Platform < maxPlatform {
				p.Platform++
			}
			if inp.Keys.Down && !prev.Down && p.Platform > 0 {
				p.Platform--
			}
		}

		// Pass tool along the chain: base(-1) → MID(0) → TOP(1).
		// Must be on the same platform as the target.
		if inp.Keys.Space && !prev.Space && p.HasTool {
			targetIndex := p.ClimberIndex + 1 // base(-1)→0, MID(0)→1
			for otherID, other := range next.Players {
				if otherID == id {
					continue
				}
				if other.ClimberIndex == targetIndex && other.Platform == p.Platform {
					p.HasTool = false
					other.HasTool = true
					break
				}
			}
		}

		p.PrevKeys = inp.Keys
	}

	// Win: only the TOP climber (index 1) reaching the top platform with the tool wins
	for _, p := range next.Players {
		if p.Role == schema.RoleClimber && p.ClimberIndex == 1 && p.Platform == NumPlatforms-1 && p.HasTool {
			next.Phase = schema.PhaseWon
			break
		}
	}

	return next
}
