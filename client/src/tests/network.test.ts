import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { NetworkClient } from "../network";
import type { SnapshotPayload, WelcomePayload } from "../schema";
import { MsgType } from "../schema";

class FakeWS {
  onmessage: ((e: MessageEvent) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onopen: (() => void) | null = null;
  sent: string[] = [];
  readyState = 1; // OPEN

  send(data: string) { this.sent.push(data); }
  close() { this.readyState = 3; this.onclose?.(); }

  receive(payload: unknown) {
    this.onmessage?.({ data: JSON.stringify(payload) } as MessageEvent);
  }

  simulateError() { this.onerror?.(); }
}

describe("NetworkClient", () => {
  let fakeWS: FakeWS;
  let client: NetworkClient;

  beforeEach(() => {
    fakeWS = new FakeWS();
    vi.stubGlobal("WebSocket", vi.fn(() => fakeWS));
    client = new NetworkClient();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("sends Join on connect", () => {
    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    fakeWS.onopen?.();
    expect(fakeWS.sent).toHaveLength(1);
    const msg = JSON.parse(fakeWS.sent[0]);
    expect(msg.type).toBe(MsgType.Join);
    expect(msg.payload.name).toBe("alice");
    expect(msg.payload.roomCode).toBe("ABCD");
  });

  it("calls onWelcome when welcome arrives", () => {
    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    fakeWS.onopen?.();

    const welcome: WelcomePayload = { yourId: "p1", roomCode: "ABCD", tickRate: 20, color: "#e74c3c" };
    let received: WelcomePayload | null = null;
    client.onWelcome((w) => { received = w; });
    fakeWS.receive({ type: MsgType.Welcome, payload: welcome });

    expect(received).not.toBeNull();
    expect(received!.yourId).toBe("p1");
  });

  it("calls onSnapshot when snapshot arrives", () => {
    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    fakeWS.onopen?.();

    const snap: SnapshotPayload = { tick: 1, players: [] };
    let received: SnapshotPayload | null = null;
    client.onSnapshot((s) => { received = s; });
    fakeWS.receive({ type: MsgType.Snapshot, payload: snap });

    expect(received).not.toBeNull();
    expect(received!.tick).toBe(1);
  });

  it("send queues until connected", () => {
    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    client.sendInput({ tick: 1, keys: { up: true, down: false, left: false, right: false, space: false } });
    fakeWS.onopen?.();
    expect(fakeWS.sent.length).toBeGreaterThanOrEqual(1);
  });

  // Regression: "stuck on connecting" — server reachable but Welcome never arrives
  it("calls onClose after timeout when server never sends Welcome", () => {
    vi.useFakeTimers();
    let closeReason: string | null = null;
    client.onClose((r) => { closeReason = r; });

    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    fakeWS.onopen?.(); // connection succeeded, Join sent, Welcome never arrives

    vi.advanceTimersByTime(9_999);
    expect(closeReason).toBeNull(); // not fired yet

    vi.advanceTimersByTime(1);
    expect(closeReason).toMatch(/did not respond/i);
  });

  // Regression: timer is cancelled when Welcome arrives before timeout
  it("does not call onClose when Welcome arrives within timeout", () => {
    vi.useFakeTimers();
    let closeReason: string | null = null;
    client.onClose((r) => { closeReason = r; });
    client.onWelcome(() => {});

    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    fakeWS.onopen?.();
    fakeWS.receive({ type: MsgType.Welcome, payload: { yourId: "p1", roomCode: "ABCD", tickRate: 20, color: "#e74c3c" } });

    vi.advanceTimersByTime(15_000);
    expect(closeReason).toBeNull();
  });

  // Regression: WebSocket error triggers onClose, not silent hang
  it("calls onClose when WebSocket emits an error", () => {
    let closeReason: string | null = null;
    client.onClose((r) => { closeReason = r; });

    client.connect("ABCD", "alice", "#ff0000", "ws://localhost/r/ABCD");
    fakeWS.onopen?.();
    fakeWS.simulateError();

    expect(closeReason).toMatch(/error/i);
  });
});
