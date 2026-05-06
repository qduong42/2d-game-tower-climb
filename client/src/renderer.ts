import type { PlayerState } from "./schema";

const PLAYER_RADIUS = 16;
const FONT = "12px monospace";

export interface Renderer {
  clear(): void;
  drawPlayer(p: PlayerState, isMe: boolean): void;
  resize(w: number, h: number): void;
}

export class CanvasRenderer implements Renderer {
  private ctx: CanvasRenderingContext2D;

  constructor(private canvas: HTMLCanvasElement) {
    this.ctx = canvas.getContext("2d")!;
  }

  resize(w: number, h: number): void {
    this.canvas.width = w;
    this.canvas.height = h;
  }

  clear(): void {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    this.ctx.fillStyle = "#1a1a2e";
    this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
  }

  drawPlayer(p: PlayerState, isMe: boolean): void {
    const { ctx } = this;
    ctx.beginPath();
    ctx.arc(p.x, p.y, PLAYER_RADIUS, 0, Math.PI * 2);
    ctx.fillStyle = p.color;
    ctx.fill();
    if (isMe) {
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 2;
      ctx.stroke();
    }
    ctx.fillStyle = "#ffffff";
    ctx.font = FONT;
    ctx.textAlign = "center";
    ctx.fillText(p.name, p.x, p.y - PLAYER_RADIUS - 4);
  }
}
