import { showLobby } from "./lobby";
import { MenuOverlay } from "./menu";
import { NetworkClient } from "./network";
import { InputHandler } from "./input";
import { CanvasRenderer } from "./renderer";
import { InterpolationBuffer } from "./interpolation";
import { initDebugOverlay, logEvent, updateDebugOverlay } from "./logging";
import type { SnapshotPayload } from "./schema";

const CLIENT_TICK_HZ = 30;
const WORLD_W = 800;
const WORLD_H = 600;

async function main() {
  const app = document.getElementById("app")!;
  initDebugOverlay();

  const { roomCode, name, color: preferredColor, isPrivate } = await showLobby(app);

  const canvas = document.createElement("canvas");
  canvas.width = WORLD_W;
  canvas.height = WORLD_H;
  app.appendChild(canvas);

  const renderer = new CanvasRenderer(canvas);
  const buffer = new InterpolationBuffer();
  const input = new InputHandler();
  const net = new NetworkClient();
  new MenuOverlay(app, () => { net.disconnect(); location.reload(); });

  const statusEl = document.createElement("div");
  statusEl.style.cssText = "position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);color:#fff;font:1rem monospace;pointer-events:none;text-align:center;white-space:pre-line";
  statusEl.textContent = "Connecting...";
  app.style.position = "relative";
  app.appendChild(statusEl);

  const roomBadge = document.createElement("div");
  roomBadge.style.cssText = "position:absolute;top:8px;left:8px;background:rgba(0,0,0,0.55);color:#fff;font:bold 0.85rem monospace;padding:4px 10px;border-radius:4px;pointer-events:none;letter-spacing:0.1em";
  roomBadge.textContent = `Room: ${roomCode}`;
  app.appendChild(roomBadge);

  let myId = "";
  let myRole = "";
  let lastSnap: SnapshotPayload | null = null;
  let tick = 0;
  let frameCount = 0;
  let lastFpsTime = Date.now();

  net.onWelcome((w) => {
    myId = w.yourId;
    statusEl.textContent = "Waiting for players (1/3)...";
    logEvent("welcome", { myId, roomCode: w.roomCode });
  });

  net.onSnapshot((snap) => {
    buffer.push(snap);
    lastSnap = snap;

    if (snap.phase === "waiting") {
      const count = snap.players.length;
      statusEl.textContent = `Waiting for players (${count}/3)...\nNeed ${3 - count} more`;
    } else {
      statusEl.textContent = "";
      if (!myRole) {
        const me = snap.players.find(p => p.id === myId);
        if (me) {
          myRole = me.role;
        }
      }
    }
  });

  net.onEvent((e) => {
    logEvent("event", { type: e.eventType, player: e.playerId });
  });

  net.onClose((reason) => {
    statusEl.textContent = `⚠ ${reason} — refresh to reconnect`;
    statusEl.style.color = "#e74c3c";
  });

  net.connect(roomCode, name, preferredColor, isPrivate);
  input.start(window);

  setInterval(() => {
    const inp = input.getInput(tick++);
    net.sendInput(inp);
  }, 1000 / CLIENT_TICK_HZ);

  function frame() {
    const now = Date.now();
    frameCount++;
    if (now - lastFpsTime >= 1000) {
      const fps = Math.round(frameCount * 1000 / (now - lastFpsTime));
      updateDebugOverlay({ fps, ping: 0, snapshotAge: 0 });
      frameCount = 0;
      lastFpsTime = now;
    }

    if (lastSnap) {
      const me = lastSnap.players.find(p => p.id === myId);
      let boundaryHint = "";
      if (me?.role === "climber") {
        const keys = input.peekKeys();
        if (me.climberIndex === 0 && keys.up && me.platform === 3) {
          boundaryHint = "You are only responsible for Ground → Floor 3!";
        } else if (me.climberIndex === 1 && keys.down && me.platform === 3) {
          boundaryHint = "You are only responsible for Floor 3 → Top!";
        }
      }
      renderer.render(lastSnap, myId, boundaryHint);
    }

    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
}

main().catch(console.error);
