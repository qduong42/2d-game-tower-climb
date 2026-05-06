# Milestone Status

## Milestone 1 — Cursor party

> Goal: 2–4 players open a shared URL, see each other's coloured dots move in real time.

### Wave 1 (schema — must complete before Wave 2)
- [x] S5 message schema (Go + TS) — done

### Wave 2 (parallel — blocked on Wave 1)
- [x] S1 WebSocket gateway — done
- [x] S2 room manager — done
- [x] C1 network client — done
- [x] C5 lobby UI — done
- [x] C6 menu overlay — done
- [ ] X1 deploy pipeline — not started (Render deploy, Task 18)
- [x] X2 local dev setup — done (Makefile)

### Wave 3 (parallel — blocked on Wave 2)
- [x] S3 tick loop — done
- [x] C2 input handler — done
- [x] C3 renderer — done
- [x] C4 interpolation buffer — done
- [x] S6 server logging — done (slog JSON)
- [x] C7 client logging — done (debug overlay)

### Integration
- [x] Wire server main.go — done (embed + WebSocket gateway)
- [x] Wire client main.ts — done (full game loop)
- [x] Manual integration test — done (smoke test HTTP 200)
- [ ] Render deploy — not started (Task 18)

## Milestone 2 — Vertical slice
Not started. Begins after Milestone 1 is deployed and working.
