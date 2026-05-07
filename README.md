# Turbine Repair

2D side-view multiplayer co-op game. Players climb a wind turbine together.

## Quick start

Prerequisites: Go 1.22+, Node 20+

```bash
# Install client deps
cd client && npm install && cd ..

# Run dev server + client (two terminals)
go run ./cmd/server       # terminal 1 — listens on :8080
cd client && npm run dev  # terminal 2 — Vite dev server on :5173
```

Open `http://localhost:5173/r/TEST` in two browsers.

## Run tests

```bash
make test
```

## Deploy to Render

Connect this repo on render.com — `render.yaml` handles the rest. No credit card required on the free tier.

## Architecture

See `docs/superpowers/specs/2026-05-06-rope-team-turbine-climb-design.md` for the full design spec including component map, milestone plan, and contribution model.

## Current milestone

See `STATUS.md`.

## Issues:

Please create issues with this format

```
[Given]: The initial state or context, such as pre-conditions.
[When]: The specific action, event, or behavior carried out.
[Then]: The expected result, consequence, or outcome.
[And] / [But]: Used to add further steps to any of the three previous keywords.
```
