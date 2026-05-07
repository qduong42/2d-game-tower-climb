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
		// deep-copy slice so mutations don't alias the original
		if len(p.HeldTools) > 0 {
			cp.HeldTools = make([]schema.ToolType, len(p.HeldTools))
			copy(cp.HeldTools, p.HeldTools)
		}
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
			// MID: platforms 0–MidMaxPlatform; TOP: platforms MidMaxPlatform–NumPlatforms-1
			maxPlatform := NumPlatforms - 1
			minPlatform := 0
			if p.ClimberIndex == 0 {
				maxPlatform = MidMaxPlatform
			} else {
				minPlatform = MidMaxPlatform
			}
			if inp.Keys.Up && !prev.Up && p.Platform < maxPlatform {
				p.Platform++
			}
			if inp.Keys.Down && !prev.Down && p.Platform > minPlatform {
				p.Platform--
			}
		}

		if p.Role == schema.RoleBase && len(p.HeldTools) > 0 {
			// Up/Down cycles the selected tool (rising edge)
			if inp.Keys.Right && !prev.Right {
				p.SelectedIdx = (p.SelectedIdx + 1) % len(p.HeldTools)
			}
			if inp.Keys.Left && !prev.Left {
				p.SelectedIdx = (p.SelectedIdx - 1 + len(p.HeldTools)) % len(p.HeldTools)
			}
		}

		// Pass tool along the chain — SPACE tries UP first, then DOWN.
		// Up: base→MID, MID→TOP (same platform required).
		// Down: TOP→MID at MidMaxPlatform, MID→base at platform 0 (same platform required).
		if inp.Keys.Space && !prev.Space {
			passed := false

			// Try passing UP (ClimberIndex + 1)
			upTargetIdx := p.ClimberIndex + 1
			for otherID, other := range next.Players {
				if otherID == id || other.ClimberIndex != upTargetIdx || other.Platform != p.Platform {
					continue
				}
				if p.Role == schema.RoleBase && len(p.HeldTools) > 0 {
					selected := p.HeldTools[p.SelectedIdx]
					p.HeldTools = append(p.HeldTools[:p.SelectedIdx], p.HeldTools[p.SelectedIdx+1:]...)
					if p.SelectedIdx >= len(p.HeldTools) && len(p.HeldTools) > 0 {
						p.SelectedIdx = len(p.HeldTools) - 1
					}
					other.Tool = selected
					passed = true
					break
				}
				if p.Role == schema.RoleClimber && p.Tool != schema.ToolNone {
					other.Tool = p.Tool
					p.Tool = schema.ToolNone
					passed = true
					break
				}
			}

			// Try passing DOWN if UP didn't fire
			if !passed && p.Role == schema.RoleClimber && p.Tool != schema.ToolNone {
				canPassDown := (p.ClimberIndex == 1 && p.Platform == MidMaxPlatform) ||
					(p.ClimberIndex == 0 && p.Platform == 0)
				if canPassDown {
					downTargetIdx := p.ClimberIndex - 1
					for otherID, other := range next.Players {
						if otherID == id || other.ClimberIndex != downTargetIdx || other.Platform != p.Platform {
							continue
						}
						if other.Role == schema.RoleBase {
							other.HeldTools = append(other.HeldTools, p.Tool)
						} else {
							other.Tool = p.Tool
						}
						p.Tool = schema.ToolNone
						break
					}
				}
			}
		}

		p.PrevKeys = inp.Keys
	}

	// Win: TOP climber (index 1) at summit carrying any tool
	for _, p := range next.Players {
		if p.Role == schema.RoleClimber && p.ClimberIndex == 1 && p.Platform == NumPlatforms-1 && p.Tool != schema.ToolNone {
			next.Phase = schema.PhaseWon
			break
		}
	}

	return next
}
