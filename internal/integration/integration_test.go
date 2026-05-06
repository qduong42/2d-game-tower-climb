// Package integration contains end-to-end tests that exercise the full
// HTTP → WebSocket → Join → Welcome → Snapshot pipeline.
// These are the tests that catch "player joins but canvas stays empty".
package integration_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qduong42/2d-game-tower-climb/internal/gateway"
	"github.com/qduong42/2d-game-tower-climb/internal/room"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
	"nhooyr.io/websocket"
)

// testServer starts a real HTTP test server with the gateway wired up.
func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	mgr := room.NewManager()
	gw := gateway.New(mgr)
	srv := httptest.NewServer(gw)
	t.Cleanup(srv.Close)
	return srv
}

// wsURL converts an httptest server URL to a ws:// URL.
func wsURL(srv *httptest.Server, roomCode string) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/r/" + roomCode
}

type wsClient struct {
	conn *websocket.Conn
	ctx  context.Context
}

// join connects to the server, sends a Join message, and returns the client.
func join(t *testing.T, srv *httptest.Server, roomCode, name, color string) *wsClient {
	t.Helper()
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, wsURL(srv, roomCode), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })

	env, _ := json.Marshal(schema.Envelope{
		Type:    schema.MsgJoin,
		Payload: mustMarshal(schema.JoinPayload{RoomCode: roomCode, Name: name, Color: color}),
	})
	if err := conn.Write(ctx, websocket.MessageText, env); err != nil {
		t.Fatalf("write join: %v", err)
	}
	return &wsClient{conn: conn, ctx: ctx}
}

// readEnvelope reads one message and decodes it.
func (c *wsClient) readEnvelope(t *testing.T) schema.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()
	_, data, err := c.conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var env schema.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return env
}

// readUntil reads messages until pred returns true, or times out.
func (c *wsClient) readUntil(t *testing.T, pred func(schema.Envelope) bool) schema.Envelope {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for expected message")
			return schema.Envelope{}
		default:
			env := c.readEnvelope(t)
			if pred(env) {
				return env
			}
		}
	}
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// --- Specification tests ---

func TestE2E_JoinReceivesWelcome(t *testing.T) {
	srv := testServer(t)
	c := join(t, srv, "ROOM1", "alice", "#e74c3c")

	env := c.readEnvelope(t)
	if env.Type != schema.MsgWelcome {
		t.Fatalf("expected welcome, got %q", env.Type)
	}
	var w schema.WelcomePayload
	json.Unmarshal(env.Payload, &w)

	if w.YourID == "" {
		t.Error("Welcome.yourId is empty")
	}
	if w.RoomCode != "ROOM1" {
		t.Errorf("Welcome.roomCode = %q, want ROOM1", w.RoomCode)
	}
	if w.TickRate != 30 {
		t.Errorf("Welcome.tickRate = %d, want 30", w.TickRate)
	}
}

func TestE2E_WelcomeFollowedBySnapshotContainingPlayer(t *testing.T) {
	srv := testServer(t)
	c := join(t, srv, "ROOM2", "alice", "#e74c3c")

	// Read Welcome, grab our ID.
	welcomeEnv := c.readEnvelope(t)
	var w schema.WelcomePayload
	json.Unmarshal(welcomeEnv.Payload, &w)
	myID := w.YourID

	// Read until we get a snapshot containing us.
	snapEnv := c.readUntil(t, func(e schema.Envelope) bool { return e.Type == schema.MsgSnapshot })
	var snap schema.SnapshotPayload
	json.Unmarshal(snapEnv.Payload, &snap)

	found := false
	for _, p := range snap.Players {
		if p.ID == myID {
			found = true
			if p.Name != "alice" {
				t.Errorf("player name = %q, want alice", p.Name)
			}
		}
	}
	if !found {
		t.Errorf("own player ID %q not found in snapshot (players: %v)", myID, snap.Players)
	}
}

func TestE2E_TwoPlayersSeeBothInSnapshot(t *testing.T) {
	srv := testServer(t)
	c1 := join(t, srv, "ROOM3", "alice", "#e74c3c")
	c2 := join(t, srv, "ROOM3", "bob", "#3498db")

	// Both should eventually receive a snapshot with 2 players.
	for _, tc := range []struct {
		name   string
		client *wsClient
	}{
		{"alice sees both", c1},
		{"bob sees both", c2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := tc.client.readUntil(t, func(e schema.Envelope) bool {
				if e.Type != schema.MsgSnapshot {
					return false
				}
				var snap schema.SnapshotPayload
				json.Unmarshal(e.Payload, &snap)
				return len(snap.Players) == 2
			})
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			if len(snap.Players) != 2 {
				t.Errorf("expected 2 players, got %d", len(snap.Players))
			}
		})
	}
}

func TestE2E_GameStartsWhenThreePlayersJoin(t *testing.T) {
	srv := testServer(t)
	c1 := join(t, srv, "ROOM4", "alice", "#e74c3c")
	c2 := join(t, srv, "ROOM4", "bob", "#3498db")
	c3 := join(t, srv, "ROOM4", "carol", "#2ecc71")

	// Discard welcome messages from each
	c1.readEnvelope(t)
	c2.readEnvelope(t)
	c3.readEnvelope(t)

	// All three clients should see a playing-phase snapshot with roles assigned
	for _, tc := range []struct {
		name   string
		client *wsClient
	}{
		{"alice sees playing", c1},
		{"bob sees playing", c2},
		{"carol sees playing", c3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := tc.client.readUntil(t, func(e schema.Envelope) bool {
				if e.Type != schema.MsgSnapshot {
					return false
				}
				var snap schema.SnapshotPayload
				json.Unmarshal(e.Payload, &snap)
				return snap.Phase == schema.PhasePlaying
			})
			var snap schema.SnapshotPayload
			json.Unmarshal(env.Payload, &snap)
			if len(snap.Players) != 3 {
				t.Errorf("expected 3 players, got %d", len(snap.Players))
			}
			// Exactly one base operator with tool
			bases := 0
			toolCount := 0
			for _, p := range snap.Players {
				if p.Role == schema.RoleBase {
					bases++
				}
				if p.HasTool {
					toolCount++
				}
			}
			if bases != 1 {
				t.Errorf("expected 1 base operator, got %d", bases)
			}
			if toolCount != 1 {
				t.Errorf("expected exactly 1 player with tool, got %d", toolCount)
			}
		})
	}
}
