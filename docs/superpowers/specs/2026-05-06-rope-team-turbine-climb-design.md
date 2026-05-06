# High-Level Components Draft — Rope-Team Turbine Climb

**Status:** Living document — updated as milestones complete.
**Last updated:** 2026-05-06
**Source context:** see `SESSION.md` for game concept and prior reasoning.

---

## 1. What we're building

A 2D side-view, browser-based, 3-5 player co-op game. Players climb a wind turbine roped together. One player at the base reads telemetry (wind, weather) and warns the climbers; climbers must brace against gusts, pass tools hand-to-hand, and cooperate on a puzzle in the nacelle at the top. Rounds are 3-5 minutes.

The design property that makes the game worth building is **forced coupling**: mechanics require real-time coordination between players, so it cannot collapse into parallel single-player.

---

## 2. Locked-in stack decisions

| Decision | Choice | Why | 
|---|---|---|
| Client | TypeScript + HTML5 Canvas, runs in browser | Zero install for players; share a URL |
| Server | Go, single binary | Goroutines fit "one room per goroutine"; deploys easily anywhere |
| Transport | WebSocket (single connection per client, both directions) | One protocol, ordered, easy to reason about for v1; we trade UDP-style speed for simplicity |
| Hosting | **Render free tier** as primary; **self-host via Cloudflare Tunnel** or **LAN-only** as fallbacks | Render auto-deploys on `git push`, no CLI, no credit card required. fly.io is excluded because it now requires a credit card for the free credit, and the dispatcher prefers not to enter one. Switch triggers documented below. |
| Lobby | Room codes in URL (`/r/ABCD`) | No accounts, no persistence, ephemeral rooms |
| Controls | Keyboard primary (WASD + space + multi-key combos for bracing); mouse capture wired but unused at v1 | Continuous keystate fits action gameplay; mouse stays available for future base-operator console |
| Visuals | Canvas primitives at v1; sprite swap behind a renderer abstraction at later milestone | Avoids the asset-pipeline rabbit hole |
| Tickrate | 30 Hz server tick, broadcast snapshots; clients render from latest snapshot (no interpolation delay in current build) | Raised from 20 Hz for responsiveness; interpolation buffer simplified to passthrough for platform-based movement |
| Discipline | Test-driven for server-side game logic and message handling; client logic isolated from rendering so it is testable | Locks in correctness early; safe refactors |

**Hosting: switch from Render to a fallback if any of these hit:**

- Cold-start delay (>10 s) is visible to a friend clicking the URL during a demo or playtest.
- Ping from a German player to the server consistently exceeds ~100 ms (Render free tier is Oregon-only).
- Free-tier resource limits (RAM, bandwidth, monthly hours) start blocking iteration.
- WebSocket reliability problems that don't reproduce locally.

**Fallback ladder (no credit card required):**

1. **Self-host via Cloudflare Tunnel** (or Tailscale Funnel, ngrok free). Your dev machine becomes the server with a public URL. Free, no CC, eliminates Oregon latency, no cold start. Cost: laptop must stay on while playing.
2. **LAN-only**. Host on local network, friends connect to your LAN IP. Free, only works for in-person play. The absolute floor — game still works.
3. **Koyeb / Render paid hobby** — only if both 1 and 2 are unworkable, and only if a CC-free path can be confirmed at that moment.

The components in §4 are platform-agnostic. The deploy pipeline (X1) holds the only platform-specific code; switching is one component's worth of work, not a rewrite.

---

## 3. Milestones

The plan is **strictly incremental**. Each milestone is independently runnable and demoable; we do not start the next one until the current one is deployed and works.

### ✅ Milestone 1 — Cursor party (DONE)

Proved the entire networking + deploy pipeline before any game logic existed.

- Room codes in URL (`/r/ABCD`), WebSocket, Go server, Vite client, Render deploy.
- Players pick a color, see each other move in real time via WASD.
- localStorage persists name/color/room across refresh.
- Client-side prediction, 30 Hz tick, WebSocket keepalive.

