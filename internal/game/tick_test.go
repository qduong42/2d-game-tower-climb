package game_test

import (
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

func playingState(players map[string]*game.Player) game.GameState {
	return game.GameState{Tick: 0, Phase: schema.PhasePlaying, Players: players}
}

// climber creates a MID climber (ClimberIndex=0, capped at MidMaxPlatform).
func climber(id string, platform int, tool schema.ToolType) *game.Player {
	return &game.Player{
		ID: id, Color: "#fff", Name: "test",
		Role: schema.RoleClimber, ClimberIndex: 0,
		Platform: platform, Tool: tool,
	}
}

// topClimber creates a TOP climber (ClimberIndex=1, can reach NumPlatforms-1).
func topClimber(id string, platform int, tool schema.ToolType) *game.Player {
	return &game.Player{
		ID: id, Color: "#fff", Name: "test",
		Role: schema.RoleClimber, ClimberIndex: 1,
		Platform: platform, Tool: tool,
	}
}

// baseOp creates a BASE operator with the given inventory of tools.
func baseOp(id string, tools ...schema.ToolType) *game.Player {
	return &game.Player{
		ID: id, Color: "#fff", Name: "base",
		Role: schema.RoleBase, ClimberIndex: -1,
		Platform: 0, HeldTools: tools,
	}
}

func TestTick_ClimberMovesUpOnUpKey(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 0, schema.ToolNone)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != 1 {
		t.Errorf("expected platform 1, got %d", next.Players["p1"].Platform)
	}
}

func TestTick_ClimberMovesDownOnDownKey(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 2, schema.ToolNone)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Down: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != 1 {
		t.Errorf("expected platform 1, got %d", next.Players["p1"].Platform)
	}
}

func TestTick_ClimberCannotGoBelowGround(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 0, schema.ToolNone)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Down: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != 0 {
		t.Errorf("expected platform 0, got %d", next.Players["p1"].Platform)
	}
}

func TestTick_MidClimberCannotGoAboveMidMax(t *testing.T) {
	mid := game.MidMaxPlatform
	state := playingState(map[string]*game.Player{"p1": climber("p1", mid, schema.ToolNone)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != mid {
		t.Errorf("MID climber: expected platform %d, got %d", mid, next.Players["p1"].Platform)
	}
}

func TestTick_TopClimberCanReachTop(t *testing.T) {
	top := game.NumPlatforms - 1
	state := playingState(map[string]*game.Player{"p1": topClimber("p1", top-1, schema.ToolNone)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != top {
		t.Errorf("TOP climber: expected platform %d, got %d", top, next.Players["p1"].Platform)
	}
}

func TestTick_MovementIsRisingEdgeOnly(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 0, schema.ToolNone)})
	inp := schema.InputPayload{Keys: schema.InputKeys{Up: true}}

	next1 := game.Tick(state, map[string]schema.InputPayload{"p1": inp}, 1.0/30.0)
	if next1.Players["p1"].Platform != 1 {
		t.Fatalf("tick1: expected platform 1, got %d", next1.Players["p1"].Platform)
	}

	// Up still held — should NOT move again
	next2 := game.Tick(next1, map[string]schema.InputPayload{"p1": inp}, 1.0/30.0)
	if next2.Players["p1"].Platform != 1 {
		t.Errorf("tick2: expected platform 1 (held key, no move), got %d", next2.Players["p1"].Platform)
	}
}

func TestTick_BasePassesToolToMidOnSamePlatform(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"base": baseOp("base", schema.ToolWrench),
		"c1":   climber("c1", 0, schema.ToolNone),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if len(next.Players["base"].HeldTools) != 0 {
		t.Error("base should no longer hold any tools after passing")
	}
	if next.Players["c1"].Tool != schema.ToolWrench {
		t.Errorf("c1 should hold wrench, got %q", next.Players["c1"].Tool)
	}
}

func TestTick_BaseSelectsToolWithLeftRight(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"base": baseOp("base", schema.ToolWrench, schema.ToolHammer),
	})
	// Press Right → SelectedIdx cycles from 0 to 1 (hammer)
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Right: true}}}, 1.0/30.0)
	if next.Players["base"].SelectedIdx != 1 {
		t.Errorf("expected SelectedIdx 1, got %d", next.Players["base"].SelectedIdx)
	}
	// Press Left → cycles back to 0 (wrench)
	next2 := game.Tick(next, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Left: true}}}, 1.0/30.0)
	if next2.Players["base"].SelectedIdx != 0 {
		t.Errorf("expected SelectedIdx 0, got %d", next2.Players["base"].SelectedIdx)
	}
}

