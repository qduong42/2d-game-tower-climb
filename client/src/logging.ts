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