Exit criteria met: players in different networks see each other move. Server logs join/leave. Tests cover room manager and message round-trip.

### ✅ Level 0 — MVP platform climb (DONE)

The smallest complete game loop: 3 players, 1 tool, no wind, no timer.

- Tower with 4 platforms (0=ground, 1, 2, 3=top) and 2 climber columns.
- 3 players required to start: 1 base operator + 2 climbers, roles assigned randomly.
- Base operator holds the tool and passes it via `Space` to a climber at the same platform.
- Climbers move up/down with `Up`/`Down` (rising-edge, one platform per keypress).
- `Space` passes tool to any player on the same platform.
- Win: any climber reaches platform 3 carrying the tool.
- Waiting phase shows "x/3 players" until 3 join; room locks at 3.

Exit criteria met: a round can be played end-to-end. Roles, tool passing, and win condition all work.

**Design decisions locked in at Level 0 (carry forward):**
- 4 platforms per column, not 3.
- Space key to pass (not automatic at top platform).
- All players visible on one shared tower canvas (base is between the two columns).

### Milestone 2 — Wind and role-specific UI (next)

Add the coordination pressure that makes the game interesting.

- **Wind gusts**: periodic gusts that knock players off platforms or slide them down ladders if they don't brace.
- **Information asymmetry**: base operator sees the warning countdown; climbers do not. Base must call it out verbally.
- **Brace mechanic**: while on a ladder, hold the key opposite to wind direction to stay put. Cannot brace and hold a tool at the same time.
- **Base operator screen**: separate view — weather dashboard (wind speed, direction, gust countdown) + tool queue UI. No tower visible.
- **Multiple tools**: repair list with 2-3 sequential items; Climber 2 sees one item at a time.
- **Cursor-party lobby**: while waiting for 3 players, free-roaming dots (no tower, no roles).
- **Role announcement**: 3-second overlay on game start showing each player their role.
- **Fail condition**: timer expires before all items delivered.

Exit criteria: a round with wind is demonstrably harder without verbal coordination. Base must warn climbers or they get knocked down.

### Milestone 3 — Polish and stretch (remainder of budget)

Pulled from backlog, picked by remaining time:

- Disconnect handling (30s freeze, rejoin window).
- Role shown persistently in top bar (x/3 + role name).
- Sprite swap behind renderer abstraction.
- Sound effects, juice, particles.
- Traitor mode (stretch from SESSION.md).

---

## 4. Components

### 4.1 Server (Go)

| # | Component | Responsibility | Notes |
|---|---|---|---|
| S1 | **WebSocket gateway** | Accept connections, parse room code from URL, hand connection to room | Use `nhooyr.io/websocket` or `gorilla/websocket`. Pick one in plan. |
| S2 | **Room manager** | Create/find rooms by code, route messages, evict empty rooms after a timeout | One goroutine per room; channels in/out |
| S3 | **Game tick loop** | Per-room fixed-rate tick (20 Hz), advance state, broadcast snapshot | Pure function `step(state, inputs, dt) -> state` — testable without a network |
| S4 | **Game state model** | Players, positions, (later) ropes/tools/wind/tower | Snapshot serializable as JSON v1; consider binary later |
| S5 | **Message schema** | Versioned client↔server messages; one source of truth | Generated or hand-mirrored TS types on the client. Decide in plan. |
| S6 | **Logging / observability** | Structured logs (`log/slog`): join, leave, room create/destroy, error paths, tick anomalies; counter for active rooms and players | Stdout JSON; fly.io captures it. No external metrics service at v1. |

### 4.2 Client (TypeScript)

