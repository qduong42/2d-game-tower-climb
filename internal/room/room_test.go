package room_test

import (
	"testing"
	"time"

	"github.com/qduong42/2d-game-tower-climb/internal/room"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

// newTestClient creates a client with a buffered send channel for testing.
// It uses a nil websocket connection — only the send channel is inspected.
func newTestClient(id, name string) *room.Client {
	return room.NewTestClient(id, name)
}

func TestRoom_JoinSendsEvent(t *testing.T) {
	r := room.NewTestRoom("TEST")
	go r.RunForTest()

	c := newTestClient("p1", "alice")
	r.Join(c, "#ff0000")

	// Wait for the join event to arrive on c.send
	select {
	case env := <-c.SendChan():
		if env.Type != schema.MsgEvent {
			t.Errorf("expected event, got %q", env.Type)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for join event")
	}
}

func TestRoom_TickProducesSnapshot(t *testing.T) {
	r := room.NewTestRoom("SNAP")
	go r.RunForTest()

	c := newTestClient("p1", "alice")
	r.Join(c, "#00ff00")

	// Drain the join event
	select {
	case <-c.SendChan():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("no join event")
	}

	// Wait for at least one snapshot (tick fires at 20 Hz)
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case env := <-c.SendChan():
			if env.Type == schema.MsgSnapshot {
				return // success
			}
		case <-deadline:
			t.Fatal("timeout: no snapshot received")
		}
	}
}
