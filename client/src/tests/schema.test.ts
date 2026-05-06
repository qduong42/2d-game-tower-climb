import { describe, it, expect } from "vitest";
import type {
  Envelope,
  WelcomePayload,
  SnapshotPayload,
  InputPayload,
} from "../schema";
import { MsgType } from "../schema";

describe("schema types and constants", () => {
  it("MsgType constants match wire values", () => {
    expect(MsgType.Welcome).toBe("welcome");
    expect(MsgType.Snapshot).toBe("snapshot");
    expect(MsgType.Event).toBe("event");
    expect(MsgType.Join).toBe("join");
    expect(MsgType.Input).toBe("input");
  });

  it("Envelope can be constructed and serialized", () => {
    const payload: WelcomePayload = { yourId: "p1", roomCode: "ABCD", tickRate: 20 };
    const env: Envelope = { type: MsgType.Welcome, payload };
    const json = JSON.stringify(env);
    const parsed = JSON.parse(json) as Envelope;
    expect(parsed.type).toBe("welcome");
    expect((parsed.payload as WelcomePayload).yourId).toBe("p1");
  });

  it("InputPayload mouse field is optional", () => {
    const inp: InputPayload = {
      tick: 1,
      keys: { up: true, down: false, left: false, right: false, space: false },
    };
    expect(inp.mouse).toBeUndefined();
    const json = JSON.stringify(inp);
    expect(json).not.toContain("mouse");
  });

  it("SnapshotPayload players array", () => {
    const snap: SnapshotPayload = {
      tick: 42,
      players: [{ id: "p1", x: 10, y: 20, color: "#ff0000", name: "alice" }],
    };
    expect(snap.players[0]?.id).toBe("p1");
  });
});