| # | Component | Responsibility | Notes |
|---|---|---|---|
| C1 | **Network client** | WebSocket connect, send input, receive snapshots, reconnect with backoff | Single class; tests stub the WebSocket. |
| C2 | **Input handler** | Capture keyboard state continuously and mouse clicks (captured but unbound), emit to network at tickrate | Polled, not event-driven, so multi-key combos work. |
| C3 | **Renderer (abstraction)** | Public API: `drawPlayer`, `drawTower`, `drawRope`, `drawWindIndicator`, etc. | v1 implementation: canvas primitives. v3 implementation: spritesheet blits. Game logic never touches `ctx` directly. |
| C4 | **Interpolation buffer** | Render remote players ~100ms in the past, smooth between snapshots | Standard technique; isolate so it is unit-testable on a fake snapshot stream. |
| C5 | **Lobby UI** | Two screens: "create room" and "join with code"; then transitions to canvas | Plain HTML/CSS, no framework needed. |
| C6 | **Menu / pause overlay** | Leave-room, show your code, copy invite link, disconnect handling | Minimal; one HTML overlay toggled by Esc. |
| C7 | **Client logging** | Console events for connect/disconnect/error/desync; optional in-game debug overlay (FPS, ping, last snapshot age) | Toggleable with a dev key (e.g. backtick). |

### 4.3 Cross-cutting

| # | Component | Responsibility | Notes |
|---|---|---|---|
| X1 | **Deploy pipeline** | `Dockerfile` + `fly.toml`; `flyctl deploy` ships server **and** static client (Go binary serves built JS/HTML) | One artifact, one URL, one region. |
| X2 | **Local dev setup** | One command (e.g. `make dev` or `bun run dev` + `go run ./cmd/server`) starts server + watches client | A new collaborator should be playing locally within 5 minutes of clone. |
| X3 | **Test harness** | Go: standard `testing` + table-driven tests for tick step and message handling. Client: vitest (or equivalent) for network client, input handler, interpolation buffer. | Test-first for game logic and protocol. UI rendering not unit-tested at v1. |
| X4 | **Onboarding doc (`README.md`)** | What the project is, how to run, how to deploy, how to add a feature, current milestone status | Updated whenever a milestone completes. |

---

## 5. Message schema (current as of Level 0)

Source of truth: `internal/schema/messages.go` (Go) mirrored in `client/src/schema.ts` (TypeScript).

### Client → Server

- `Join { roomCode, name, color }` — sent once on connect
- `Input { tick, keys: { up, down, left, right, space }, mouse?: { x, y, click } }` — sent every client tick (30 Hz)

### Server → Client

- `Welcome { yourId, roomCode, tickRate, color }` — sent once on join
- `Snapshot { tick, phase, players: [{ id, color, name, role, climberIndex, platform, hasTool }] }` — broadcast at 30 Hz
- `Event { eventType, playerId? }` — discrete events (join, leave, error)

`phase` is one of `"waiting" | "playing" | "won"`. `role` is `"base" | "climber"`. `climberIndex` is 0 or 1 for climbers, -1 for base. `platform` is 0–3.

---

## 6. Open questions (deferred to implementation plan)

- **WebSocket library** for Go: `nhooyr.io/websocket` (modern, context-aware) vs `gorilla/websocket` (battle-tested). Lean toward nhooyr.
- **Client build tool**: Vite vs Bun bundler vs esbuild directly. Lean toward Vite for HMR.
- **Schema sharing**: hand-mirrored types vs codegen from a schema file. For v1, hand-mirroring is fine if both sides import from one canonical doc.
- **Authoritative physics**: server-authoritative (server runs physics, clients only render) vs client prediction with reconciliation. v1 = server-authoritative, no prediction. Revisit if it feels laggy.
- **Tickrate tuning**: 20 Hz is a starting point; may need 30 Hz for the brace mechanic.

---

## 7. Out of scope for the 10-12h budget

To make the simplicity bar explicit, these are **not** on the table unless a milestone finishes early:

- Accounts, persistence, leaderboards.
- Voice chat (use Discord/whatever externally).
- Anti-cheat.
- Spectator mode.
- Mobile / touch controls.
- Multiple regions or matchmaking.
- CI pipeline (run tests locally before push).
- Metrics/APM beyond stdout logs.

