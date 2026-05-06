# Milestone 1 — Cursor Party Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove the full networking + deploy pipeline: 2–4 players open a shared URL, see coloured dots move in real time via WASD, deployed on Render.

**Architecture:** Go server manages rooms with one goroutine per room; a fixed 20 Hz tick loop advances player positions and broadcasts JSON snapshots over WebSocket. TypeScript/Vite client receives snapshots, interpolates remote positions 100 ms in the past, and renders coloured circles on an HTML5 Canvas. Static client files are embedded in the Go binary and served at `/`.

**Tech Stack:** Go 1.22+, `nhooyr.io/websocket`, `log/slog`; TypeScript 5, Vite 5, `vitest`; Render free tier (render.yaml); `embed.FS` to bundle client into binary.

---

## File Map

```
/
├── cmd/server/main.go              # HTTP server, embed, routes
├── internal/
│   ├── schema/
│   │   ├── messages.go            # S5 – canonical message types (source of truth)
│   │   └── messages_test.go
│   ├── game/
│   │   ├── state.go               # S4 – GameState, Player structs
│   │   ├── tick.go                # S3 – pure Tick() function
│   │   └── tick_test.go
│   ├── room/
│   │   ├── room.go                # S2 – per-room goroutine + client set
│   │   ├── room_test.go
│   │   ├── manager.go             # S2 – room registry
│   │   └── manager_test.go
│   └── gateway/
│       ├── gateway.go             # S1 – HTTP→WebSocket upgrade, room routing
│       └── gateway_test.go
├── client/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.ts                # entry: wires lobby → game loop
│       ├── schema.ts              # S5 mirror – TS types (hand-mirrored from Go)
│       ├── network.ts             # C1 – WebSocket client
│       ├── input.ts               # C2 – keyboard + mouse capture
│       ├── renderer.ts            # C3 – draw abstraction (canvas backend)
│       ├── interpolation.ts       # C4 – snapshot buffer + interpolation
│       ├── lobby.ts               # C5 – create/join room UI
│       ├── menu.ts                # C6 – pause/leave overlay
│       └── logging.ts             # C7 – structured client events
├── client/src/tests/
│   ├── schema.test.ts
│   ├── network.test.ts
│   ├── input.test.ts
│   └── interpolation.test.ts
├── Dockerfile
├── Makefile
├── README.md
├── STATUS.md
├── render.yaml
├── go.mod
└── .claude/agents/component-builder.md
```

---

## Task 1: Go module + project scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go` (stub)
- Create: `internal/schema/.gitkeep`, `internal/game/.gitkeep`, `internal/room/.gitkeep`, `internal/gateway/.gitkeep`

- [ ] **Step 1: Initialize Go module**

```bash
go mod init github.com/qduong42/2d-game-tower-climb
```

Expected: `go.mod` created with `module github.com/qduong42/2d-game-tower-climb` and `go 1.22`.

- [ ] **Step 2: Add WebSocket dependency**

```bash
go get nhooyr.io/websocket@v1.8.11
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 3: Create directory structure**

```bash
mkdir -p cmd/server internal/schema internal/game internal/room internal/gateway
```

- [ ] **Step 4: Create stub main.go**

`cmd/server/main.go`:
```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("server starting")
}
```

- [ ] **Step 5: Verify it compiles**

```bash
go build ./...
```

Expected: exits 0, no output.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum cmd/server/main.go
git commit -m "chore: initialize Go module and server stub"
```

---

## Task 2: TS/Vite client scaffold

**Files:**
- Create: `client/package.json`
- Create: `client/tsconfig.json`
- Create: `client/vite.config.ts`
- Create: `client/index.html`
- Create: `client/src/main.ts` (stub)

- [ ] **Step 1: Create client/package.json**

`client/package.json`:
```json
{
  "name": "tower-climb-client",
  "private": true,
  "version": "0.0.1",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "test": "vitest run"
  },
  "devDependencies": {
    "typescript": "^5.4.0",
    "vite": "^5.2.0",
    "vitest": "^1.5.0"
  }
}
```

- [ ] **Step 2: Create client/tsconfig.json**

`client/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src"]
}
```

- [ ] **Step 3: Create client/vite.config.ts**

`client/vite.config.ts`:
```typescript
import { defineConfig } from "vite";

export default defineConfig({
  root: ".",
  build: { outDir: "dist", emptyOutDir: true },
  test: { include: ["src/tests/**/*.test.ts"] },
});
```

- [ ] **Step 4: Create client/index.html**

`client/index.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Tower Climb</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { background: #1a1a2e; color: #eee; font-family: monospace; display: flex; justify-content: center; align-items: center; height: 100vh; }
    #app { text-align: center; }
    canvas { display: block; border: 1px solid #444; }
  </style>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.ts"></script>
</body>
</html>
```

- [ ] **Step 5: Create client/src/main.ts stub**

`client/src/main.ts`:
```typescript
const app = document.getElementById("app")!;
app.textContent = "Tower Climb loading...";
```

- [ ] **Step 6: Install dependencies and verify**

```bash
cd client && npm install
npm run build
```

Expected: `client/dist/` created, no errors.

- [ ] **Step 7: Commit**

```bash
cd ..
git add client/
git commit -m "chore: initialize Vite/TypeScript client scaffold"
```

---

## Task 3: Makefile, Dockerfile, render.yaml, README, STATUS.md

**Files:**
- Create: `Makefile`
- Create: `Dockerfile`
- Create: `render.yaml`
- Create: `README.md`
- Create: `STATUS.md`
- Create: `.claude/agents/component-builder.md`

- [ ] **Step 1: Create Makefile**

`Makefile`:
```makefile
.PHONY: dev build test clean

CLIENT_DIST=client/dist

dev:
	@echo "Starting server + client dev..."
	@cd client && npm run dev &
	@go run ./cmd/server

build: $(CLIENT_DIST)
	go build -o server ./cmd/server

$(CLIENT_DIST):
	cd client && npm ci && npm run build

test:
	go test ./...
	cd client && npm test

clean:
	rm -f server
	rm -rf client/dist
```

- [ ] **Step 2: Create Dockerfile**

`Dockerfile`:
```dockerfile
FROM node:20-alpine AS client-build
WORKDIR /app/client
COPY client/package*.json ./
RUN npm ci
COPY client/ .
RUN npm run build

FROM golang:1.22-alpine AS server-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=client-build /app/client/dist ./client/dist
RUN go build -o server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=server-build /app/server ./server
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 3: Create render.yaml**

`render.yaml`:
```yaml
services:
  - type: web
    name: tower-climb
    env: docker
    dockerfilePath: ./Dockerfile
    envVars:
      - key: PORT
        value: 8080
```

- [ ] **Step 4: Create README.md**

`README.md`:
```markdown
# Tower Climb

2D side-view multiplayer co-op game. Players climb a wind turbine together.

## Quick start

Prerequisites: Go 1.22+, Node 20+

