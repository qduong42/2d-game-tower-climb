package game_test

import (
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

func stateWithPlayer(id string, x, y float64) game.GameState {
	return game.GameState{
		Tick: 0,
		Players: map[string]*game.Player{
			id: {ID: id, X: x, Y: y, Color: "#fff", Name: "test"},
		},
	}
}

func TestTick_MoveRight(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Right: true}},
	}
	next := game.Tick(state, inputs, 1.0/20.0)
	p := next.Players["p1"]
	if p.X <= 100 {
		t.Errorf("expected x > 100, got %f", p.X)
	}
	if p.Y != 100 {
		t.Errorf("expected y unchanged, got %f", p.Y)
	}
}

func TestTick_MoveUp(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Up: true}},
	}
	next := game.Tick(state, inputs, 1.0/20.0)
	p := next.Players["p1"]
	if p.Y >= 100 {
		t.Errorf("expected y < 100, got %f", p.Y)
	}
}

func TestTick_ClampsToBounds(t *testing.T) {
	state := stateWithPlayer("p1", 0, 0)
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Left: true, Up: true}},
	}
	next := game.Tick(state, inputs, 1.0)
	p := next.Players["p1"]
	if p.X < 0 || p.Y < 0 {
		t.Errorf("position went below 0: x=%f y=%f", p.X, p.Y)
	}
}

func TestTick_IncrementsTick(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	next := game.Tick(state, nil, 1.0/20.0)
	if next.Tick != 1 {
		t.Errorf("expected tick 1, got %d", next.Tick)
	}
}

func TestTick_NoInputNoMove(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	next := game.Tick(state, nil, 1.0/20.0)
	p := next.Players["p1"]
	if p.X != 100 || p.Y != 100 {
		t.Errorf("player moved without input: x=%f y=%f", p.X, p.Y)
	}
}

func TestTick_PreservesOtherPlayers(t *testing.T) {
	state := game.GameState{
		Tick: 0,
		Players: map[string]*game.Player{
			"p1": {ID: "p1", X: 100, Y: 100, Color: "#f00", Name: "a"},
			"p2": {ID: "p2", X: 200, Y: 200, Color: "#0f0", Name: "b"},
		},
	}
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Right: true}},
	}
	next := game.Tick(state, inputs, 1.0/20.0)
	if next.Players["p2"].X != 200 || next.Players["p2"].Y != 200 {
		t.Error("p2 moved when only p1 had input")
	}
}