---

## 8. Contribution model & parallel-agent framework

The components in §4 are intentionally drawn so most can be built in parallel. The contribution model is designed for both human collaborators and Claude subagents — anything one can do, the other can too.

### 8.1 Why parallelism is the point

The classical single-developer flow leaves leverage unused. With well-bounded components and clear interfaces, multiple workstreams can advance at once: one agent writes the room manager while another builds the renderer abstraction while a third sets up the deploy pipeline. The dispatcher (a human, or the main Claude session) only blocks on integration points, not on each component.

This is the reason §4 has so many small components rather than three big ones, and the reason §3 (milestones) is strictly sequential while §4 (components within a milestone) is mostly parallel.

### 8.2 Independence map (Milestone 1)

Components that can be built **in parallel** with no shared edits:

| Wave | Parallel batch | Why parallel |
|---|---|---|
| Wave 1 | **S5** (message schema) — written first, alone | Everything else depends on it |
| Wave 2 | **S1+S2** (gateway+room mgr), **C1** (network client), **C5+C6** (lobby+menu UI), **X1+X2** (deploy + dev setup), **X3** (test harness scaffolding) | Each touches a separate file/package; only contract is S5 |
| Wave 3 | **S3** (tick loop), **C2** (input), **C3** (renderer stub), **C4** (interpolation), **S6+C7** (logging) | Depends on Wave 2 wiring; still parallel within the wave |

The dispatcher writes/approves S5, then fans out Wave 2, then Wave 3.

### 8.3 Standard contribution package

Every contributor — human or agent — ships the same shape:

1. **Branch** named `component/<id>-<slug>`, e.g. `component/S3-tick-loop`.
2. **Tests first** for any component listed in §4.3 X3 (server game logic, network client, input handler, interpolation).
3. **Implementation** until tests pass and the public API in §4 is satisfied.
4. **No edits outside the component's owned files** except S5 (schema), which is updated only by explicit handoff.
5. **PR / commit message** lists: which component, which tests now pass, integration concerns the dispatcher should know.
6. **STATUS update** — see §8.7.

### 8.4 What a collaborator can contribute

Concrete inputs a human or agent can give, in rough order of leverage:

1. **Implement a component** end-to-end from §4 (the main flow).
2. **Review a PR** — read tests, run locally, flag schema drift or boundary leaks.
3. **Add tests** to an under-tested area (e.g. server-side gust handling).
4. **Tune parameters** — gravity, brace duration, tickrate, rope tension — and report what felt right.
5. **Bug triage** — reproduce a report, narrow it to a component, file a focused issue.
6. **Asset generation** (Milestone 3 only) — produce sprites that fit the renderer API in §4.2 C3.
7. **Documentation** — keep `README.md` and this file accurate as the project evolves.
8. **Refactor behind an interface** — e.g. swap the WebSocket library, replace primitives with sprites — without changing callers.

Things contributors should **not** do without dispatcher sign-off: change the message schema (S5), change the public renderer API (C3), introduce a new dependency, or restructure milestones.

### 8.5 Human vs. agent — what each is for

The contribution model treats humans and agents as interchangeable for many inputs, but they are not interchangeable for all of them. Being explicit prevents wasted effort on both sides.

**Agents handle well:**

- Implementing a well-specified component end-to-end (most of §4).
- Writing tests under §4.3 X3 discipline.
- Refactoring inside a boundary, swapping a library, mirroring a schema.
- Boilerplate, type plumbing, repetitive edits.
- Running many components in parallel.

**Humans (the dispatcher / product owner) carry the work agents cannot:**