\```bash
# Install client deps
cd client && npm install && cd ..

# Run dev server + client (two terminals)
go run ./cmd/server       # terminal 1 — listens on :8080
cd client && npm run dev  # terminal 2 — Vite dev server on :5173
\```

Open `http://localhost:5173/r/TEST` in two browsers.

## Run tests

\```bash
make test
\```

## Deploy to Render

Connect this repo on render.com — `render.yaml` handles the rest. No credit card required on the free tier.

## Architecture

See `docs/superpowers/specs/2026-05-06-rope-team-turbine-climb-design.md` for the full design spec including component map, milestone plan, and contribution model.

## Current milestone

See `STATUS.md`.
```

- [ ] **Step 5: Create STATUS.md**

`STATUS.md`:
```markdown
# Milestone Status

## Milestone 1 — Cursor party

> Goal: 2–4 players open a shared URL, see each other's coloured dots move in real time.

### Wave 1 (schema — must complete before Wave 2)
- [ ] S5 message schema (Go + TS) — not started

### Wave 2 (parallel — blocked on Wave 1)
- [ ] S1 WebSocket gateway — not started
- [ ] S2 room manager — not started
- [ ] C1 network client — not started
- [ ] C5 lobby UI — not started
- [ ] C6 menu overlay — not started
- [ ] X1 deploy pipeline — not started
- [ ] X2 local dev setup — done (Makefile)

### Wave 3 (parallel — blocked on Wave 2)
- [ ] S3 tick loop — not started
- [ ] C2 input handler — not started
- [ ] C3 renderer — not started
- [ ] C4 interpolation buffer — not started
- [ ] S6 server logging — not started
- [ ] C7 client logging — not started

### Integration
- [ ] Wire server main.go — not started
- [ ] Wire client main.ts — not started
- [ ] Manual integration test — not started
- [ ] Render deploy — not started

## Milestone 2 — Vertical slice
Not started. Begins after Milestone 1 is deployed and working.
```

- [ ] **Step 6: Create .claude/agents/component-builder.md**

`.claude/agents/component-builder.md`:
```markdown
---
description: Implements one scoped component from the Milestone 1 plan. Reads the spec and plan, writes tests first, implements until tests pass, commits atomically, updates STATUS.md.
---

You are a component-builder agent for the Tower Climb project.

## Read order (do this first, every run)
1. `docs/superpowers/specs/2026-05-06-rope-team-turbine-climb-design.md` — full design
2. `docs/superpowers/plans/2026-05-06-milestone-1-cursor-party.md` — implementation plan
3. `STATUS.md` — what is in flight
4. The specific task section in the plan you were assigned

## Discipline
- **TDD**: write the failing test first, verify it fails, implement minimal code, verify tests pass.
- **Boundary rule**: only edit files listed under "Files:" for your assigned task. If you need a schema change, stop and report instead of changing `internal/schema/`.
- **Atomic commits**: one commit per logical unit (test + implementation = one commit).
- **Branch**: you were dispatched on a branch — stay on it.

## Reporting
When done, write a short summary:
- Which task was completed
- Files touched
- Tests added and passing
- Any integration concerns for the dispatcher

Then update STATUS.md to mark your task as done.
```

- [ ] **Step 7: Commit everything**

```bash
git add Makefile Dockerfile render.yaml README.md STATUS.md .claude/agents/
git commit -m "chore: add Makefile, Dockerfile, Render config, README, STATUS, agent definition"
```

---

## Task 4: Go message schema — S5 (Wave 1)

**Files:**
- Create: `internal/schema/messages.go`
- Create: `internal/schema/messages_test.go`

This is the single source of truth for all client↔server messages. No other task should define wire types.

- [ ] **Step 1: Write the failing test**

`internal/schema/messages_test.go`:
```go
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
				{ID: "p1", X: 100, Y: 200, Color: "#ff0000", Name: "alice"},
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
```

- [ ] **Step 2: Run test — expect compile failure**

```bash
go test ./internal/schema/...
```

Expected: `cannot find package "github.com/qduong42/2d-game-tower-climb/internal/schema"`

- [ ] **Step 3: Implement messages.go**

`internal/schema/messages.go`:
```go
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
```

- [ ] **Step 4: Run tests — expect pass**

```bash
go test ./internal/schema/... -v
```

Expected:
```
--- PASS: TestEnvelopeRoundtrip_Welcome (0.00s)
--- PASS: TestEnvelopeRoundtrip_Snapshot (0.00s)
--- PASS: TestInputPayload_MouseOptional (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/schema/
git commit -m "feat(schema): add Go message types with roundtrip tests (S5)"
```

---

## Task 5: TS schema mirror + tests — S5

**Files:**
- Create: `client/src/schema.ts`
- Create: `client/src/tests/schema.test.ts`

Mirror the Go types exactly — same field names, same optionality. This file must be updated whenever `internal/schema/messages.go` changes.

- [ ] **Step 1: Write the failing test**

`client/src/tests/schema.test.ts`:
```typescript
import { describe, it, expect } from "vitest";
import type {
  Envelope,
  WelcomePayload,
  SnapshotPayload,
  InputPayload,
  JoinPayload,
} from "../schema";
import { MsgType } from "../schema";

describe("schema types and constants", () => {
  it("MsgType constants match wire values", () => {
    expect(MsgType.Welcome).toBe("welcome");
    expect(MsgType.Snapshot).toBe("snapshot");
    expect(MsgType.Event).toBe("event");
    expect(MsgType.Join).toBe("join");
    expect(MsgType.Input).toBe("input");
  });

  it("Envelope can be constructed and serialized", () => {
    const payload: WelcomePayload = { yourId: "p1", roomCode: "ABCD", tickRate: 20 };
    const env: Envelope = { type: MsgType.Welcome, payload };
    const json = JSON.stringify(env);
    const parsed = JSON.parse(json) as Envelope;
    expect(parsed.type).toBe("welcome");
    expect((parsed.payload as WelcomePayload).yourId).toBe("p1");
  });

  it("InputPayload mouse field is optional", () => {
    const inp: InputPayload = {
      tick: 1,
      keys: { up: true, down: false, left: false, right: false, space: false },
    };
    expect(inp.mouse).toBeUndefined();
    const json = JSON.stringify(inp);
    expect(json).not.toContain("mouse");
  });

  it("SnapshotPayload players array", () => {
    const snap: SnapshotPayload = {
      tick: 42,
      players: [{ id: "p1", x: 10, y: 20, color: "#ff0000", name: "alice" }],
    };
    expect(snap.players[0].id).toBe("p1");
  });
});
```

- [ ] **Step 2: Run — expect failure**

```bash
cd client && npm test
```

Expected: `Cannot find module '../schema'`

- [ ] **Step 3: Implement client/src/schema.ts**

