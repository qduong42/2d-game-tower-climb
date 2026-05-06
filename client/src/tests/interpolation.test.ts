import { describe, it, expect } from "vitest";
import { InterpolationBuffer } from "../interpolation";
import type { SnapshotPayload } from "../schema";

function snap(tick: number, platform: number): SnapshotPayload {
  return {
    tick,
    phase: "playing",
    players: [{ id: "p1", color: "#fff", name: "a", role: "climber", climberIndex: 0, platform, tool: "", heldTools: [], selectedTool: "" }],
  };
}

describe("InterpolationBuffer", () => {
  it("returns empty array when no snapshots", () => {
    const buf = new InterpolationBuffer();
    expect(buf.getInterpolated()).toEqual([]);
  });

  it("returns player from pushed snapshot", () => {
    const buf = new InterpolationBuffer();
    buf.push(snap(1, 2));
    const result = buf.getInterpolated();
    expect(result[0]?.id).toBe("p1");
    expect(result[0]?.platform).toBe(2);
  });

  it("returns latest snapshot when multiple pushed", () => {
    const buf = new InterpolationBuffer();
    buf.push(snap(1, 0));
    buf.push(snap(2, 1));
    const result = buf.getInterpolated();
    expect(result[0]?.platform).toBe(1);
  });

  it("includes a player present in the latest snapshot", () => {
    const buf = new InterpolationBuffer();
    buf.push({ tick: 1, phase: "playing", players: [
      { id: "p1", color: "#f00", name: "a", role: "climber", climberIndex: 0, platform: 0, tool: "", heldTools: [], selectedTool: "" },
    ]});
    buf.push({ tick: 2, phase: "playing", players: [
      { id: "p1", color: "#f00", name: "a", role: "climber", climberIndex: 0, platform: 0, tool: "", heldTools: [], selectedTool: "" },
      { id: "p2", color: "#00f", name: "b", role: "base", climberIndex: -1, platform: 0, tool: "", heldTools: ["wrench", "hammer"], selectedTool: "wrench" },
    ]});
    const result = buf.getInterpolated();
    const ids = result.map((p) => p.id);
    expect(ids).toContain("p1");
    expect(ids).toContain("p2");
  });

  it("getLatest returns the pushed snapshot", () => {
    const buf = new InterpolationBuffer();
    const s = snap(5, 3);
    buf.push(s);
    expect(buf.getLatest()).toBe(s);
  });
});
