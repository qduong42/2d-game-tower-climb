package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

func TestEnvelopeRoundtrip_Welcome(t *testing.T) {
	env := schema.Envelope{
		Type: schema.MsgWelcome,
		Payload: mustMarshal(t, schema.WelcomePayload{
			YourID:   "p1",
			RoomCode: "ABCD",
			TickRate: 20,
		}),
	}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var got schema.Envelope
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != schema.MsgWelcome {
		t.Errorf("type: got %q, want %q", got.Type, schema.MsgWelcome)
	}
	var w schema.WelcomePayload
	if err := json.Unmarshal(got.Payload, &w); err != nil {
		t.Fatal(err)
	}
	if w.YourID != "p1" || w.RoomCode != "ABCD" || w.TickRate != 20 {
		t.Errorf("unexpected payload: %+v", w)
	}
}

func TestEnvelopeRoundtrip_Snapshot(t *testing.T) {
	env := schema.Envelope{
		Type: schema.MsgSnapshot,
		Payload: mustMarshal(t, schema.SnapshotPayload{
			Tick: 42,
			Players: []schema.PlayerState{
				{ID: "p1", Color: "#ff0000", Name: "alice", Role: schema.RoleClimber, ClimberIndex: 0, Platform: 1, Tool: schema.ToolNone},
			},
		}),
	}
	data, _ := json.Marshal(env)
	var got schema.Envelope
	_ = json.Unmarshal(data, &got)
	var snap schema.SnapshotPayload
	if err := json.Unmarshal(got.Payload, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Tick != 42 || len(snap.Players) != 1 {
		t.Errorf("unexpected snapshot: %+v", snap)
	}
}

func TestInputPayload_MouseOptional(t *testing.T) {
	inp := schema.InputPayload{
		Tick: 1,
		Keys: schema.InputKeys{Up: true},
	}
	data, _ := json.Marshal(inp)
	var got schema.InputPayload
	_ = json.Unmarshal(data, &got)
	if got.Mouse != nil {
		t.Error("mouse should be nil when not set")
	}
	if !got.Keys.Up {
		t.Error("Keys.Up should be true")
	}
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
