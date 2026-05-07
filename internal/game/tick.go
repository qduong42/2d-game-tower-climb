package game

import (
	"math/rand/v2"

	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

const windKnockInterval = 2 * TicksPerSec // knock-down every 2 s during active gust

// Tick advances game state by dt seconds given the latest inputs.
// It is a pure function — no side effects.
func Tick(state GameState, inputs map[string]schema.InputPayload, dt float64) GameState {
	next := GameState{
		Tick:             state.Tick + 1,
		Phase:            state.Phase,
		Players:          make(map[string]*Player, len(state.Players)),
		RequiredTool:     state.RequiredTool,
		WindPhase:        state.WindPhase,
		WindTicksLeft:    state.WindTicksLeft,
		WindCooldownLeft: state.WindCooldownLeft,
		WindKnockTicks:   state.WindKnockTicks,
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
					if other.Tool != schema.ToolNone {
						continue // target already has a tool, skip
					}
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
					if other.Tool != schema.ToolNone {
						continue // target already has a tool, skip
					}
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
							// Guard: don't append a tool BASE already holds — prevents
							// duplicates when the same tool exists in both BASE and MID
							// (e.g. same-tick pass+return with certain map iteration order).
							alreadyHeld := false
							for _, t := range other.HeldTools {
								if t == p.Tool {
									alreadyHeld = true
									break
								}
							}
							if !alreadyHeld {
								other.HeldTools = append(other.HeldTools, p.Tool)
							}
						} else {
							if other.Tool != schema.ToolNone {
								continue // target already has a tool, skip
							}
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

	// Wind gust state machine — only runs while playing.
	if state.Phase == schema.PhasePlaying {
		switch state.WindPhase {
		case "", schema.WindNone:
			if next.WindCooldownLeft > 0 {
				next.WindCooldownLeft--
			} else {
				next.WindPhase = schema.WindWarning
				next.WindTicksLeft = TicksPerSec*5 + rand.IntN(TicksPerSec*5)
			}
		case schema.WindWarning:
			if next.WindTicksLeft > 0 {
				next.WindTicksLeft--
			} else {
				next.WindPhase = schema.WindActive
				next.WindTicksLeft = TicksPerSec*5 + rand.IntN(TicksPerSec*5)
				next.WindKnockTicks = windKnockInterval
			}
		case schema.WindActive:
			next.WindTicksLeft--
			next.WindKnockTicks--
			if next.WindKnockTicks <= 0 {
				for id, p := range next.Players {
					if p.Role != schema.RoleClimber {
						continue
					}
					inp := inputs[id]
					if !inp.Keys.Brace {
						minPlat := 0
						if p.ClimberIndex == 1 {
							minPlat = MidMaxPlatform
						}
						if p.Platform > minPlat {
							p.Platform--
						}
					}
				}
				next.WindKnockTicks = windKnockInterval
			}
			if next.WindTicksLeft <= 0 {
				next.WindPhase = schema.WindNone
				next.WindTicksLeft = 0
				next.WindCooldownLeft = TicksPerSec*10 + rand.IntN(TicksPerSec*10)
			}
		}
	}

	// Win: TOP climber (index 1) at summit carrying the required tool (never ToolNone).
	for _, p := range next.Players {
		if p.Role == schema.RoleClimber && p.ClimberIndex == 1 && p.Platform == NumPlatforms-1 &&
			state.RequiredTool != schema.ToolNone && p.Tool == state.RequiredTool {
			next.Phase = schema.PhaseWon
			break
		}
	}

	return next
}
