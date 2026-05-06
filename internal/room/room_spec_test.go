package room_test

// Specification tests: describe what the room SHOULD do, not how it does it.
// Each test maps to a user-visible requirement.

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/room"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

// drainUntilSnapshot discards events and returns the first snapshot payload.
func drainUntilSnapshot(t *testing.T, ch <-chan schema.Envelope) schema.SnapshotPayload {
	t.Helper()
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-ch:
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			if err := json.Unmarshal(env.Payload, &snap); err != nil {
				t.Fatalf("unmarshal snapshot: %v", err)
			}
			return snap
		case <-deadline:
			t.Fatal("timeout waiting for snapshot")
			return schema.SnapshotPayload{}
		}
	}
}

// drainUntilPlayingSnapshot waits for a snapshot with Phase == playing.
func drainUntilPlayingSnapshot(t *testing.T, ch <-chan schema.Envelope) schema.SnapshotPayload {
	t.Helper()
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-ch:
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			if err := json.Unmarshal(env.Payload, &snap); err != nil {
				t.Fatalf("unmarshal snapshot: %v", err)
			}
			if snap.Phase == schema.PhasePlaying {
				return snap
			}
		case <-deadline:
			t.Fatal("timeout waiting for playing snapshot")
			return schema.SnapshotPayload{}
		}
	}
}

// findPlayer returns the PlayerState for id, or fails the test.
func findPlayer(t *testing.T, snap schema.SnapshotPayload, id string) schema.PlayerState {
	t.Helper()
	for _, p := range snap.Players {
		if p.ID == id {
			return p
		}
	}
	t.Fatalf("player %q not found in snapshot (got %d players)", id, len(snap.Players))
	return schema.PlayerState{}
}

func TestSpec_SnapshotContainsJoinedPlayer(t *testing.T) {
	r := room.NewTestRoom("SPEC1")
	go r.RunForTest()

	c := room.NewTestClient("p1", "alice")
	r.Join(c, "#e74c3c")

	snap := drainUntilSnapshot(t, c.SendChan())
	p := findPlayer(t, snap, "p1")

	if p.Name != "alice" {
		t.Errorf("expected name alice, got %q", p.Name)
	}
	if p.Color != "#e74c3c" {
		t.Errorf("expected color #e74c3c, got %q", p.Color)
	}
}

func TestSpec_PlayerStartsAtGroundPlatform(t *testing.T) {
	r := room.NewTestRoom("SPEC2")
	go r.RunForTest()

	c := room.NewTestClient("p1", "alice")
	r.Join(c, "#e74c3c")

	snap := drainUntilSnapshot(t, c.SendChan())
	p := findPlayer(t, snap, "p1")

	if p.Platform != 0 {
		t.Errorf("expected platform 0, got %d", p.Platform)
	}
}

func TestSpec_TwoPlayersSeeBothInSnapshot(t *testing.T) {
	r := room.NewTestRoom("SPEC3")
	go r.RunForTest()

	c1 := room.NewTestClient("p1", "alice")
	c2 := room.NewTestClient("p2", "bob")
	r.Join(c1, "#e74c3c")
	r.Join(c2, "#3498db")

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c1.SendChan():
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			if len(snap.Players) == 2 {
				return // both players visible
			}
		case <-deadline:
			t.Fatal("timeout: second player never appeared in c1's snapshots")
		}
	}
}

func TestSpec_PlayerAbsentFromSnapshotAfterLeave(t *testing.T) {
	r := room.NewTestRoom("SPEC4")
	go r.RunForTest()

	c1 := room.NewTestClient("p1", "alice")
	c2 := room.NewTestClient("p2", "bob")
	r.Join(c1, "#e74c3c")
	r.Join(c2, "#3498db")

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c1.SendChan():
			if env.Type == schema.MsgSnapshot {
				var snap schema.SnapshotPayload
				json.Unmarshal(env.Payload, &snap)
				if len(snap.Players) == 2 {
					goto bothJoined
				}
			}
		case <-deadline:
			t.Fatal("both players never appeared")
		}
	}
bothJoined:
	r.Leave("p2")

	deadline2 := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c1.SendChan():
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			if len(snap.Players) == 1 {
				return // p2 gone
			}
		case <-deadline2:
			t.Fatal("p2 still in snapshot after leave")
		}
	}
}

func TestSpec_ClimberMovesUpOnUpInput(t *testing.T) {
	r := room.NewTestRoom("SPEC5")
	go r.RunForTest()

	c1 := room.NewTestClient("p1", "alice")
	c2 := room.NewTestClient("p2", "bob")
	c3 := room.NewTestClient("p3", "carol")
	r.Join(c1, "#e74c3c")
	r.Join(c2, "#3498db")
	r.Join(c3, "#2ecc71")

	// Find a climber ID from the first playing snapshot
	snap := drainUntilPlayingSnapshot(t, c1.SendChan())
	var climberID string
	for _, p := range snap.Players {
		if p.Role == schema.RoleClimber {
			climberID = p.ID
			break
		}
	}
	if climberID == "" {
		t.Fatal("no climber found in playing snapshot")
	}

	r.ReceiveInput(climberID, schema.InputPayload{Keys: schema.InputKeys{Up: true}})

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c1.SendChan():
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var s schema.SnapshotPayload
			json.Unmarshal(env.Payload, &s)
			for _, p := range s.Players {
				if p.ID == climberID && p.Platform > 0 {
					return
				}
			}
		case <-deadline:
			t.Fatalf("climber %q never moved up", climberID)
		}
	}
}

func TestSpec_ColorDeduplication(t *testing.T) {
	r := room.NewTestRoom("SPEC6")
	go r.RunForTest()

	c1 := room.NewTestClient("p1", "alice")
	c2 := room.NewTestClient("p2", "bob")
	r.Join(c1, "#e74c3c")
	r.Join(c2, "#e74c3c") // same color — server should assign a different one

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c1.SendChan():
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			if len(snap.Players) < 2 {
				continue
			}
			colors := map[string]string{}
			for _, p := range snap.Players {
				colors[p.ID] = p.Color
			}
			if colors["p1"] == colors["p2"] {
				t.Errorf("both players have same color %q", colors["p1"])
			}
			return
		case <-deadline:
			t.Fatal("two players never appeared in snapshot")
		}
	}
}

func TestSpec_RoomRejectsMoreThanMaxPlayers(t *testing.T) {
	r := room.NewTestRoom("SPEC7")
	go r.RunForTest()

	for i := 0; i < game.MaxPlayers; i++ {
		c := room.NewTestClient(fmt.Sprintf("p%d", i+1), fmt.Sprintf("player%d", i+1))
		if ok := r.Join(c, "#e74c3c"); !ok {
			t.Fatalf("join %d/%d rejected unexpectedly", i+1, game.MaxPlayers)
		}
	}

	extra := room.NewTestClient("p_extra", "extra")
	if ok := r.Join(extra, "#ffffff"); ok {
		t.Error("4th player should have been rejected")
	}
}
