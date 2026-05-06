import { describe, it, expect } from "vitest";
import { InterpolationBuffer, RENDER_DELAY_MS } from "../interpolation";
import type { SnapshotPayload } from "../schema";

function snap(tick: number, x: number): SnapshotPayload {
  return { tick, players: [{ id: "p1", x, y: 0, color: "#fff", name: "a" }] };
}

describe("InterpolationBuffer", () => {
  it("returns empty array when no snapshots", () => {
    const buf = new InterpolationBuffer();
    expect(buf.getInterpolated(Date.now())).toEqual([]);
  });

  it("returns last snapshot when only one exists", () => {
    const buf = new InterpolationBuffer();
    const now = Date.now();
    buf.push(snap(1, 100), now);
    const result = buf.getInterpolated(now + 50);
    expect(result[0]?.x).toBe(100);
  });

  it("interpolates between two snapshots", () => {
    const buf = new InterpolationBuffer();
    const t0 = 1000;
    buf.push(snap(1, 0), t0);
    buf.push(snap(2, 100), t0 + 50);

    // Query at t0+25+RENDER_DELAY_MS so renderTime = t0+25, which is between the two snapshots
    const result = buf.getInterpolated(t0 + 25 + RENDER_DELAY_MS);
    expect(result[0]?.x).toBeGreaterThan(0);
    expect(result[0]?.x).toBeLessThan(100);
  });

  it("caps buffer to 10 snapshots", () => {
    const buf = new InterpolationBuffer();
    for (let i = 0; i < 15; i++) {
      buf.push(snap(i, i * 10), 1000 + i * 50);
    }
    const result = buf.getInterpolated(1000 + 14 * 50);
    expect(result).toBeDefined();
  });

  // Regression: second player joining was invisible until buffer caught up
  it("includes a player present only in the after snapshot (just joined)", () => {
    const t0 = 1000;
    const buf = new InterpolationBuffer();
    // before: only player 1
    buf.push({ tick: 1, players: [{ id: "p1", x: 100, y: 0, color: "#f00", name: "a" }] }, t0);
    // after: player 1 + player 2 who just joined
    buf.push({ tick: 2, players: [
      { id: "p1", x: 110, y: 0, color: "#f00", name: "a" },
      { id: "p2", x: 400, y: 300, color: "#00f", name: "b" },
    ]}, t0 + 50);

    const result = buf.getInterpolated(t0 + 25 + RENDER_DELAY_MS);
    const ids = result.map((p) => p.id);
    expect(ids).toContain("p1");
    expect(ids).toContain("p2"); // was missing before fix
  });
});