1. **Taste.** Does the brace combo feel satisfying? Is gravity right? Does the gust warning give enough time? Tests pass on tunable values; only a player decides which value is *right*. This is the largest single human role on a game project.
2. **Dispatcher judgement.** What gets built next, what gets cut, when a milestone is done, when a design needs to be thrown away. Delegating this causes drift.
3. **Domain expertise.** Wind turbines, base-operator UX, what "feels authentic." Agents only know the spec; the dispatcher knows the world.
4. **Architectural noticing.** Agents stay inside their assigned component (it is the boundary rule). Noticing inter-component drift, promoting a pattern, restructuring milestones — these are cross-cutting and human.
5. **Playtesting.** Agents run tests; only humans run playtests with real friends and report what is confusing or fun.
6. **Creative leaps.** Mechanic re-themings, scope inversions, "what if it were a social deduction game" — humans still lead.

**Caveats, honestly:**

- The line moves. Anything in the "agents handle well" list above expanded recently; expect more to migrate. Do not anchor a workflow on the current split.
- For a 10-12h hackathon, the human role concentrates on **dispatcher + taste + playtester + creative director**. Implementation is mostly delegated. That is the leverage.
- The product owner role does not delegate. If you skip it, the project drifts.

A useful daily rhythm in this model: dispatcher writes/refines a component spec → fans out 2-3 agents in parallel → reviews and integrates → playtests → adjusts feel → repeat.

### 8.6 Subagent framework

Create `.claude/agents/component-builder.md` (during implementation, not now). The agent's job is to take a single component from §4 from spec to passing tests. Its prompt should encode:

- **Read order**: `SESSION.md` → `high-level-components-draft.md` → `STATUS.md` → the component's own spec section.
- **Discipline**: TDD per §4.3 X3. Tests fail before implementation exists; tests pass before commit.
- **Boundary rule**: only edit files inside the component's owned path (defined in the agent prompt per dispatch).
- **Schema rule**: never change `schema/` files; if the component needs a schema change, stop and report instead.
- **Reporting**: at completion, write a short summary — files touched, tests added, anything the dispatcher should integrate. Update STATUS.

The dispatcher (main session) launches multiple `component-builder` agents in a single message via the `Agent` tool — this is the parallelism mechanism. Each runs in its own context, so the main session's context is not consumed by implementation detail.

A second agent type, `component-reviewer`, may be useful: takes a branch, runs tests, checks the boundary rule, reports issues. Lower priority for v1.

### 8.7 Coordination surface — `STATUS.md`

A single file, kept short, updated by every contributor on PR open and merge:

```
## Milestone 1 — Cursor party
- [x] S5 message schema — merged
- [ ] S1 gateway — in progress (agent: builder-3)
- [ ] S2 room manager — in progress (agent: builder-1)
- [ ] C1 network client — open PR
- [ ] C5 lobby UI — not started
…
```

This is the work queue. The dispatcher reads it, picks the next unblocked component, dispatches an agent or assigns a human. Anyone can read it to see what is safe to grab.

### 8.8 New collaborator quickstart (5 minutes)

1. Read `SESSION.md` for the *why*.
2. Read this file for the *what* and *how*.
3. Read `STATUS.md` for what is in flight.
4. Pick an unblocked component from §4. If you're a human, claim it in `STATUS.md`. If you're an agent, you were told which one.
5. Branch, test-first, implement, PR, update `STATUS.md`.

If you are the dispatcher (human product owner): your job is §8.5 items 1-6, not implementation. Resist the pull to write code yourself when an agent could do it; spend the time on taste, playtests, and what to build next.

---

## 9. Level design — Milestone 2

**Last updated:** 2026-05-06

### 9.1 Tower layout

The screen is divided into vertical segments, one per climber. Each segment has **four platforms**: ground (0), low (1), mid (2), top (3). Platforms are fixed horizontal surfaces a player can stand on.

```
[TOP    (3)]      ← win condition / nacelle delivery
      |
  (ladder)
      |
[MID    (2)]
      |
  (ladder)
      |
[LOW    (1)]
      |
  (ladder)
      |
[GROUND (0)]      ← starting platform; base operator and climbers meet here

        [BASE OPERATOR]   ← between the two columns, stays at ground level
```

The base technician is shown between the two climber columns at ground level. The nacelle (repair target) sits above the top climber's top platform.

