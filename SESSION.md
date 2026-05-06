# Session Export — 2D Multiplayer Game Project

**Date:** 2026-05-06
**Working dir:** `/home/huy/throwaway/2d-game-claude/`
**User context:** Wind turbine data engineer (Python/SQL daily). Wants a 10-12h "vibe code" project in an unfamiliar language to test Claude's capabilities.

---

## How we landed here

1. Started by asking for project ideas unrelated to wind turbine data engineering.
2. Narrowed to 10-12h scope, deliberately in a language the user doesn't know.
3. Pivoted to multiplayer sandbox interest → networking primer.
4. Narrowed to 4-5 player simple games.
5. Re-themed around wind turbines (user's domain knowledge as design advantage).
6. Picked the option with the highest *real* player-to-player interaction.

---

## Networking primer (recap)

Three architectures, simplest → hardest:

1. **LAN-only** — host opens TCP/UDP socket, others connect to local IP. Same WiFi only.
2. **Dedicated cloud server** — $5/mo VPS (Hetzner / DigitalOcean / Fly.io). Public IP, no NAT issues. Standard for indie multiplayer.
3. **Peer-to-peer with NAT traversal** — STUN/TURN, hole-punching, libp2p or WebRTC. Skip for v1.

State exchange basics:
- Tickrate 20-60 Hz; client → server → broadcast.
- UDP for fast-moving (positions). TCP for important events (chat, game state changes).
- Render others ~100 ms in the past + interpolate to hide jitter.

**Recommended stack:** Browser client (TypeScript + WebSocket + Canvas) + small Node/Bun server on Fly.io free tier. Lowest install friction, cross-platform free, dopamine-fast iteration.

---

## Chosen game: Rope-team turbine climb

Won the "highest player interaction" filter because mechanics *force* coordination. You can't fake it with parallel single-player.

### Why this one

- **Forced coupling**: tools must pass hand-to-hand, ropes physically link climbers, base operator sees data climbers cannot.
- **Voice chat becomes essential** — true marker of real multiplayer.
- **Short rounds (3-5 min)** → high replay, easy onboarding.
- **Small synced state**: positions, inventory, who holds what rope.
- **Asymmetry is built in**: base operator reads wind/weather; climbers act on shouted warnings — leverages user's domain expertise.

### v0 spec

- **View:** 2D side-view tower, ~200 m tall.
- **Roles:**
  - 1 base operator at console (wind data, weather radar, harness lock/unlock).
  - 3-4 climbers ascending in a chain, linked by rope.
- **Tools:** shared pool at base. Must be passed hand-to-hand up the chain.
- **Top of tower:** 60-second nacelle puzzle requiring 2 climbers cooperating (one holds panel open, other reroutes cables).
- **Random gust events:** base must call them out; climbers must brace within 2 seconds or fall (caught by rope, costs time).

---

## Open scope decisions for next session

- **Language**: Rust + WASM? TypeScript end-to-end? Godot (cheating mode — multiplayer almost free)?
- **Hosting**: local LAN first vs. straight to Fly.io.
- **Visual style**: pixel art vs. vector / canvas primitives.
- **Persistence**: ephemeral rooms vs. accounts/leaderboards.
- **Stretch**: traitor mode (one player as remote attacker injecting false telemetry — flips it into social deduction).

---

## Other game ideas we discussed (kept for reference)

| Idea | Type | Interaction level |
|---|---|---|
| Rope-team turbine climb | Co-op | **Highest** ← chosen |
| Control room with traitor mode | Social deduction co-op | Very high |
| Wind Wars (.io style) | Competitive | High (turbines steal wind via wake) |
| Forecast Bluff Poker | Card / bluff | Medium |
| Wind Farm Pictionary | Party | Medium (turn-based) |
| Overcooked control room (no traitor) | Co-op | Risk of parallel single-player |
| Wind Farm Tycoon | Co-op roles | Risk of parallel single-player |
| Shared turbine designer | Sandbox | Low |

---

## Next-session starting point

Pick:
1. Language + framework.
2. LAN or hosted.
3. Whether to start with a "cursor party" warm-up to nail networking before adding game mechanics.

Then scope architecture: server tick loop, message schema, client state model, render loop.
