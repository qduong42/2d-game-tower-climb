import type { InputPayload, MouseState } from "./schema";

const KEY_MAP: Record<string, keyof { up: boolean; down: boolean; left: boolean; right: boolean; space: boolean }> = {
  w: "up", ArrowUp: "up",
  s: "down", ArrowDown: "down",
  a: "left", ArrowLeft: "left",
  d: "right", ArrowRight: "right",
  " ": "space",
};

export class InputHandler {
  private keys = { up: false, down: false, left: false, right: false, space: false };
  private mouse: MouseState | null = null;
  private boundKeyDown: ((e: KeyboardEvent) => void) | null = null;
  private boundKeyUp: ((e: KeyboardEvent) => void) | null = null;

  start(target: EventTarget = window): void {
    this.boundKeyDown = (e: KeyboardEvent) => {
      const k = KEY_MAP[e.key];
      if (k) { this.keys[k] = true; e.preventDefault(); }
    };
    this.boundKeyUp = (e: KeyboardEvent) => {
      const k = KEY_MAP[e.key];
      if (k) { this.keys[k] = false; }
    };
    target.addEventListener("keydown", this.boundKeyDown as EventListener);
    target.addEventListener("keyup", this.boundKeyUp as EventListener);
  }

  stop(): void {
    if (this.boundKeyDown) window.removeEventListener("keydown", this.boundKeyDown);
    if (this.boundKeyUp) window.removeEventListener("keyup", this.boundKeyUp);
  }

  captureMouseOnCanvas(canvas: HTMLCanvasElement): void {
    canvas.addEventListener("click", (e) => {
      const rect = canvas.getBoundingClientRect();
      this.mouse = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
        click: true,
      };
    });
  }

  peekKeys(): typeof this.keys { return { ...this.keys }; }

  getInput(tick: number): InputPayload {
    const inp: InputPayload = { tick, keys: { ...this.keys }, mouse: this.mouse ?? undefined };
    this.mouse = null;
    return inp;
  }

  simulateKeyDown(key: string): void {
    const k = KEY_MAP[key];
    if (k) this.keys[k] = true;
  }
  simulateKeyUp(key: string): void {
    const k = KEY_MAP[key];
    if (k) this.keys[k] = false;
  }
  simulateMouseClick(x: number, y: number): void {
    this.mouse = { x, y, click: true };
  }
}
