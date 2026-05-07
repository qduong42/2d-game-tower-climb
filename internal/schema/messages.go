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

// Role identifies a player's function in the game.
type Role string

const (
	RoleBase    Role = "base"
	RoleClimber Role = "climber"
)

// ToolType identifies which physical tool a player is holding.
type ToolType string

const (
	ToolNone   ToolType = ""
	ToolWrench ToolType = "wrench"
	ToolHammer ToolType = "hammer"
)

// Phase describes the current game state.
type Phase string

const (
	PhaseWaiting Phase = "waiting" // fewer than MaxPlayers joined
	PhasePlaying Phase = "playing"
	PhaseWon     Phase = "won"
)

// --- Server → Client ---

type WelcomePayload struct {
	YourID   string `json:"yourId"`
	RoomCode string `json:"roomCode"`
	TickRate int    `json:"tickRate"`
	Color    string `json:"color"` // actual assigned color (may differ if requested color was taken)
}

type PlayerState struct {
	ID           string     `json:"id"`
	Color        string     `json:"color"`
	Name         string     `json:"name"`
	Role         Role       `json:"role"`
	ClimberIndex int        `json:"climberIndex"`  // 0 or 1 for climbers; -1 for base operator
	Platform     int        `json:"platform"`      // 0=ground … NumPlatforms-1=top
	Tool         ToolType   `json:"tool"`          // tool currently carried ("" = none)
	HeldTools    []ToolType `json:"heldTools"`     // BASE inventory (nil for climbers)
	SelectedTool ToolType   `json:"selectedTool"`  // BASE: which tool is queued to pass next
}

type SnapshotPayload struct {
	Tick         uint64        `json:"tick"`
	Phase        Phase         `json:"phase"`
	Players      []PlayerState `json:"players"`
	RequiredTool ToolType      `json:"requiredTool"`
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