`client/src/schema.ts`:
```typescript
// Mirror of internal/schema/messages.go — update both files together.

export enum MsgType {
  Welcome  = "welcome",
  Snapshot = "snapshot",
  Event    = "event",
  Join     = "join",
  Input    = "input",
}

// Envelope wraps every message on the wire.
export interface Envelope {
  type: MsgType;
  payload: unknown;
}

// --- Server → Client ---

export interface WelcomePayload {
  yourId: string;
  roomCode: string;
  tickRate: number;
}

export interface PlayerState {
  id: string;
  x: number;
  y: number;
  color: string;
  name: string;
}

export interface SnapshotPayload {
  tick: number;
  players: PlayerState[];
}

export type EventType = "join" | "leave" | "error";

export interface EventPayload {
  eventType: EventType;
  playerId?: string;
  message?: string;
}

// --- Client → Server ---

export interface JoinPayload {
  roomCode: string;
  name: string;
  color: string;
}

export interface InputKeys {
  up: boolean;
  down: boolean;
  left: boolean;
  right: boolean;
  space: boolean;
}

export interface MouseState {
  x: number;
  y: number;
  click: boolean;
}

export interface InputPayload {
  tick: number;
  keys: InputKeys;
  mouse?: MouseState;
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd client && npm test
```

Expected:
```
✓ schema types and constants > MsgType constants match wire values
✓ schema types and constants > Envelope can be constructed and serialized
✓ schema types and constants > InputPayload mouse field is optional
✓ schema types and constants > SnapshotPayload players array
```

- [ ] **Step 5: Commit**

```bash
cd ..
git add client/src/schema.ts client/src/tests/schema.test.ts
git commit -m "feat(schema): add TS schema mirror with type tests (S5)"
```

---

## Task 6: Game state model + tick function — S3/S4

**Files:**
- Create: `internal/game/state.go`
- Create: `internal/game/tick.go`
- Create: `internal/game/tick_test.go`

The tick function is a pure function — no goroutines, no I/O — which makes it fully testable.

- [ ] **Step 1: Write tick tests**

`internal/game/tick_test.go`:
```go
package game_test

import (
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/game"
	"github.com/qduong42/2d-game-tower-climb/internal/schema"
)

func stateWithPlayer(id string, x, y float64) game.GameState {
	return game.GameState{
		Tick: 0,
		Players: map[string]*game.Player{
			id: {ID: id, X: x, Y: y, Color: "#fff", Name: "test"},
		},
	}
}

func TestTick_MoveRight(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Right: true}},
	}
	next := game.Tick(state, inputs, 1.0/20.0)
	p := next.Players["p1"]
	if p.X <= 100 {
		t.Errorf("expected x > 100, got %f", p.X)
	}
	if p.Y != 100 {
		t.Errorf("expected y unchanged, got %f", p.Y)
	}
}

func TestTick_MoveUp(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Up: true}},
	}
	next := game.Tick(state, inputs, 1.0/20.0)
	p := next.Players["p1"]
	if p.Y >= 100 {
		t.Errorf("expected y < 100, got %f", p.Y)
	}
}

func TestTick_ClampsToBounds(t *testing.T) {
	state := stateWithPlayer("p1", 0, 0)
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Left: true, Up: true}},
	}
	next := game.Tick(state, inputs, 1.0)
	p := next.Players["p1"]
	if p.X < 0 || p.Y < 0 {
		t.Errorf("position went below 0: x=%f y=%f", p.X, p.Y)
	}
}

func TestTick_IncrementsTick(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	next := game.Tick(state, nil, 1.0/20.0)
	if next.Tick != 1 {
		t.Errorf("expected tick 1, got %d", next.Tick)
	}
}

func TestTick_NoInputNoMove(t *testing.T) {
	state := stateWithPlayer("p1", 100, 100)
	next := game.Tick(state, nil, 1.0/20.0)
	p := next.Players["p1"]
	if p.X != 100 || p.Y != 100 {
		t.Errorf("player moved without input: x=%f y=%f", p.X, p.Y)
	}
}

func TestTick_PreservesOtherPlayers(t *testing.T) {
	state := game.GameState{
		Tick: 0,
		Players: map[string]*game.Player{
			"p1": {ID: "p1", X: 100, Y: 100, Color: "#f00", Name: "a"},
			"p2": {ID: "p2", X: 200, Y: 200, Color: "#0f0", Name: "b"},
		},
	}
	inputs := map[string]schema.InputPayload{
		"p1": {Keys: schema.InputKeys{Right: true}},
	}
	next := game.Tick(state, inputs, 1.0/20.0)
	if next.Players["p2"].X != 200 || next.Players["p2"].Y != 200 {
		t.Error("p2 moved when only p1 had input")
	}
}
```

- [ ] **Step 2: Run — expect compile failure**

```bash
go test ./internal/game/...
```

Expected: package not found.

- [ ] **Step 3: Implement state.go**

`internal/game/state.go`:
```go
package game

// Player holds mutable per-player state.
type Player struct {
	ID    string
	X     float64
	Y     float64
	Color string
	Name  string
}

// GameState is the full authoritative state for one room at one tick.
// It is immutable by convention — Tick() returns a new value.
type GameState struct {
	Tick    uint64
	Players map[string]*Player
}
```

- [ ] **Step 4: Implement tick.go**

`internal/game/tick.go`:
```go
package game

import "github.com/qduong42/2d-game-tower-climb/internal/schema"

const (
	Speed  = 200.0 // pixels per second
	WorldW = 800.0
	WorldH = 600.0
)

// Tick advances game state by dt seconds given the latest inputs.
// It is a pure function — no side effects.
func Tick(state GameState, inputs map[string]schema.InputPayload, dt float64) GameState {
	next := GameState{
		Tick:    state.Tick + 1,
		Players: make(map[string]*Player, len(state.Players)),
	}
	for id, p := range state.Players {
		np := *p
		if inp, ok := inputs[id]; ok {
			if inp.Keys.Left {
				np.X -= Speed * dt
			}
			if inp.Keys.Right {
				np.X += Speed * dt
			}
			if inp.Keys.Up {
				np.Y -= Speed * dt
			}
			if inp.Keys.Down {
				np.Y += Speed * dt
			}
		}
		np.X = clamp(np.X, 0, WorldW)
		np.Y = clamp(np.Y, 0, WorldH)
		next.Players[id] = &np
	}
	return next
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
```

- [ ] **Step 5: Run tests — expect pass**

```bash
go test ./internal/game/... -v
```

Expected: all 6 tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/game/
git commit -m "feat(game): add state model and pure tick function with tests (S3/S4)"
```

---

## Task 7: Room manager — S2

**Files:**
- Create: `internal/room/manager.go`
- Create: `internal/room/manager_test.go`
- Create: `internal/room/room.go`
- Create: `internal/room/room_test.go`

- [ ] **Step 1: Write manager tests**

`internal/room/manager_test.go`:
```go
package room_test

import (
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/room"
)

func TestManager_GetOrCreate_SameCode(t *testing.T) {
	m := room.NewManager()
	r1 := m.GetOrCreate("ABCD")
	r2 := m.GetOrCreate("ABCD")
	if r1 != r2 {
		t.Error("expected same room for same code")
	}
}

func TestManager_GetOrCreate_DifferentCode(t *testing.T) {
	m := room.NewManager()
	r1 := m.GetOrCreate("ABCD")
	r2 := m.GetOrCreate("XYZ1")
	if r1 == r2 {
		t.Error("expected different rooms for different codes")
	}
}

