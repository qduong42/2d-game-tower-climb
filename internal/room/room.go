package room

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
	"nhooyr.io/websocket"
)

const tickRate = 30 // Hz

var colorPalette = []string{"#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6", "#1abc9c", "#e67e22", "#8e44ad"}

// pickColor returns the requested color if unused, otherwise the first available palette color.
func pickColor(requested string, state game.GameState) string {
	used := make(map[string]bool, len(state.Players))
	for _, p := range state.Players {
		used[p.Color] = true
	}
	if !used[requested] {
		return requested
	}
	for _, c := range colorPalette {
		if !used[c] {
			return c
		}
	}
	return requested // fallback: allow duplicate if palette exhausted
}

// Client represents one connected player.
type Client struct {
	id            string
	name          string
	conn          *websocket.Conn
	send          chan schema.Envelope
	assignedColor string // set by room after color deduplication
}

type joinReq struct {
	client *Client
	color  string
	done   chan bool // true = accepted, false = room full
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
		state:  game.GameState{Phase: schema.PhaseWaiting, Players: make(map[string]*game.Player)},
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

// Join queues a new client into the room and waits for color assignment.
// Returns false if the room is full.
func (r *Room) Join(c *Client, color string) bool {
	done := make(chan bool, 1)
	r.join <- joinReq{client: c, color: color, done: done}
	return <-done
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
			r.mu.Lock()
			if len(r.state.Players) >= game.MaxPlayers {
				r.mu.Unlock()
				req.done <- false
				continue
			}
			assignedColor := pickColor(req.color, r.state)
			r.state.Players[c.id] = &game.Player{
				ID: c.id, Color: assignedColor, Name: c.name,
			}
			c.assignedColor = assignedColor
			if len(r.state.Players) == game.MaxPlayers {
				r.assignRoles()
			}
			r.mu.Unlock()
			clients[c.id] = c
			req.done <- true
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
			if len(clients) == 0 {
				r.state = game.GameState{Phase: schema.PhaseWaiting, Players: make(map[string]*game.Player)}
				r.inputs = make(map[string]schema.InputPayload)
			}
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

// assignRoles randomly assigns roles when the third player joins.
// Must be called with r.mu held.
func (r *Room) assignRoles() {
	ids := make([]string, 0, len(r.state.Players))
	for id := range r.state.Players {
		ids = append(ids, id)
	}
	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })

	r.state.Players[ids[0]].Role = schema.RoleBase
	r.state.Players[ids[0]].ClimberIndex = -1
	r.state.Players[ids[0]].HeldTools = []schema.ToolType{schema.ToolWrench, schema.ToolHammer}
	r.state.Players[ids[0]].SelectedIdx = 0

	r.state.Players[ids[1]].Role = schema.RoleClimber
	r.state.Players[ids[1]].ClimberIndex = 0
	r.state.Players[ids[1]].Platform = game.MidMaxPlatform

	r.state.Players[ids[2]].Role = schema.RoleClimber
	r.state.Players[ids[2]].ClimberIndex = 1
	r.state.Players[ids[2]].Platform = game.NumPlatforms - 1

	// Randomly pick which tool is required to win.
	tools := []schema.ToolType{schema.ToolWrench, schema.ToolHammer}
	r.state.RequiredTool = tools[rand.IntN(len(tools))]

	r.state.Phase = schema.PhasePlaying
}

func buildSnapshot(s game.GameState) schema.SnapshotPayload {
	players := make([]schema.PlayerState, 0, len(s.Players))
	for _, p := range s.Players {
		var selectedTool schema.ToolType
		if p.Role == schema.RoleBase && len(p.HeldTools) > 0 {
			selectedTool = p.HeldTools[p.SelectedIdx]
		}
		players = append(players, schema.PlayerState{
			ID:           p.ID,
			Color:        p.Color,
			Name:         p.Name,
			Role:         p.Role,
			ClimberIndex: p.ClimberIndex,
			Platform:     p.Platform,
			Tool:         p.Tool,
			HeldTools:    p.HeldTools,
			SelectedTool: selectedTool,
		})
	}
	return schema.SnapshotPayload{Tick: s.Tick, Phase: s.Phase, Players: players, RequiredTool: s.RequiredTool}
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

const pingInterval = 30 * time.Second

// writePump drains c.send into the WebSocket connection and sends periodic
// pings to keep the connection alive through proxy idle timeouts.
func writePump(ctx context.Context, c *Client) {
	ping := time.NewTicker(pingInterval)
	defer ping.Stop()
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
		case <-ping.C:
			if err := c.conn.Ping(ctx); err != nil {
				return
			}
		}
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = cryptorand.Read(b)
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

// AssignedColor returns the color the room gave this player (available after Join is processed).
func (c *Client) AssignedColor() string { return c.assignedColor }

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
