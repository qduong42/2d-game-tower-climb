package room_test

// Specification tests: describe what the room SHOULD do, not how it does it.
// Each test maps to a user-visible requirement.

import (
	"encoding/json"
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

func TestSpec_PlayerStartsInsideWorldBounds(t *testing.T) {
	r := room.NewTestRoom("SPEC2")
	go r.RunForTest()

	c := room.NewTestClient("p1", "alice")
	r.Join(c, "#e74c3c")

	snap := drainUntilSnapshot(t, c.SendChan())
	p := findPlayer(t, snap, "p1")

	if p.X < 0 || p.X > game.WorldW || p.Y < 0 || p.Y > game.WorldH {
		t.Errorf("spawn position (%g,%g) outside world bounds", p.X, p.Y)
	}
}

func TestSpec_TwoPlayersSeeBothInSnapshot(t *testing.T) {
	r := room.NewTestRoom("SPEC3")
	go r.RunForTest()

	c1 := room.NewTestClient("p1", "alice")
	c2 := room.NewTestClient("p2", "bob")
	r.Join(c1, "#e74c3c")
	r.Join(c2, "#3498db")

	// Wait until c1 receives a snapshot with both players.
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

	// Wait for both to appear.
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

	// Now wait for a snapshot where p2 is gone.
	deadline2 := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c1.SendChan():
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			for _, p := range snap.Players {
				if p.ID == "p2" {
					continue // still there, keep waiting
				}
			}
			if len(snap.Players) == 1 {
				return // p2 gone
			}
		case <-deadline2:
			t.Fatal("p2 still in snapshot after leave")
		}
	}
}

func TestSpec_InputMovesPlayerPosition(t *testing.T) {
	r := room.NewTestRoom("SPEC5")
	go r.RunForTest()

	c := room.NewTestClient("p1", "alice")
	r.Join(c, "#e74c3c")

	// Record starting x from first snapshot.
	snap0 := drainUntilSnapshot(t, c.SendChan())
	startX := findPlayer(t, snap0, "p1").X

	// Send right-key input.
	r.ReceiveInput("p1", schema.InputPayload{
		Keys: schema.InputKeys{Right: true},
	})

	// Wait for a snapshot where x has increased.
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c.SendChan():
			if env.Type != schema.MsgSnapshot {
				continue
			}
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			p := findPlayer(t, snap, "p1")
			if p.X > startX {
				return // player moved right
			}
		case <-deadline:
			t.Fatalf("player x never increased (started at %g)", startX)
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

	// Wait for both players in a snapshot.
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