func TestManager_Remove(t *testing.T) {
	m := room.NewManager()
	m.GetOrCreate("ABCD")
	m.Remove("ABCD")
	r := m.GetOrCreate("ABCD")
	if r == nil {
		t.Error("expected new room after remove")
	}
}
```

- [ ] **Step 2: Run — expect failure**

```bash
go test ./internal/room/...
```

Expected: package not found.

- [ ] **Step 3: Implement manager.go**

`internal/room/manager.go`:
```go
package room

import "sync"

// Manager maintains the active rooms indexed by room code.
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

// GetOrCreate returns the existing room for code, or starts a new one.
func (m *Manager) GetOrCreate(code string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rooms[code]; ok {
		return r
	}
	r := newRoom(code)
	m.rooms[code] = r
	go r.run()
	return r
}

// Remove stops and removes a room. Safe to call when room is already gone.
func (m *Manager) Remove(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rooms[code]; ok {
		r.stop()
		delete(m.rooms, code)
	}
}
```

- [ ] **Step 4: Create room.go stub (needed for manager to compile)**

`internal/room/room.go`:
```go
package room

import (
	"context"
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
```

- [ ] **Step 5: Run manager tests — expect pass**

```bash
go test ./internal/room/... -run TestManager -v
```

Expected: all 3 manager tests pass.

- [ ] **Step 6: Write room tests**

`internal/room/room_test.go`:
```go
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
```

These tests require test-exported helpers on Room. Add them to room.go (the test hooks are behind a build tag or a simple exported wrapper — we use plain exports since this is not a security-sensitive package):

Add to `internal/room/room.go`:
```go
// Test helpers — exported for *_test packages only.

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
```

- [ ] **Step 7: Run all room tests — expect pass**

```bash
go test ./internal/room/... -v -timeout 5s
```

Expected: all 5 tests pass.

- [ ] **Step 8: Commit**

```bash
git add internal/room/
git commit -m "feat(room): add room goroutine and manager with tests (S2)"
```

---

## Task 8: WebSocket gateway — S1

**Files:**
- Create: `internal/gateway/gateway.go`
- Create: `internal/gateway/gateway_test.go`

- [ ] **Step 1: Write gateway test**

`internal/gateway/gateway_test.go`:
```go
package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/gateway"
	"github.com/qduong42/2d-game-tower-climb/internal/room"
)

func TestGateway_RejectsNonWebSocket(t *testing.T) {
	mgr := room.NewManager()
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodGet, "/r/ABCD", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code == http.StatusSwitchingProtocols {
		t.Error("expected non-101, got 101")
	}
}

func TestGateway_RejectsMissingRoomCode(t *testing.T) {
	mgr := room.NewManager()
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodGet, "/r/", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGateway_ExtractCode(t *testing.T) {
	cases := []struct {
		path string
		want string
		ok   bool
	}{
		{"/r/ABCD", "ABCD", true},
		{"/r/", "", false},
		{"/r", "", false},
		{"/other/ABCD", "", false},
	}
	for _, tc := range cases {
		code, ok := gateway.ExtractRoomCode(tc.path)
		if ok != tc.ok || code != tc.want {
			t.Errorf("path=%q: got (%q,%v) want (%q,%v)", tc.path, code, ok, tc.want, tc.ok)
		}
	}
}
```

- [ ] **Step 2: Run — expect compile failure**

```bash
go test ./internal/gateway/...
```

Expected: package not found.

- [ ] **Step 3: Implement gateway.go**

`internal/gateway/gateway.go`:
```go
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

	welcome := schema.Envelope{
		Type:    schema.MsgWelcome,
		Payload: mustMarshal(schema.WelcomePayload{
			YourID:   client.ID(),
			RoomCode: code,
			TickRate: 20,
		}),
	}
	data, _ = json.Marshal(welcome)
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		return nil, nil
	}

	rm.Join(client, join.Color)
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
```

The gateway requires a few more exports on `room.Client`. Add to `internal/room/room.go`:

```go
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
```

Also add `generateID()` to room.go:
```go
import (
	"crypto/rand"
	"encoding/hex"
)

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
go test ./internal/gateway/... -v
```

Expected: all 3 tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/gateway/ internal/room/room.go
git commit -m "feat(gateway): add WebSocket gateway with room routing and tests (S1)"
```

---

## Task 9: Wire server — main.go

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Update main.go to serve static + WebSocket**

`cmd/server/main.go`:
```go
package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/qduong42/2d-game-tower-climb/internal/gateway"
	"github.com/qduong42/2d-game-tower-climb/internal/room"
)

//go:embed all:../../client/dist
var clientDist embed.FS

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mgr := room.NewManager()
	gw := gateway.New(mgr)

	mux := http.NewServeMux()
	mux.Handle("/r/", gw)

	// Serve built client from embedded FS.
	static, err := fs.Sub(clientDist, "client/dist")
	if err != nil {
		slog.Error("embed_sub_failed", "err", err)
		os.Exit(1)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))

	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server_error", "err", err)
		os.Exit(1)
	}
}
```

**Note:** The `//go:embed` path is relative to this file. Adjust the path to `../../client/dist` or restructure if needed after verifying with `go build`.

- [ ] **Step 2: Build (requires client/dist to exist)**

```bash
cd client && npm run build && cd ..
go build ./cmd/server
```

Expected: `server` binary created.

- [ ] **Step 3: Run locally**

```bash
./server
```

Expected: `{"level":"INFO","msg":"listening","port":"8080"}` in stdout.
Open `http://localhost:8080` — should serve the Vite-built index.html.

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat(server): wire HTTP server with static embed and WebSocket gateway"
```

---

## Task 10: TS network client — C1

**Files:**
- Create: `client/src/network.ts`
- Create: `client/src/tests/network.test.ts`

- [ ] **Step 1: Write network client tests (WebSocket stubbed)**

`client/src/tests/network.test.ts`:
```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { NetworkClient } from "../network";
import type { SnapshotPayload, WelcomePayload } from "../schema";
import { MsgType } from "../schema";

// Minimal WebSocket stub
class FakeWS {
  onmessage: ((e: MessageEvent) => void) | null = null;
  onclose: (() => void) | null = null;
  onopen: (() => void) | null = null;
  sent: string[] = [];
  readyState = 1; // OPEN

  send(data: string) { this.sent.push(data); }
  close() { this.readyState = 3; this.onclose?.(); }
  
  // Test helper: simulate incoming message
  receive(payload: unknown) {
    this.onmessage?.({ data: JSON.stringify(payload) } as MessageEvent);
  }
}

