import { showLobby } from "./lobby";
import { MenuOverlay } from "./menu";
import { NetworkClient } from "./network";
import { InputHandler } from "./input";
import { CanvasRenderer } from "./renderer";
import { InterpolationBuffer } from "./interpolation";
import { initDebugOverlay, logEvent, updateDebugOverlay } from "./logging";
import type { SnapshotPayload } from "./schema";

const CLIENT_TICK_HZ = 30;
const SPEED = 200; // must match server game.Speed
const WORLD_W = 800;
const WORLD_H = 600;

async function main() {
  const app = document.getElementById("app")!;
  initDebugOverlay();

  const { roomCode, name, color: preferredColor } = await showLobby(app);

  const canvas = document.createElement("canvas");
  canvas.width = WORLD_W;
  canvas.height = WORLD_H;
  app.appendChild(canvas);

  const renderer = new CanvasRenderer(canvas);
  const buffer = new InterpolationBuffer();
  const input = new InputHandler();
  const net = new NetworkClient();
  new MenuOverlay(app, () => { net.disconnect(); location.reload(); });

  // Connection status overlay
  const statusEl = document.createElement("div");
  statusEl.style.cssText = "position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);color:#fff;font:1rem monospace;pointer-events:none";
  statusEl.textContent = "Connecting...";
  app.style.position = "relative";
  app.appendChild(statusEl);

  // Room code badge — top-left of canvas
  const roomBadge = document.createElement("div");
  roomBadge.style.cssText = "position:absolute;top:8px;left:8px;background:rgba(0,0,0,0.55);color:#fff;font:bold 0.85rem monospace;padding:4px 10px;border-radius:4px;pointer-events:none;letter-spacing:0.1em";
  roomBadge.textContent = `Room: ${roomCode}`;
  app.appendChild(roomBadge);

  let myId = "";
  let myColor = preferredColor;
  let tick = 0;
  let lastSnap: SnapshotPayload | null = null;
  let frameCount = 0;
  let lastFpsTime = Date.now();

  // Client-side prediction for own player
  let predX = 0;
  let predY = 0;
  let predicting = false;

  net.onWelcome((w) => {
    myId = w.yourId;
    myColor = w.color;
    statusEl.textContent = "";
    logEvent("welcome", { myId, roomCode: w.roomCode, color: myColor });
  });

  net.onSnapshot((snap) => {
    buffer.push(snap);
    lastSnap = snap;
    const me = snap.players.find(p => p.id === myId);
    if (me) {
      if (!predicting) {
        predX = me.x;
        predY = me.y;
        predicting = true;
      } else {
        // Smooth reconciliation: blend toward server position if drift is large
        const drift = Math.hypot(me.x - predX, me.y - predY);
        if (drift > 60) {
          predX = me.x;
          predY = me.y;
        } else if (drift > 4) {
          predX += (me.x - predX) * 0.3;
          predY += (me.y - predY) * 0.3;
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

  net.connect(roomCode, name, preferredColor);
  input.start(window);
  input.captureMouseOnCanvas(canvas);

  const dt = 1 / CLIENT_TICK_HZ;
  setInterval(() => {
    const inp = input.getInput(tick++);
    net.sendInput(inp);

    // Predict own movement immediately — same physics as server
    if (predicting) {
      if (inp.keys.left)  predX -= SPEED * dt;
      if (inp.keys.right) predX += SPEED * dt;
      if (inp.keys.up)    predY -= SPEED * dt;
      if (inp.keys.down)  predY += SPEED * dt;
      predX = Math.max(0, Math.min(WORLD_W, predX));
      predY = Math.max(0, Math.min(WORLD_H, predY));
    }
  }, 1000 / CLIENT_TICK_HZ);

  function frame() {
    const now = Date.now();
    frameCount++;
    if (now - lastFpsTime >= 1000) {
      const fps = Math.round(frameCount * 1000 / (now - lastFpsTime));
      const age = lastSnap ? now - lastFpsTime : 0;
      updateDebugOverlay({ fps, ping: 0, snapshotAge: age });
      frameCount = 0;
      lastFpsTime = now;
    }

    renderer.clear();
    const players = buffer.getInterpolated(now);
    for (const p of players) {
      if (p.id === myId && predicting) {
        renderer.drawPlayer({ ...p, x: predX, y: predY }, true);
      } else {
        renderer.drawPlayer(p, false);
      }
    }

    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
}

main().catch(console.error);
