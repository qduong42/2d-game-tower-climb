# High-Level Components Draft — Rope-Team Turbine Climb

**Status:** Draft for collaborator onboarding. Not yet a binding implementation plan.
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
| Tickrate | 20 Hz server tick, broadcast snapshots; clients render with ~100ms interpolation buffer | Standard for small-state co-op |
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

### Milestone 1 — Cursor party (target: ~2-3h)

Prove the entire networking + deploy pipeline before any game logic exists.

- Open `https://<app>.fly.dev/r/ABCD` in two browsers.
- Each player picks a color, sees their own dot and the other dots moving in real time via WASD.
- Server runs in fly.io. Local dev works with one command.
- No game rules. No tower. Just dots in a box.

Exit criteria: 2-4 players in different networks see each other move smoothly. Server logs show join/leave events. Tests cover the room manager and message round-trip.

### Milestone 2 — Vertical slice of the climb (target: ~4-6h)

The smallest game that demonstrates forced coupling.

- A vertical tower (one tall rectangle).
- Climbers move up/down with WASD, gravity pulls them down if they let go.
- A rope constraint: if climber A is too far from climber B, A is pulled.
- One scripted gust event: a warning appears, climbers must hold a brace combo (e.g. Up+Right) for 2 seconds or get pushed off (caught by rope, costs time).
- Win state: top reached.

Exit criteria: a round can be played end-to-end, the rope and the brace mechanic both demonstrably require coordination.

### Milestone 3 — Asymmetry, polish, stretch (target: remainder of budget)

Pulled from a backlog, picked by remaining time:

- Base operator role with a separate UI (wind dashboard, gust call-out button).
- Tool pass mechanic.
- Nacelle puzzle at the top.
- Sprite swap (AI-generated or Kenney pack) behind the renderer abstraction.
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

## 5. Message schema (sketch, will harden in plan)

### Client → Server

- `Join { roomCode, name, color }` — sent once on connect
- `Input { tick, keys: { up, down, left, right, space, … }, mouse?: { x, y, clicks } }` — sent every client tick (~30 Hz)

### Server → Client

- `Welcome { yourId, roomCode, tickRate }` — sent once on join
- `Snapshot { tick, players: [{ id, x, y, color, state }] }` — broadcast at server tickrate
- `Event { type, payload }` — discrete events (join, leave, gust warning, win, error)

Schema is versioned with an integer; mismatched versions reject the connection with a clear error.

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