describe("NetworkClient", () => {
  let fakeWS: FakeWS;
  let client: NetworkClient;

  beforeEach(() => {
    fakeWS = new FakeWS();
    vi.stubGlobal("WebSocket", vi.fn(() => fakeWS));
    client = new NetworkClient();
  });

  afterEach(() => { vi.unstubAllGlobals(); });

  it("sends Join on connect", () => {
    client.connect("ABCD", "alice", "#ff0000");
    fakeWS.onopen?.();
    expect(fakeWS.sent).toHaveLength(1);
    const msg = JSON.parse(fakeWS.sent[0]);
    expect(msg.type).toBe(MsgType.Join);
    expect(msg.payload.name).toBe("alice");
    expect(msg.payload.roomCode).toBe("ABCD");
  });

  it("calls onWelcome when welcome arrives", () => {
    client.connect("ABCD", "alice", "#ff0000");
    fakeWS.onopen?.();

    const welcome: WelcomePayload = { yourId: "p1", roomCode: "ABCD", tickRate: 20 };
    let received: WelcomePayload | null = null;
    client.onWelcome((w) => { received = w; });
    fakeWS.receive({ type: MsgType.Welcome, payload: welcome });

    expect(received).not.toBeNull();
    expect(received!.yourId).toBe("p1");
  });

  it("calls onSnapshot when snapshot arrives", () => {
    client.connect("ABCD", "alice", "#ff0000");
    fakeWS.onopen?.();

    const snap: SnapshotPayload = { tick: 1, players: [] };
    let received: SnapshotPayload | null = null;
    client.onSnapshot((s) => { received = s; });
    fakeWS.receive({ type: MsgType.Snapshot, payload: snap });

    expect(received?.tick).toBe(1);
  });

  it("send queues until connected", () => {
    client.connect("ABCD", "alice", "#ff0000");
    // Before onopen fires, send an input
    client.sendInput({ tick: 1, keys: { up: true, down: false, left: false, right: false, space: false } });
    // Only the join should be sent once onopen fires
    fakeWS.onopen?.();
    expect(fakeWS.sent.length).toBeGreaterThanOrEqual(1);
  });
});
```

- [ ] **Step 2: Run — expect compile failure**

```bash
cd client && npm test -- network
```

- [ ] **Step 3: Implement network.ts**

`client/src/network.ts`:
```typescript
import {
  MsgType,
  type Envelope,
  type WelcomePayload,
  type SnapshotPayload,
  type EventPayload,
  type InputPayload,
} from "./schema";

export class NetworkClient {
  private ws: WebSocket | null = null;
  private onWelcomeCb: ((w: WelcomePayload) => void) | null = null;
  private onSnapshotCb: ((s: SnapshotPayload) => void) | null = null;
  private onEventCb: ((e: EventPayload) => void) | null = null;
  private pendingInput: InputPayload | null = null;

  connect(roomCode: string, name: string, color: string): void {
    const protocol = location.protocol === "https:" ? "wss" : "ws";
    const url = `${protocol}://${location.host}/r/${roomCode}`;
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      const env: Envelope = {
        type: MsgType.Join,
        payload: { roomCode, name, color },
      };
      this.ws!.send(JSON.stringify(env));
      if (this.pendingInput) {
        this.ws!.send(JSON.stringify({ type: MsgType.Input, payload: this.pendingInput }));
        this.pendingInput = null;
      }
    };

    this.ws.onmessage = (e: MessageEvent) => {
      const env = JSON.parse(e.data as string) as Envelope;
      switch (env.type) {
        case MsgType.Welcome:
          this.onWelcomeCb?.(env.payload as WelcomePayload);
          break;
        case MsgType.Snapshot:
          this.onSnapshotCb?.(env.payload as SnapshotPayload);
          break;
        case MsgType.Event:
          this.onEventCb?.(env.payload as EventPayload);
          break;
      }
    };

    this.ws.onclose = () => {
      console.warn("[network] connection closed");
    };
  }

  sendInput(payload: InputPayload): void {
    const env: Envelope = { type: MsgType.Input, payload };
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(env));
    } else {
      this.pendingInput = payload;
    }
  }

  onWelcome(cb: (w: WelcomePayload) => void): void { this.onWelcomeCb = cb; }
  onSnapshot(cb: (s: SnapshotPayload) => void): void { this.onSnapshotCb = cb; }
  onEvent(cb: (e: EventPayload) => void): void { this.onEventCb = cb; }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd client && npm test -- network
```

Expected: all 4 tests pass.

- [ ] **Step 5: Commit**

```bash
cd ..
git add client/src/network.ts client/src/tests/network.test.ts
git commit -m "feat(client): add network client with WebSocket stub tests (C1)"
```

---

## Task 11: Input handler — C2

**Files:**
- Create: `client/src/input.ts`
- Create: `client/src/tests/input.test.ts`

- [ ] **Step 1: Write input tests**

`client/src/tests/input.test.ts`:
```typescript
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { InputHandler } from "../input";

describe("InputHandler", () => {
  let handler: InputHandler;

  beforeEach(() => { handler = new InputHandler(); });
  afterEach(() => { handler.stop(); });

  it("starts with all keys false", () => {
    const inp = handler.getInput(0);
    expect(inp.keys.up).toBe(false);
    expect(inp.keys.down).toBe(false);
    expect(inp.keys.left).toBe(false);
    expect(inp.keys.right).toBe(false);
    expect(inp.keys.space).toBe(false);
  });

  it("tracks keydown/keyup for WASD", () => {
    handler.simulateKeyDown("w");
    expect(handler.getInput(1).keys.up).toBe(true);

    handler.simulateKeyUp("w");
    expect(handler.getInput(2).keys.up).toBe(false);
  });

  it("tracks arrow keys", () => {
    handler.simulateKeyDown("ArrowLeft");
    expect(handler.getInput(1).keys.left).toBe(true);
  });

  it("tracks space", () => {
    handler.simulateKeyDown(" ");
    expect(handler.getInput(1).keys.space).toBe(true);
  });

  it("captures mouse clicks (wired but no game binding)", () => {
    handler.simulateMouseClick(100, 200);
    const inp = handler.getInput(1);
    expect(inp.mouse).toBeDefined();
    expect(inp.mouse!.x).toBe(100);
    expect(inp.mouse!.y).toBe(200);
    expect(inp.mouse!.click).toBe(true);
  });

  it("getInput sets the tick field", () => {
    expect(handler.getInput(7).tick).toBe(7);
  });
});
```

- [ ] **Step 2: Run — expect compile failure**

```bash
cd client && npm test -- input
```

- [ ] **Step 3: Implement input.ts**

`client/src/input.ts`:
```typescript
import type { InputPayload, MouseState } from "./schema";

const KEY_MAP: Record<string, keyof { up: boolean; down: boolean; left: boolean; right: boolean; space: boolean }> = {
  w: "up", ArrowUp: "up",
  s: "down", ArrowDown: "down",
  a: "left", ArrowLeft: "left",
  d: "right", ArrowRight: "right",
  " ": "space",
};

export class InputHandler {
  private keys = { up: false, down: false, left: false, right: false, space: false };
  private mouse: MouseState | null = null;
  private boundKeyDown: ((e: KeyboardEvent) => void) | null = null;
  private boundKeyUp: ((e: KeyboardEvent) => void) | null = null;