### 9.2 Movement

| Input | Action |
|---|---|
| `Space` + `Left` | Jump left to the platform on the same level in the segment to the left |
| `Space` + `Right` | Jump right (same level, segment to the right) |
| `Up` / `Down` | Move up/down the ladder within own segment |
| `Left` / `Right` on ladder | Hold to resist wind (see §9.4) |

Players cannot move to a different segment mid-climb — segments are owned. Jumping left/right only applies at platform level for item hand-offs.

### 9.3 Item transfer

- Press **`Space`** to pass the tool to any other player on the **same platform**.
- The tool transfers instantly — no animation delay.
- Only one item exists at a time in Level 0; multi-item queuing is a Milestone 2 feature.
- The top climber delivers to the nacelle the same way: reach platform 3 carrying the tool → win.

### 9.4 Wind mechanics

Each gust has two phases:

| Phase | Visible to | Description |
|---|---|---|
| **Warning** | Base operator only | Countdown on weather dashboard before gust hits. Climbers do not see it — base must call it out verbally. |
| **Duration** | All players | Active gust timer shown on every screen while wind is blowing. |

Wind direction varies per gust (left or right) — shown on the base operator's dashboard during the warning phase, and as a visible wind indicator on all screens during the active phase.

**On a platform:**
- Wind knocks the player off the platform down to the next lower platform in their segment.
- No way to resist wind while standing on a platform — you must accept the knockdown.
- Carrying an item when knocked off: item is **dropped** onto the landing platform (not lost).

**On a ladder (between platforms):**
- Player must hold the **arrow key opposite to the wind direction** for the duration of the gust to stay on the ladder.
- If they fail (wrong key, no key, or carrying an item): they slide down to the platform below.
- You **cannot hold on and carry an item at the same time** — the core tension of the game.

### 9.5 Base technician

- Sees a weather dashboard: wind speed, direction, and a countdown to next gust.
- Can **queue items** to send up (wrench, sensor, coolant canister, etc.) — items appear at climber 1's bottom platform.
- Cannot climb. Cannot be hit by wind.
- The only player with full visibility of incoming weather; all other players depend on the base calling out gusts.

### 9.6 Wrong tool penalty (Option A — dead time)

If the wrong tool arrives at the nacelle the top climber must send it back down. No hard block — the chain just has to run in reverse, costing time. The base operator is responsible for reading what the nacelle needs and sending in the right order.

- Top climber carries wrong tool back to their bottom platform → it transfers back to the climber below → passed down the chain → returned to base.
- Base can then queue and send the correct tool.
- Clock keeps running throughout. No item is ever destroyed.

### 9.7 Drop rules

Items fall exactly **one platform level** when dropped — never more.

- Dropped on top platform → lands on mid platform (same segment).
- Dropped on mid platform → lands on bot platform (same segment).
- Dropped on bot platform → lands on bot platform (stays, does not leave the tower).
- Dropped on ladder → lands on the platform immediately below the player's current position.

This means a fumble during a gust is always recoverable — you climb back down one level and pick it up. A chain of fumbles costs time but never loses the item permanently.

### 9.8 Win / fail condition

- **Win**: all required items delivered to the nacelle before the timer expires.
- **Fail**: timer runs out. No permanent loss of items or players — difficulty comes from time pressure and coordination, not permadeath.

### 9.9 Open questions for playtesting

- Upward-travelling wind delay between segments — creates a warning window or feels unfair?
- How many items per round? (Starting guess: 3 items for a 3-minute round.)

---

## 10. Full game specification (3-player) — Milestone 2 target

**Last updated:** 2026-05-06

Items marked ✅ are implemented in Level 0. Unmarked items are targets for Milestone 2.

### 10.1 Player count and room rules

