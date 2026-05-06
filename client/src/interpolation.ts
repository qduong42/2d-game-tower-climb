import type { PlayerState, SnapshotPayload } from "./schema";

interface TimedSnapshot {
  snap: SnapshotPayload;
  receivedAt: number;
}

export const RENDER_DELAY_MS = 100;
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

    let before: TimedSnapshot | null = null;
    let after: TimedSnapshot | null = null;

    for (const ts of this.buffer) {
      if (ts.receivedAt <= renderTime) {
        before = ts;
      } else if (after === null) {
        after = ts;
      }
    }

    if (!before) return this.buffer[0]!.snap.players;
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
  const aMap = new Map(a.map((p) => [p.id, p]));
  const bMap = new Map(b.map((p) => [p.id, p]));
  const result: PlayerState[] = a.map((pa) => {
    const pb = bMap.get(pa.id);
    if (!pb) return pa;
    return { ...pa, x: lerp(pa.x, pb.x, t), y: lerp(pa.y, pb.y, t) };
  });
  // Include players present in `after` but not `before` (just joined)
  for (const pb of b) {
    if (!aMap.has(pb.id)) result.push(pb);
  }
  return result;
}

function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * Math.min(1, Math.max(0, t));
}