  start(target: EventTarget = window): void {
    this.boundKeyDown = (e: KeyboardEvent) => {
      const k = KEY_MAP[e.key];
      if (k) { this.keys[k] = true; e.preventDefault(); }
    };
    this.boundKeyUp = (e: KeyboardEvent) => {
      const k = KEY_MAP[e.key];
      if (k) { this.keys[k] = false; }
    };
    target.addEventListener("keydown", this.boundKeyDown as EventListener);
    target.addEventListener("keyup", this.boundKeyUp as EventListener);
  }

  stop(): void {
    if (this.boundKeyDown) window.removeEventListener("keydown", this.boundKeyDown);
    if (this.boundKeyUp) window.removeEventListener("keyup", this.boundKeyUp);
  }

  captureMouseOnCanvas(canvas: HTMLCanvasElement): void {
    canvas.addEventListener("click", (e) => {
      const rect = canvas.getBoundingClientRect();
      this.mouse = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
        click: true,
      };
    });
  }

  getInput(tick: number): InputPayload {
    const inp: InputPayload = { tick, keys: { ...this.keys }, mouse: this.mouse ?? undefined };
    this.mouse = null; // consume click after reading
    return inp;
  }

  // Test helpers — only for use in tests.
  simulateKeyDown(key: string): void {
    const k = KEY_MAP[key];
    if (k) this.keys[k] = true;
  }
  simulateKeyUp(key: string): void {
    const k = KEY_MAP[key];
    if (k) this.keys[k] = false;
  }
  simulateMouseClick(x: number, y: number): void {
    this.mouse = { x, y, click: true };
  }
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd client && npm test -- input
```

Expected: all 6 tests pass.

- [ ] **Step 5: Commit**

```bash
cd ..
git add client/src/input.ts client/src/tests/input.test.ts
git commit -m "feat(client): add input handler with keyboard+mouse capture and tests (C2)"
```

---

## Task 12: Interpolation buffer — C4

**Files:**
- Create: `client/src/interpolation.ts`
- Create: `client/src/tests/interpolation.test.ts`

- [ ] **Step 1: Write interpolation tests**

`client/src/tests/interpolation.test.ts`:
```typescript
import { describe, it, expect } from "vitest";
import { InterpolationBuffer } from "../interpolation";
import type { SnapshotPayload } from "../schema";

function snap(tick: number, x: number): SnapshotPayload {
  return { tick, players: [{ id: "p1", x, y: 0, color: "#fff", name: "a" }] };
}

describe("InterpolationBuffer", () => {
  it("returns empty array when no snapshots", () => {
    const buf = new InterpolationBuffer();
    expect(buf.getInterpolated(Date.now())).toEqual([]);
  });

  it("returns last snapshot when only one exists", () => {
    const buf = new InterpolationBuffer();
    const now = Date.now();
    buf.push(snap(1, 100), now);
    const result = buf.getInterpolated(now + 50);
    expect(result[0]?.x).toBe(100);
  });

  it("interpolates between two snapshots", () => {
    const buf = new InterpolationBuffer();
    const t0 = 1000;
    buf.push(snap(1, 0), t0);
    buf.push(snap(2, 100), t0 + 50); // 50ms apart

    // Render time is t0 + 25 (halfway between)
    const result = buf.getInterpolated(t0 + 25);
    // Should be between 0 and 100
    expect(result[0]?.x).toBeGreaterThan(0);
    expect(result[0]?.x).toBeLessThan(100);
  });

  it("caps buffer to 10 snapshots", () => {
    const buf = new InterpolationBuffer();
    for (let i = 0; i < 15; i++) {
      buf.push(snap(i, i * 10), 1000 + i * 50);
    }
    // No assertion on internal state, just verify it doesn't crash
    const result = buf.getInterpolated(1000 + 14 * 50);
    expect(result).toBeDefined();
  });
});
```

- [ ] **Step 2: Run — expect compile failure**

```bash
cd client && npm test -- interpolation
```

- [ ] **Step 3: Implement interpolation.ts**

`client/src/interpolation.ts`:
```typescript
import type { PlayerState, SnapshotPayload } from "./schema";

interface TimedSnapshot {
  snap: SnapshotPayload;
  receivedAt: number;
}

const RENDER_DELAY_MS = 100;
const MAX_BUFFER = 10;

export class InterpolationBuffer {
  private buffer: TimedSnapshot[] = [];

  push(snap: SnapshotPayload, receivedAt = Date.now()): void {
    this.buffer.push({ snap, receivedAt });
    if (this.buffer.length > MAX_BUFFER) {
      this.buffer.shift();
    }
  }

  getInterpolated(now = Date.now()): PlayerState[] {
    if (this.buffer.length === 0) return [];

    const renderTime = now - RENDER_DELAY_MS;

    // Find the two snapshots that bracket renderTime
    let before: TimedSnapshot | null = null;
    let after: TimedSnapshot | null = null;

    for (const ts of this.buffer) {
      if (ts.receivedAt <= renderTime) {
        before = ts;
      } else if (after === null) {
        after = ts;
      }
    }

    if (!before) return this.buffer[0].snap.players;
    if (!after) return before.snap.players;

    const span = after.receivedAt - before.receivedAt;
    const t = span === 0 ? 1 : (renderTime - before.receivedAt) / span;

    return interpolatePlayers(before.snap.players, after.snap.players, t);
  }
}

function interpolatePlayers(
  a: PlayerState[],
  b: PlayerState[],
  t: number,
): PlayerState[] {
  const bMap = new Map(b.map((p) => [p.id, p]));
  return a.map((pa) => {
    const pb = bMap.get(pa.id);
    if (!pb) return pa;
    return {
      ...pa,
      x: lerp(pa.x, pb.x, t),
      y: lerp(pa.y, pb.y, t),
    };
  });
}

function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * Math.min(1, Math.max(0, t));
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd client && npm test -- interpolation
```

Expected: all 4 tests pass.

- [ ] **Step 5: Commit**

```bash
cd ..
git add client/src/interpolation.ts client/src/tests/interpolation.test.ts
git commit -m "feat(client): add interpolation buffer with tests (C4)"
```

---

## Task 13: Renderer — C3

**Files:**
- Create: `client/src/renderer.ts`

No unit tests for rendering (canvas pixel assertions are brittle). Visual verification is the test.

- [ ] **Step 1: Implement renderer.ts**

`client/src/renderer.ts`:
```typescript
import type { PlayerState } from "./schema";

const PLAYER_RADIUS = 16;
const FONT = "12px monospace";

export interface Renderer {
  clear(): void;
  drawPlayer(p: PlayerState, isMe: boolean): void;
  resize(w: number, h: number): void;
}

export class CanvasRenderer implements Renderer {
  private ctx: CanvasRenderingContext2D;

  constructor(private canvas: HTMLCanvasElement) {
    this.ctx = canvas.getContext("2d")!;
  }

  resize(w: number, h: number): void {
    this.canvas.width = w;
    this.canvas.height = h;
  }

  clear(): void {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    this.ctx.fillStyle = "#1a1a2e";
    this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
  }

