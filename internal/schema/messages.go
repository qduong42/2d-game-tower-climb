package schema

import "encoding/json"

// MsgType identifies the message direction and content.
type MsgType string

const (
	MsgWelcome  MsgType = "welcome"
	MsgSnapshot MsgType = "snapshot"
	MsgEvent    MsgType = "event"
	MsgJoin     MsgType = "join"
	MsgInput    MsgType = "input"
)

// Envelope wraps every message sent over the wire.
type Envelope struct {
	Type    MsgType         `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// --- Server → Client ---

type WelcomePayload struct {
	YourID   string `json:"yourId"`
	RoomCode string `json:"roomCode"`
	TickRate int    `json:"tickRate"`
}

type PlayerState struct {
	ID    string  `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
	Name  string  `json:"name"`
}

type SnapshotPayload struct {
	Tick    uint64        `json:"tick"`
	Players []PlayerState `json:"players"`
}

type EventType string

const (
	EventJoin  EventType = "join"
	EventLeave EventType = "leave"
	EventError EventType = "error"
)

type EventPayload struct {
	EventType EventType `json:"eventType"`
	PlayerID  string    `json:"playerId,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// --- Client → Server ---

type JoinPayload struct {
	RoomCode string `json:"roomCode"`
	Name     string `json:"name"`
	Color    string `json:"color"`
}

type InputKeys struct {
	Up    bool `json:"up"`
	Down  bool `json:"down"`
	Left  bool `json:"left"`
	Right bool `json:"right"`
	Space bool `json:"space"`
}

type MouseState struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Click bool    `json:"click"`
}

type InputPayload struct {
	Tick  uint64      `json:"tick"`
	Keys  InputKeys   `json:"keys"`
	Mouse *MouseState `json:"mouse,omitempty"`
}
