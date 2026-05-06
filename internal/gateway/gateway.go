package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/qduong42/2d-game-tower-climb/internal/room"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
	"nhooyr.io/websocket"
)

// Gateway handles HTTP → WebSocket upgrades and routes players to rooms.
type Gateway struct {
	mgr *room.Manager
}

func New(mgr *room.Manager) *Gateway {
	return &Gateway{mgr: mgr}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, ok := ExtractRoomCode(r.URL.Path)
	if !ok {
		http.Error(w, "invalid room path — use /r/<CODE>", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins for dev; tighten for prod
	})
	if err != nil {
		slog.Warn("ws_accept_failed", "err", err)
		return
	}

	ctx := r.Context()
	client, rm := g.handshake(ctx, conn, code)
	if client == nil {
		conn.Close(websocket.StatusPolicyViolation, "bad join message")
		return
	}

	defer func() {
		rm.Leave(client.ID())
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	// writePump runs in this goroutine's background via room.
	go room.WritePump(ctx, client, conn)

	// readPump blocks here.
	g.readPump(ctx, client, rm)
}

func (g *Gateway) handshake(ctx context.Context, conn *websocket.Conn, code string) (*room.Client, *room.Room) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, nil
	}
	var env schema.Envelope
	if err := json.Unmarshal(data, &env); err != nil || env.Type != schema.MsgJoin {
		return nil, nil
	}
	var join schema.JoinPayload
	if err := json.Unmarshal(env.Payload, &join); err != nil {
		return nil, nil
	}

	rm := g.mgr.GetOrCreate(code)
	client := room.NewConnectedClient(join.Name, conn)
	if ok := rm.Join(client, join.Color); !ok {
		conn.Close(websocket.StatusPolicyViolation, "room is full")
		return nil, nil
	}

	welcome := schema.Envelope{
		Type: schema.MsgWelcome,
		Payload: mustMarshal(schema.WelcomePayload{
			YourID:   client.ID(),
			RoomCode: code,
			TickRate: 30,
			Color:    client.AssignedColor(),
		}),
	}
	data, _ = json.Marshal(welcome)
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		return nil, nil
	}

	return client, rm
}

func (g *Gateway) readPump(ctx context.Context, client *room.Client, rm *room.Room) {
	for {
		_, data, err := client.Conn().Read(ctx)
		if err != nil {
			return
		}
		var env schema.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		if env.Type != schema.MsgInput {
			continue
		}
		var inp schema.InputPayload
		if err := json.Unmarshal(env.Payload, &inp); err != nil {
			continue
		}
		rm.ReceiveInput(client.ID(), inp)
	}
}

// ExtractRoomCode parses "/r/<CODE>" and returns the code.
func ExtractRoomCode(path string) (string, bool) {
	trimmed := strings.TrimPrefix(path, "/r/")
	if trimmed == path || trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