- ✅ **Max players per room: 3** (1 base operator + 2 climbers).
- ✅ **Room locks** when the game starts (3/3 reached). No joins after that.
- Player count badge shows **x/3** in the top of the game view at all times. *(text shown in waiting phase; persistent badge not yet implemented)*
- A 4th player attempting to join a full room sees a popup:
  > *"This game is full. Would you like to create your own?"*
  with a newly generated room code pre-filled. They can edit and join/create.

### 10.2 Lobby (waiting state)

- ✅ Top of screen shows: **"Waiting for players: x/3"**
- ✅ No game clock. Wind does not run.
- While fewer than 3 players are connected, everyone should be in **cursor-party mode**: free-roaming dots, no roles, no tower. *(currently shows waiting text only — no dots)*

### 10.3 Game start and role assignment

- ✅ When the 3rd player joins, roles are **assigned randomly**:
  - 1× Base operator (gets the tool)
  - 1× Climber 1 (`climberIndex: 0`)
  - 1× Climber 2 (`climberIndex: 1`)
- All players should see a **role announcement** on screen (e.g. "You are: Base Operator") for ~3 seconds before gameplay begins. *(not yet implemented)*
- Each player's role should be shown persistently in the top bar alongside the x/3 count. *(not yet implemented; role badge is drawn below player circle in the tower view)*

### 10.4 Screen layouts by role

**Base operator**
- *(Milestone 2 target)* Completely different screen — no tower visible. Weather dashboard: wind speed, direction, countdown to next gust. Tool queue UI: select and send tools one at a time.
- *(Level 0 current)* Shown between the two climber columns at ground level. Holds the tool; passes it with `Space` to a climber at the same platform. Cannot move up.
- Cannot be affected by wind.

**Climber 1 (middle segment)**
- Sees their own 3 platforms (bottom, mid, top) connected by a ladder.
- Ladder extends visually upward (implying the tower continues above).
- Tools appear at their bottom platform when the base operator sends them.

**Climber 2 (top segment)**
- Sees their own 3 platforms connected by a ladder.
- Top platform shows two repair components:
  - **Blade** — right side
  - **Generator** (drawn as a box) — left side
- Repair list visible only to this player: shows **one item at a time** — the next required tool is revealed only after the previous one is successfully delivered.
- Tools appear at their bottom platform when Climber 1 reaches their top platform carrying the tool.

### 10.5 Tool transfer chain

```
Base operator selects tool
  → appears at Climber 1's BOTTOM platform

Climber 1 picks up tool, climbs to their TOP platform
  → tool disappears from Climber 1
  → tool appears at Climber 2's BOTTOM platform

Climber 2 picks up tool, climbs to their TOP platform
  → stands next to the correct component (blade or generator)
  → delivers tool → component repaired
```

- Only one item in transit per segment at a time.
- Transfer is automatic when the carrier reaches the top platform — no button press.
- **Delivery requires standing next to the correct component.** Wrong component = rejected, no penalty beyond wasted movement.

### 10.6 Disconnect handling

- Any player disconnecting mid-game triggers a **full freeze**:
  - All movement stops.
  - Wind pauses.
  - Round clock pauses.
  - Screen dims for all players.
  - Countdown shown to all: **"[Name] disconnected — resuming in 30s or when they return"**
- If player reconnects within 30s → game resumes from frozen state.
- If 30s expires → game exits, all players return to lobby (same room code, roles cleared).
- Slot does not reopen — room stays locked at 3 players throughout.

### 10.7 Specification tests (user perspective)

Each test describes observable behaviour from a player's point of view.

#### Lobby and room filling

| # | Scenario | Expected |
|---|---|---|
| L1 | First player joins room ABCD | Sees "Waiting for players: 1/3", free-roaming dot |
| L2 | Second player joins same room | Both see "Waiting for players: 2/3" |
| L3 | Third player joins | All three see role announcement, then game starts |
| L4 | Fourth player tries to join full room | Sees popup "This game is full. Would you like to create your own?" with new code pre-filled |
| L5 | Fourth player clicks "yes" on popup | Lobby pre-filled with new code, ready to create new room |

#### Role assignment

