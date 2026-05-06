package room

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
	"nhooyr.io/websocket"
)

const tickRate = 20 // Hz

// Client represents one connected player.
type Client struct {
	id   string
	name string
	conn *websocket.Conn
	send chan schema.Envelope
}

type joinReq struct {
	client *Client
	color  string
}

// Room manages one game session.
type Room struct {
	code   string
	mu     sync.Mutex
	state  game.GameState
	inputs map[string]schema.InputPayload
	join   chan joinReq
	leave  chan string
	input  chan inputMsg
	done   chan struct{}
	once   sync.Once
}

type inputMsg struct {
	playerID string
	payload  schema.InputPayload
}

func newRoom(code string) *Room {
	return &Room{
		code:   code,
		state:  game.GameState{Players: make(map[string]*game.Player)},
		inputs: make(map[string]schema.InputPayload),
		join:   make(chan joinReq, 8),
		leave:  make(chan string, 8),
		input:  make(chan inputMsg, 64),
		done:   make(chan struct{}),
	}
}

func (r *Room) stop() {
	r.once.Do(func() { close(r.done) })
}

// Join queues a new client into the room.
func (r *Room) Join(c *Client, color string) {
	r.join <- joinReq{client: c, color: color}
}

// Leave queues a client removal.
func (r *Room) Leave(playerID string) {
	r.leave <- playerID
}

// ReceiveInput queues an input message.
func (r *Room) ReceiveInput(playerID string, inp schema.InputPayload) {
	r.input <- inputMsg{playerID: playerID, payload: inp}
}

func (r *Room) run() {
	ticker := time.NewTicker(time.Second / tickRate)
	defer ticker.Stop()
	clients := make(map[string]*Client)

	for {
		select {
		case <-r.done:
			return

		case req := <-r.join:
			c := req.client
			clients[c.id] = c
			r.mu.Lock()
			r.state.Players[c.id] = &game.Player{
				ID: c.id, X: 400, Y: 300, Color: req.color, Name: c.name,
			}
			r.mu.Unlock()
			slog.Info("player_join", "room", r.code, "player", c.id, "name", c.name)
			r.broadcast(clients, schema.Envelope{
				Type:    schema.MsgEvent,
				Payload: mustMarshal(schema.EventPayload{EventType: schema.EventJoin, PlayerID: c.id}),
			})

		case id := <-r.leave:
			delete(clients, id)
			r.mu.Lock()
			delete(r.state.Players, id)
			delete(r.inputs, id)
			r.mu.Unlock()
			slog.Info("player_leave", "room", r.code, "player", id)
			r.broadcast(clients, schema.Envelope{
				Type:    schema.MsgEvent,
				Payload: mustMarshal(schema.EventPayload{EventType: schema.EventLeave, PlayerID: id}),
			})

		case msg := <-r.input:
			r.mu.Lock()
			r.inputs[msg.playerID] = msg.payload
			r.mu.Unlock()

		case <-ticker.C:
			start := time.Now()
			r.mu.Lock()
			r.state = game.Tick(r.state, r.inputs, 1.0/tickRate)
			snap := buildSnapshot(r.state)
			r.mu.Unlock()
			r.broadcast(clients, schema.Envelope{
				Type:    schema.MsgSnapshot,
				Payload: mustMarshal(snap),
			})
			if elapsed := time.Since(start); elapsed > 2*(time.Second/tickRate) {
				slog.Warn("tick_slow", "room", r.code, "elapsed_ms", elapsed.Milliseconds())
			}
		}
	}
}

func buildSnapshot(s game.GameState) schema.SnapshotPayload {
	players := make([]schema.PlayerState, 0, len(s.Players))
	for _, p := range s.Players {
		players = append(players, schema.PlayerState{
			ID: p.ID, X: p.X, Y: p.Y, Color: p.Color, Name: p.Name,
		})
	}
	return schema.SnapshotPayload{Tick: s.Tick, Players: players}
}

func (r *Room) broadcast(clients map[string]*Client, env schema.Envelope) {
	for _, c := range clients {
		select {
		case c.send <- env:
		default:
		}
	}
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// writePump drains c.send into the WebSocket connection.
func writePump(ctx context.Context, c *Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case env, ok := <-c.send:
			if !ok {
				return
			}
			data, _ := json.Marshal(env)
			if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}
		}
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// --- Test helpers — exported for *_test packages ---

// NewTestClient creates a Client with a nil websocket (for tests).
func NewTestClient(id, name string) *Client {
	return &Client{id: id, name: name, send: make(chan schema.Envelope, 16)}
}

// NewTestRoom creates a Room without starting it (call RunForTest to start).
func NewTestRoom(code string) *Room {
	return newRoom(code)
}

// RunForTest runs the room loop — call as a goroutine in tests.
func (r *Room) RunForTest() { r.run() }

// SendChan returns the client's send channel for test assertions.
func (c *Client) SendChan() <-chan schema.Envelope { return c.send }

// --- Exported for gateway package ---

// ID returns the player's unique ID.
func (c *Client) ID() string { return c.id }

// Conn returns the underlying WebSocket connection.
func (c *Client) Conn() *websocket.Conn { return c.conn }

// NewConnectedClient creates a Client with a real WebSocket connection.
func NewConnectedClient(name string, conn *websocket.Conn) *Client {
	return &Client{
		id:   generateID(),
		name: name,
		conn: conn,
		send: make(chan schema.Envelope, 32),
	}
}

// WritePump is exported so gateway can call it.
func WritePump(ctx context.Context, c *Client, conn *websocket.Conn) {
	writePump(ctx, c)
}