func TestTick_BasePassesSelectedTool(t *testing.T) {
	// BASE has wrench(0) and hammer(1); select hammer(1), then pass
	base := baseOp("base", schema.ToolWrench, schema.ToolHammer)
	base.SelectedIdx = 1
	state := playingState(map[string]*game.Player{
		"base": base,
		"c1":   climber("c1", 0, schema.ToolNone),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["c1"].Tool != schema.ToolHammer {
		t.Errorf("c1 should hold hammer, got %q", next.Players["c1"].Tool)
	}
	if len(next.Players["base"].HeldTools) != 1 || next.Players["base"].HeldTools[0] != schema.ToolWrench {
		t.Errorf("base should still hold wrench, got %v", next.Players["base"].HeldTools)
	}
}

func TestTick_SpaceDoesNotPassIfNoOneSamePlatform(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"base": baseOp("base", schema.ToolWrench),
		"c1":   climber("c1", 1, schema.ToolNone), // different platform
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if len(next.Players["base"].HeldTools) != 1 {
		t.Error("base should still hold tool — no one at same platform")
	}
}

func TestTick_SpaceDoesNotSkipChain(t *testing.T) {
	// Base cannot pass directly to TOP climber (index 1) — must go through MID (index 0)
	state := playingState(map[string]*game.Player{
		"base": baseOp("base", schema.ToolWrench),
		"top":  topClimber("top", 0, schema.ToolNone), // same platform as base, but wrong index in chain
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if len(next.Players["base"].HeldTools) != 1 {
		t.Error("base should not be able to pass directly to TOP — MID must be in chain")
	}
}

func TestTick_MidPassesToolToTopAtHandoffPlatform(t *testing.T) {
	// TOP climbs up to MidMaxPlatform to meet MID and receive the tool
	state := playingState(map[string]*game.Player{
		"mid": climber("mid", game.MidMaxPlatform, schema.ToolWrench),
		"top": topClimber("top", game.MidMaxPlatform, schema.ToolNone),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"mid": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["mid"].Tool != schema.ToolNone {
		t.Errorf("MID should have no tool after passing, got %q", next.Players["mid"].Tool)
	}
	if next.Players["top"].Tool != schema.ToolWrench {
		t.Errorf("TOP should hold wrench, got %q", next.Players["top"].Tool)
	}
}

func TestTick_WinWhenTopClimberReachesTopWithTool(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"c1": topClimber("c1", game.NumPlatforms-2, schema.ToolWrench),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"c1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Phase != schema.PhaseWon {
		t.Errorf("expected phase won, got %q", next.Phase)
	}
}

func TestTick_VictoryCondition(t *testing.T) {
	// Victory: TOP climber (index 1) at the top platform with any tool → PhaseWon.
	// Win check runs every tick regardless of input.
	state := playingState(map[string]*game.Player{
		"top": topClimber("top", game.NumPlatforms-1, schema.ToolWrench),
	})
	next := game.Tick(state, nil, 1.0/30.0)
	if next.Phase != schema.PhaseWon {
		t.Errorf("expected PhaseWon, got %q", next.Phase)
	}
}

func TestTick_VictoryRequiresTool(t *testing.T) {
	// TOP at summit without any tool must NOT win.
	state := playingState(map[string]*game.Player{
		"top": topClimber("top", game.NumPlatforms-1, schema.ToolNone),
	})
	next := game.Tick(state, nil, 1.0/30.0)
	if next.Phase == schema.PhaseWon {
		t.Error("TOP at summit without tool should not win")
	}
}

func TestTick_VictoryRequiresTopClimber(t *testing.T) {
	// MID climber at their max with tool must NOT win (MID can never reach summit).
	state := playingState(map[string]*game.Player{
		"mid": climber("mid", game.MidMaxPlatform, schema.ToolWrench),
	})
	next := game.Tick(state, nil, 1.0/30.0)
	if next.Phase == schema.PhaseWon {
		t.Error("MID climber with tool should not win — only TOP can reach the summit")
	}
}

func TestTick_NoWinWithoutTool(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"c1": topClimber("c1", game.NumPlatforms-2, schema.ToolNone),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"c1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Phase == schema.PhaseWon {
		t.Error("should not win without tool")
	}
}

func TestTick_MidClimberCannotWin(t *testing.T) {
	// MID climber is capped at MidMaxPlatform and cannot trigger win even with tool
	state := playingState(map[string]*game.Player{
		"c1": climber("c1", game.MidMaxPlatform, schema.ToolWrench),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"c1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Phase == schema.PhaseWon {
		t.Error("MID climber should not be able to win")
	}
}

func TestTick_TopPassesToolDownToMid(t *testing.T) {
	// TOP at MidMaxPlatform with tool passes DOWN to MID at same platform
	state := playingState(map[string]*game.Player{
		"mid": climber("mid", game.MidMaxPlatform, schema.ToolNone),
		"top": topClimber("top", game.MidMaxPlatform, schema.ToolHammer),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"top": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["top"].Tool != schema.ToolNone {
		t.Errorf("TOP should have no tool after passing down, got %q", next.Players["top"].Tool)
	}
	if next.Players["mid"].Tool != schema.ToolHammer {
		t.Errorf("MID should hold hammer after receiving from TOP, got %q", next.Players["mid"].Tool)
	}
}

func TestTick_TopCannotPassDownFromNonHandoffPlatform(t *testing.T) {
	// TOP at platform 4 (not MidMaxPlatform) cannot pass DOWN
	state := playingState(map[string]*game.Player{
		"top": topClimber("top", game.MidMaxPlatform+1, schema.ToolHammer),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"top": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["top"].Tool != schema.ToolHammer {
		t.Error("TOP should keep tool — can only pass down at MidMaxPlatform")
	}
}

func TestTick_MidPassesToolDownToBase(t *testing.T) {
	// MID at platform 0 with tool passes DOWN to BASE — tool added to BASE's inventory
	state := playingState(map[string]*game.Player{
		"base": baseOp("base"),
		"mid":  climber("mid", 0, schema.ToolWrench),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"mid": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["mid"].Tool != schema.ToolNone {
		t.Errorf("MID should have no tool after returning to base, got %q", next.Players["mid"].Tool)
	}
	if len(next.Players["base"].HeldTools) != 1 || next.Players["base"].HeldTools[0] != schema.ToolWrench {
		t.Errorf("BASE should have wrench in inventory, got %v", next.Players["base"].HeldTools)
	}
}

// TestTick_BaseCannotPassToMidWhoAlreadyHasTool verifies issue #19:
// BASE tries to pass a second tool to MID who already holds one.
// The pass should be silently rejected; MID keeps original tool, BASE keeps its tools.
func TestTick_BaseCannotPassToMidWhoAlreadyHasTool(t *testing.T) {
	base := baseOp("base", schema.ToolWrench, schema.ToolHammer)
	state := playingState(map[string]*game.Player{
		"base": base,
		"c1":   climber("c1", 0, schema.ToolHammer), // MID already holds a tool
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["c1"].Tool != schema.ToolHammer {
		t.Errorf("MID should still hold hammer after rejected pass, got %q", next.Players["c1"].Tool)
	}
	if len(next.Players["base"].HeldTools) != 2 {
		t.Errorf("BASE should still hold both tools after rejected pass, got %v", next.Players["base"].HeldTools)
	}
}

// TestTick_MidCannotPassToTopWhoAlreadyHasTool verifies the same guard for the MID→TOP path.
// MID tries to pass to TOP who already holds a tool; the pass should be silently rejected.
func TestTick_MidCannotPassToTopWhoAlreadyHasTool(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"mid": climber("mid", game.MidMaxPlatform, schema.ToolWrench),
		"top": topClimber("top", game.MidMaxPlatform, schema.ToolHammer), // TOP already holds a tool
	})
	next := game.Tick(state, map[string]schema.InputPayload{"mid": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["mid"].Tool != schema.ToolWrench {
		t.Errorf("MID should still hold wrench after rejected pass, got %q", next.Players["mid"].Tool)
	}
	if next.Players["top"].Tool != schema.ToolHammer {
		t.Errorf("TOP should still hold hammer after rejected pass, got %q", next.Players["top"].Tool)
	}
}

func TestTick_MidCannotPassDownFromNonGroundPlatform(t *testing.T) {
	// MID at platform 1 cannot pass DOWN
	state := playingState(map[string]*game.Player{
		"base": baseOp("base"),
		"mid":  climber("mid", 1, schema.ToolWrench),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"mid": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["mid"].Tool != schema.ToolWrench {
		t.Error("MID should keep tool — can only pass down at platform 0")
	}
}

func TestTick_InputsIgnoredInWaitingPhase(t *testing.T) {
	state := game.GameState{
		Tick: 0, Phase: schema.PhaseWaiting,
		Players: map[string]*game.Player{"c1": climber("c1", 0, schema.ToolNone)},
	}
	next := game.Tick(state, map[string]schema.InputPayload{"c1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Players["c1"].Platform != 0 {
		t.Error("inputs should be ignored in waiting phase")
	}
}

func TestTick_IncrementsTick(t *testing.T) {
	state := game.GameState{Tick: 5, Phase: schema.PhaseWaiting, Players: map[string]*game.Player{}}
	next := game.Tick(state, nil, 1.0/30.0)
	if next.Tick != 6 {
		t.Errorf("expected tick 6, got %d", next.Tick)
	}
}

// TestTick_NoToolDuplicateAfterRoundTrip verifies issue #24:
// BASE starts with [wrench, hammer] (SelectedIdx=1). BASE passes hammer to MID.
// MID returns hammer to BASE. BASE should end with [wrench, hammer] — not [hammer, hammer, wrench].
//
// The root cause: when MID returns a tool to BASE and BASE still holds a copy
// (possible when BASE's UP-pass was blocked by the guard in the same tick that
// MID presses SPACE), the DOWN-pass appends unconditionally, creating a duplicate.
// The fix guards the DOWN-pass append with a duplicate check.
func TestTick_NoToolDuplicateAfterRoundTrip(t *testing.T) {
	// Part 1 — sequential round-trip (2 ticks): basic correctness.
	base := baseOp("base", schema.ToolWrench, schema.ToolHammer)
	base.SelectedIdx = 1 // hammer selected
	state := playingState(map[string]*game.Player{
		"base": base,
		"mid":  climber("mid", 0, schema.ToolNone),
	})

	// Tick 1: BASE presses SPACE → passes hammer to MID.
	state1 := game.Tick(state, map[string]schema.InputPayload{
		"base": {Keys: schema.InputKeys{Space: true}},
	}, 1.0/30.0)

	if state1.Players["mid"].Tool != schema.ToolHammer {
		t.Fatalf("tick1: MID should hold hammer, got %q", state1.Players["mid"].Tool)
	}
	if len(state1.Players["base"].HeldTools) != 1 || state1.Players["base"].HeldTools[0] != schema.ToolWrench {
		t.Fatalf("tick1: BASE should hold only [wrench], got %v", state1.Players["base"].HeldTools)
	}

	// Tick 2: MID presses SPACE at platform 0 → returns hammer to BASE.
	state2 := game.Tick(state1, map[string]schema.InputPayload{
		"mid": {Keys: schema.InputKeys{Space: true}},
	}, 1.0/30.0)

	if state2.Players["mid"].Tool != schema.ToolNone {
		t.Errorf("tick2: MID should have no tool after returning hammer, got %q", state2.Players["mid"].Tool)
	}
	if len(state2.Players["base"].HeldTools) != 2 {
		t.Errorf("tick2: BASE should have exactly 2 tools, got %d: %v",
			len(state2.Players["base"].HeldTools), state2.Players["base"].HeldTools)
	}

	// Part 2 — duplicate-guard: MID holds hammer while BASE also holds [wrench, hammer].
	// Without the fix, DOWN-pass appends unconditionally → BASE = [wrench, hammer, hammer].
	base2 := baseOp("base", schema.ToolWrench, schema.ToolHammer)
	state3 := playingState(map[string]*game.Player{
		"base": base2,
		"mid":  climber("mid", 0, schema.ToolHammer), // MID already holds hammer
	})
	state4 := game.Tick(state3, map[string]schema.InputPayload{
		"mid": {Keys: schema.InputKeys{Space: true}},
	}, 1.0/30.0)

	if state4.Players["mid"].Tool != schema.ToolNone {
		t.Errorf("part2: MID should have no tool after returning, got %q", state4.Players["mid"].Tool)
	}
	if len(state4.Players["base"].HeldTools) != 2 {
		t.Errorf("part2: BASE should have exactly 2 tools after MID returns duplicate, got %d: %v",
			len(state4.Players["base"].HeldTools), state4.Players["base"].HeldTools)
	}
}