| # | Scenario | Expected |
|---|---|---|
| R1 | Game starts (3/3) | Each player sees their role name on screen for ~3s |
| R2 | After announcement | Role shown persistently in top bar next to x/3 |
| R3 | Base operator's screen | Weather dashboard, tool queue — no tower |
| R4 | Climber 1's screen | 3 platforms, ladder, no repair list visible |
| R5 | Climber 2's screen | 3 platforms, blade on right, generator on left, repair list visible |

#### Tool transfer

| # | Scenario | Expected |
|---|---|---|
| T1 | Base sends wrench | Wrench appears at Climber 1's bottom platform |
| T2 | Climber 1 picks up wrench and reaches top platform | Wrench disappears from Climber 1, appears at Climber 2's bottom platform |
| T3 | Climber 1 reaches top platform without a tool | Nothing appears at Climber 2's bottom platform |
| T4 | Climber 2 stands next to blade with correct tool | Tool consumed, blade marked repaired |
| T5 | Climber 2 stands next to blade with wrong tool | Rejected — tool stays in hand, no repair |
| T6 | Climber 2 delivers wrong tool to nacelle area | Must carry back down, base resends correct tool |
| T7 | Base tries to send second tool while first is still at Climber 1's bottom platform | Second tool queued — not sent until first clears |

#### Wind mechanics

| # | Scenario | Expected |
|---|---|---|
| W1 | Wind hits player standing on mid platform | Player knocked down to bottom platform |
| W2 | Wind hits player on top platform | Player knocked down to mid platform |
| W3 | Wind hits player on ladder, holds opposite key | Player stays on ladder |
| W4 | Wind hits player on ladder, no key held | Player slides down to platform below |
| W5 | Wind hits player carrying item on platform | Player knocked down one level, item dropped on landing platform |
| W6 | Wind hits player carrying item on ladder | Player slides down, item drops to platform below current position |
| W7 | Item dropped from top platform during wind | Item lands on mid platform (not bottom) |
| W8 | Item dropped from bottom platform | Item stays on bottom platform |

#### Disconnect

| # | Scenario | Expected |
|---|---|---|
| D1 | Player disconnects mid-game | Screen dims for all, countdown "…resuming in 30s…" visible |
| D2 | Disconnected player reconnects in time | Countdown clears, game resumes, all movement restores |
| D3 | 30s expires without reconnect | All players returned to lobby, same room code, roles cleared |
| D4 | 4th player tries to join during freeze | Still rejected — room is locked |

#### Wind timing and information asymmetry

| # | Scenario | Expected |
|---|---|---|
| G1 | Gust incoming — warning phase | Base operator sees countdown + direction on dashboard. Climbers see nothing. |
| G2 | Gust hits — active phase | All players see active gust timer and wind direction indicator on screen |
| G3 | Wind blows left, climber on ladder holds Right | Climber stays on ladder |
| G4 | Wind blows left, climber on ladder holds Left | Climber slides down to platform below |
| G5 | Wind blows right, climber on ladder holds Right | Climber slides down to platform below |
| G6 | Wind blows right, climber on ladder holds Left | Climber stays on ladder |
| G7 | Gust ends | Wind indicator disappears, all players free to move normally |

#### Repair list (Climber 2)

| # | Scenario | Expected |
|---|---|---|
| P1 | Game starts | Climber 2 sees first required tool only (e.g. "Blade needs: Wrench") |
| P2 | First tool delivered successfully | First item clears, second item revealed (e.g. "Generator needs: Sensor") |
| P3 | Wrong tool delivered to correct component | Rejected, same item still shown in repair list |
| P4 | All items delivered | Repair list shows complete, win condition triggers |

#### Win / fail

| # | Scenario | Expected |
|---|---|---|
| V1 | All required items delivered before timer | Win screen shown to all players |
| V2 | Timer expires before all items delivered | Fail screen shown to all players |
| V3 | Win or fail screen shown | Room returns to lobby state (cursor party), same code |
