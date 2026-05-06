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

  const canvas = document.createElement("canvas");
  canvas.width = 800;
  canvas.height = 600;
  app.appendChild(canvas);
  canvas.focus();

  const renderer = new CanvasRenderer(canvas);
  const buffer = new InterpolationBuffer();
  const input = new InputHandler();
  const net = new NetworkClient();
  new MenuOverlay(app, () => { net.disconnect(); location.reload(); });

  let myId = "";
  let tick = 0;
  let lastSnap: SnapshotPayload | null = null;
  let frameCount = 0;
  let lastFpsTime = Date.now();

  net.onWelcome((w) => {
    myId = w.yourId;
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

  setInterval(() => {
    net.sendInput(input.getInput(tick++));
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
      renderer.drawPlayer(p, p.id === myId);
    }

    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
}

main().catch(console.error);
