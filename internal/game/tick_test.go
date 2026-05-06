package game_test

import (
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

func playingState(players map[string]*game.Player) game.GameState {
	return game.GameState{Tick: 0, Phase: schema.PhasePlaying, Players: players}
}

func climber(id string, platform int, hasTool bool) *game.Player {
	return &game.Player{
		ID: id, Color: "#fff", Name: "test",
		Role: schema.RoleClimber, ClimberIndex: 0,
		Platform: platform, HasTool: hasTool,
	}
}

func baseOp(id string, hasTool bool) *game.Player {
	return &game.Player{
		ID: id, Color: "#fff", Name: "base",
		Role: schema.RoleBase, ClimberIndex: -1,
		Platform: 0, HasTool: hasTool,
	}
}

func TestTick_ClimberMovesUpOnUpKey(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 0, false)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != 1 {
		t.Errorf("expected platform 1, got %d", next.Players["p1"].Platform)
	}
}

func TestTick_ClimberMovesDownOnDownKey(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 2, false)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Down: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != 1 {
		t.Errorf("expected platform 1, got %d", next.Players["p1"].Platform)
	}
}

func TestTick_ClimberCannotGoBelowGround(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 0, false)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Down: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != 0 {
		t.Errorf("expected platform 0, got %d", next.Players["p1"].Platform)
	}
}

func TestTick_ClimberCannotGoAboveTop(t *testing.T) {
	top := game.NumPlatforms - 1
	state := playingState(map[string]*game.Player{"p1": climber("p1", top, false)})
	next := game.Tick(state, map[string]schema.InputPayload{"p1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Players["p1"].Platform != top {
		t.Errorf("expected platform %d, got %d", top, next.Players["p1"].Platform)
	}
}

func TestTick_MovementIsRisingEdgeOnly(t *testing.T) {
	state := playingState(map[string]*game.Player{"p1": climber("p1", 0, false)})
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

func TestTick_SpacePassesToolToPlayerOnSamePlatform(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"base": baseOp("base", true),
		"c1":   climber("c1", 0, false),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if next.Players["base"].HasTool {
		t.Error("base should no longer have tool after passing")
	}
	if !next.Players["c1"].HasTool {
		t.Error("c1 should have received the tool")
	}
}

func TestTick_SpaceDoesNotPassIfNoOneSamePlatform(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"base": baseOp("base", true),
		"c1":   climber("c1", 1, false), // different platform
	})
	next := game.Tick(state, map[string]schema.InputPayload{"base": {Keys: schema.InputKeys{Space: true}}}, 1.0/30.0)
	if !next.Players["base"].HasTool {
		t.Error("base should still have tool — no one at same platform")
	}
}

func TestTick_WinWhenClimberReachesTopWithTool(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"c1": climber("c1", game.NumPlatforms-2, true),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"c1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Phase != schema.PhaseWon {
		t.Errorf("expected phase won, got %q", next.Phase)
	}
}

func TestTick_NoWinWithoutTool(t *testing.T) {
	state := playingState(map[string]*game.Player{
		"c1": climber("c1", game.NumPlatforms-2, false),
	})
	next := game.Tick(state, map[string]schema.InputPayload{"c1": {Keys: schema.InputKeys{Up: true}}}, 1.0/30.0)
	if next.Phase == schema.PhaseWon {
		t.Error("should not win without tool")
	}
}

func TestTick_InputsIgnoredInWaitingPhase(t *testing.T) {
	state := game.GameState{
		Tick: 0, Phase: schema.PhaseWaiting,
		Players: map[string]*game.Player{"c1": climber("c1", 0, false)},
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
