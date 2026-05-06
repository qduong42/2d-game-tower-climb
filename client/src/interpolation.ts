import type { PlayerState, SnapshotPayload } from "./schema";

export const RENDER_DELAY_MS = 0; // kept for compat

export class InterpolationBuffer {
  private latest: SnapshotPayload | null = null;

  push(snap: SnapshotPayload): void {
    this.latest = snap;
  }

  getInterpolated(_now = Date.now()): PlayerState[] {
    return this.latest?.players ?? [];
  }

  getLatest(): SnapshotPayload | null {
    return this.latest;
  }
}
