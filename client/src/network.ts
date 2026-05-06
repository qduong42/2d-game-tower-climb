import {
  MsgType,
  type Envelope,
  type WelcomePayload,
  type SnapshotPayload,
  type EventPayload,
  type InputPayload,
} from "./schema";

const WELCOME_TIMEOUT_MS = 10_000;

export class NetworkClient {
  private ws: WebSocket | null = null;
  private onWelcomeCb: ((w: WelcomePayload) => void) | null = null;
  private onSnapshotCb: ((s: SnapshotPayload) => void) | null = null;
  private onEventCb: ((e: EventPayload) => void) | null = null;
  private onCloseCb: ((reason: string) => void) | null = null;
  private pendingInput: InputPayload | null = null;
  private welcomeTimer: ReturnType<typeof setTimeout> | null = null;
  private closeReported = false;

  connect(roomCode: string, name: string, color: string, wsUrl?: string): void {
    this.closeReported = false;
    const url = wsUrl ?? (() => {
      const protocol = location.protocol === "https:" ? "wss" : "ws";
      return `${protocol}://${location.host}/r/${roomCode}`;
    })();
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      const env: Envelope = {
        type: MsgType.Join,
        payload: { roomCode, name, color },
      };
      this.ws!.send(JSON.stringify(env));
      if (this.pendingInput) {
        this.ws!.send(JSON.stringify({ type: MsgType.Input, payload: this.pendingInput }));
        this.pendingInput = null;
      }
      this.welcomeTimer = setTimeout(() => {
        this.welcomeTimer = null;
        this.closeReported = true;
        this.onCloseCb?.("Server did not respond — try again");
        this.ws?.close();
      }, WELCOME_TIMEOUT_MS);
    };

    this.ws.onmessage = (e: MessageEvent) => {
      const env = JSON.parse(e.data as string) as Envelope;
      switch (env.type) {
        case MsgType.Welcome:
          if (this.welcomeTimer !== null) {
            clearTimeout(this.welcomeTimer);
            this.welcomeTimer = null;
          }
          this.onWelcomeCb?.(env.payload as WelcomePayload);
          break;
        case MsgType.Snapshot:
          this.onSnapshotCb?.(env.payload as SnapshotPayload);
          break;
        case MsgType.Event:
          this.onEventCb?.(env.payload as EventPayload);
          break;
      }
    };

    this.ws.onerror = () => {
      console.error("[network] WebSocket error");
      if (this.welcomeTimer !== null) {
        clearTimeout(this.welcomeTimer);
        this.welcomeTimer = null;
      }
      if (!this.closeReported) {
        this.closeReported = true;
        this.onCloseCb?.("Connection error — check server is running");
      }
    };

    this.ws.onclose = () => {
      console.warn("[network] connection closed");
      if (this.welcomeTimer !== null) {
        clearTimeout(this.welcomeTimer);
        this.welcomeTimer = null;
      }
      if (!this.closeReported) {
        this.closeReported = true;
        this.onCloseCb?.("Disconnected");
      }
    };
  }

  sendInput(payload: InputPayload): void {
    const env: Envelope = { type: MsgType.Input, payload };
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(env));
    } else {
      this.pendingInput = payload;
    }
  }

  onWelcome(cb: (w: WelcomePayload) => void): void { this.onWelcomeCb = cb; }
  onSnapshot(cb: (s: SnapshotPayload) => void): void { this.onSnapshotCb = cb; }
  onEvent(cb: (e: EventPayload) => void): void { this.onEventCb = cb; }
  onClose(cb: (reason: string) => void): void { this.onCloseCb = cb; }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}