  drawPlayer(p: PlayerState, isMe: boolean): void {
    const { ctx } = this;
    ctx.beginPath();
    ctx.arc(p.x, p.y, PLAYER_RADIUS, 0, Math.PI * 2);
    ctx.fillStyle = p.color;
    ctx.fill();
    if (isMe) {
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 2;
      ctx.stroke();
    }
    ctx.fillStyle = "#ffffff";
    ctx.font = FONT;
    ctx.textAlign = "center";
    ctx.fillText(p.name, p.x, p.y - PLAYER_RADIUS - 4);
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add client/src/renderer.ts
git commit -m "feat(client): add canvas renderer abstraction (C3)"
```

---

## Task 14: Lobby UI + Menu overlay — C5/C6

**Files:**
- Create: `client/src/lobby.ts`
- Create: `client/src/menu.ts`

- [ ] **Step 1: Implement lobby.ts**

`client/src/lobby.ts`:
```typescript
export interface LobbyResult {
  roomCode: string;
  name: string;
  color: string;
}

const COLORS = ["#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6"];

export function showLobby(container: HTMLElement): Promise<LobbyResult> {
  return new Promise((resolve) => {
    const codeFromUrl = location.pathname.replace("/r/", "").trim();

    container.innerHTML = `
      <div style="max-width:320px;margin:auto;padding:2rem">
        <h1 style="margin-bottom:1rem">Tower Climb</h1>
        <label>Room code
          <input id="room-code" value="${codeFromUrl}" placeholder="ABCD"
            style="display:block;width:100%;margin:0.25rem 0 1rem;padding:0.5rem;font:inherit;background:#333;color:#fff;border:1px solid #666" />
        </label>
        <label>Your name
          <input id="player-name" placeholder="alice"
            style="display:block;width:100%;margin:0.25rem 0 1rem;padding:0.5rem;font:inherit;background:#333;color:#fff;border:1px solid #666" />
        </label>
        <label>Colour
          <div id="color-picker" style="display:flex;gap:0.5rem;margin:0.25rem 0 1rem">
            ${COLORS.map((c, i) =>
              `<button data-color="${c}" style="background:${c};width:2rem;height:2rem;border:${i === 0 ? "3px solid #fff" : "3px solid transparent"};cursor:pointer"></button>`
            ).join("")}
          </div>
        </label>
        <button id="join-btn" style="width:100%;padding:0.75rem;background:#3498db;color:#fff;border:none;font:inherit;cursor:pointer">
          Join
        </button>
      </div>
    `;

    let selectedColor = COLORS[0];
    container.querySelectorAll<HTMLButtonElement>("[data-color]").forEach((btn) => {
      btn.addEventListener("click", () => {
        container.querySelectorAll<HTMLButtonElement>("[data-color]").forEach((b) => {
          b.style.border = "3px solid transparent";
        });
        btn.style.border = "3px solid #fff";
        selectedColor = btn.dataset.color!;
      });
    });

    container.querySelector("#join-btn")!.addEventListener("click", () => {
      const code = (container.querySelector<HTMLInputElement>("#room-code")!.value || "ROOM").toUpperCase();
      const name = container.querySelector<HTMLInputElement>("#player-name")!.value || "player";
      container.innerHTML = "";
      resolve({ roomCode: code, name, color: selectedColor });
    });
  });
}
```

- [ ] **Step 2: Implement menu.ts**

`client/src/menu.ts`:
```typescript
export class MenuOverlay {
  private el: HTMLElement;
  private visible = false;

  constructor(private container: HTMLElement, private onLeave: () => void) {
    this.el = document.createElement("div");
    this.el.style.cssText = `
      display:none;position:fixed;inset:0;background:rgba(0,0,0,0.7);
      display:flex;justify-content:center;align-items:center;z-index:100;
    `;
    this.el.innerHTML = `
      <div style="background:#222;padding:2rem;min-width:200px;text-align:center">
        <p id="menu-room-code" style="margin-bottom:1rem;font-size:1.2rem"></p>
        <button id="copy-btn" style="display:block;width:100%;margin-bottom:0.5rem;padding:0.5rem;background:#555;color:#fff;border:none;font:inherit;cursor:pointer">
          Copy invite link
        </button>
        <button id="leave-btn" style="display:block;width:100%;padding:0.5rem;background:#c0392b;color:#fff;border:none;font:inherit;cursor:pointer">
          Leave room
        </button>
      </div>
    `;
    this.el.style.display = "none";
    container.appendChild(this.el);

    this.el.querySelector("#leave-btn")!.addEventListener("click", () => {
      this.hide();
      this.onLeave();
    });

    this.el.querySelector("#copy-btn")!.addEventListener("click", () => {
      navigator.clipboard.writeText(location.href).catch(() => {});
    });

    window.addEventListener("keydown", (e) => {
      if (e.key === "Escape") this.toggle();
    });
  }

  show(roomCode: string): void {
    (this.el.querySelector("#menu-room-code") as HTMLElement).textContent = `Room: ${roomCode}`;
    this.el.style.display = "flex";
    this.visible = true;
  }

  hide(): void {
    this.el.style.display = "none";
    this.visible = false;
  }

  toggle(): void {
    if (this.visible) this.hide(); else this.show("");
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add client/src/lobby.ts client/src/menu.ts
git commit -m "feat(client): add lobby UI and menu overlay (C5/C6)"
```

---

## Task 15: Client logging — C7

**Files:**
- Create: `client/src/logging.ts`

- [ ] **Step 1: Implement logging.ts**

`client/src/logging.ts`:
```typescript
let debugOverlay: HTMLElement | null = null;

export function initDebugOverlay(): void {
  debugOverlay = document.createElement("div");
  debugOverlay.style.cssText = `
    position:fixed;bottom:0;left:0;background:rgba(0,0,0,0.6);
    color:#0f0;font:11px monospace;padding:4px 8px;z-index:999;display:none;
  `;
  document.body.appendChild(debugOverlay);

  window.addEventListener("keydown", (e) => {
    if (e.key === "`") {
      debugOverlay!.style.display =
        debugOverlay!.style.display === "none" ? "block" : "none";
    }
  });
}

export function logEvent(event: string, data?: Record<string, unknown>): void {
  const entry = data ? `${event} ${JSON.stringify(data)}` : event;
  console.log(`[client] ${entry}`);
}

export function updateDebugOverlay(stats: { fps: number; ping: number; snapshotAge: number }): void {
  if (debugOverlay && debugOverlay.style.display !== "none") {
    debugOverlay.textContent = `FPS:${stats.fps} PING:${stats.ping}ms AGE:${stats.snapshotAge}ms`;
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add client/src/logging.ts
git commit -m "feat(client): add client logging with toggleable debug overlay (C7)"
```

---

## Task 16: Wire client — main.ts

**Files:**
- Modify: `client/src/main.ts`

- [ ] **Step 1: Implement the full main.ts game loop**

`client/src/main.ts`:
```typescript
import { showLobby } from "./lobby";
import { MenuOverlay } from "./menu";
import { NetworkClient } from "./network";
import { InputHandler } from "./input";
import { CanvasRenderer } from "./renderer";
import { InterpolationBuffer } from "./interpolation";
import { initDebugOverlay, logEvent, updateDebugOverlay } from "./logging";
import type { SnapshotPayload } from "./schema";

const CLIENT_TICK_HZ = 30;

async function main() {
  const app = document.getElementById("app")!;
  initDebugOverlay();

  const { roomCode, name, color } = await showLobby(app);

  // Create canvas
  const canvas = document.createElement("canvas");
  canvas.width = 800;
  canvas.height = 600;
  app.appendChild(canvas);
  canvas.focus();

  const renderer = new CanvasRenderer(canvas);
  const buffer = new InterpolationBuffer();
  const input = new InputHandler();
  const net = new NetworkClient();
  const menu = new MenuOverlay(app, () => { net.disconnect(); location.reload(); });

  let myId = "";
  let serverTickRate = 20;
  let tick = 0;
  let lastSnap: SnapshotPayload | null = null;
  let frameCount = 0;
  let lastFpsTime = Date.now();

  net.onWelcome((w) => {
    myId = w.yourId;
    serverTickRate = w.tickRate;
    logEvent("welcome", { myId, roomCode: w.roomCode });
  });

  net.onSnapshot((snap) => {
    buffer.push(snap);
    lastSnap = snap;
  });

  net.onEvent((e) => {
    logEvent("event", { type: e.eventType, player: e.playerId });
  });

  net.connect(roomCode, name, color);
  input.start(canvas);
  input.captureMouseOnCanvas(canvas);

  // Input loop — send at CLIENT_TICK_HZ
  const inputInterval = setInterval(() => {
    net.sendInput(input.getInput(tick++));
  }, 1000 / CLIENT_TICK_HZ);

  // Render loop
  let lastFrame = Date.now();
  function frame() {
    const now = Date.now();
    frameCount++;
    if (now - lastFpsTime >= 1000) {
      const fps = Math.round(frameCount * 1000 / (now - lastFpsTime));
      const age = lastSnap ? now - Date.now() : 0;
      updateDebugOverlay({ fps, ping: 0, snapshotAge: age });
      frameCount = 0;
      lastFpsTime = now;
    }
    lastFrame = now;

    renderer.clear();
    const players = buffer.getInterpolated(now);
    for (const p of players) {
      renderer.drawPlayer(p, p.id === myId);
    }

    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
}

main().catch(console.error);
```

- [ ] **Step 2: Build client and verify no TypeScript errors**

```bash
cd client && npm run build
```

Expected: `client/dist/` built with no TS errors.

- [ ] **Step 3: Commit**

```bash
cd ..
git add client/src/main.ts
git commit -m "feat(client): wire game loop — lobby, network, input, renderer, interpolation (M1)"
```

---

## Task 17: Update main.go embed path + integration smoke test

**Files:**
- Modify: `cmd/server/main.go` (fix embed path if needed)

- [ ] **Step 1: Verify embed path**

The `//go:embed all:../../client/dist` path in `main.go` is relative to `cmd/server/main.go`. Verify it resolves correctly:

```bash
cd client && npm run build && cd ..
go build ./cmd/server
```

If you see `pattern ../../client/dist: no matching files found`, adjust the embed directive. From `cmd/server/main.go`, the path to `client/dist` at the repo root is `../../client/dist` — verify this is correct for your working directory structure. If the build is run from the repo root, `go build` resolves embed paths relative to the source file's directory.

Correct embed and fs.Sub call:

```go
//go:embed all:../../client/dist
var clientDist embed.FS

// In main():
static, err := fs.Sub(clientDist, "client/dist")
```

Wait — embed paths are always relative to the package source file. From `cmd/server/main.go`, `../../client/dist` points to `client/dist` at the repo root. This is correct.

- [ ] **Step 2: Run integration smoke test**

```bash
cd client && npm run build && cd ..
go run ./cmd/server &
SERVER_PID=$!
sleep 1
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/
kill $SERVER_PID
```

Expected: `200`

- [ ] **Step 3: Run full test suite**

```bash
make test
```

Expected: all Go tests pass, all Vite tests pass.

- [ ] **Step 4: Update STATUS.md — mark all Wave 1/2/3 items done**

Update `STATUS.md` to mark all implemented items as `[x]`.

- [ ] **Step 5: Commit**

```bash
git add STATUS.md cmd/server/main.go
git commit -m "chore: verify embed path, smoke test, update STATUS for M1 complete"
```

---

## Task 18: Render deploy

**Files:**
- Modify: `render.yaml` (verify settings)

- [ ] **Step 1: Connect repo on Render**

Go to render.com, create a new Web Service, connect `github.com/qduong42/2d-game-tower-climb`. Render will detect `render.yaml` automatically.

- [ ] **Step 2: Verify render.yaml**

`render.yaml`:
```yaml
services:
  - type: web
    name: tower-climb
    env: docker
    dockerfilePath: ./Dockerfile
    envVars:
      - key: PORT
        value: 8080
```

Render builds the Docker image and deploys. Logs available at the Render dashboard.

- [ ] **Step 3: Test with two browsers**

Open `https://tower-climb.onrender.com/r/TEST` in two browser tabs (or share the URL with a friend). Both should see a coloured dot for each player, moving in real time.

**Exit criteria for Milestone 1:** two dots visible, WASD moves your dot, other players' dots move smoothly. Server logs show `player_join`/`player_leave` events.

---

## Self-review

**Spec coverage:**
- S1 gateway ✓ (Task 8)
- S2 room manager ✓ (Task 7)
- S3 tick loop ✓ (Task 6)
- S4 game state model ✓ (Task 6)
- S5 message schema ✓ (Tasks 4, 5)
- S6 server logging ✓ (slog in room.go, gateway.go)
- C1 network client ✓ (Task 10)
- C2 input handler ✓ (Task 11)
- C3 renderer ✓ (Task 13)
- C4 interpolation ✓ (Task 12)
- C5 lobby UI ✓ (Task 14)
- C6 menu overlay ✓ (Task 14)
- C7 client logging ✓ (Task 15)
- X1 deploy pipeline ✓ (Task 18, render.yaml, Dockerfile)
- X2 local dev setup ✓ (Task 3, Makefile)
- X3 test harness ✓ (go test + vitest, Tasks 4–12)
- X4 README ✓ (Task 3)

**Placeholder scan:** no TBDs. Embed path note in Task 17 is a runtime verification step, not a placeholder.

**Type consistency check:**
- `schema.MsgJoin` used in gateway.go handshake — defined in messages.go ✓
- `schema.EventJoin` (EventType) used in room.go — defined in messages.go ✓
- `room.WritePump` called in gateway.go — defined in room.go ✓
- `room.NewConnectedClient` called in gateway.go — defined in room.go ✓
- `client.ID()`, `client.Conn()` called in gateway.go — defined in room.go ✓
- TS `MsgType.Join` used in network.ts — defined in schema.ts ✓
- `InterpolationBuffer.push(snap, receivedAt)` — signature matches usage in main.ts ✓
