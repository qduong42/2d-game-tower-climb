import type { PlayerState, SnapshotPayload } from "./schema";

const NUM_PLATFORMS = 4;
// Y position for each platform level (0=ground at bottom, 3=top)
const PLATFORM_Y = [490, 360, 230, 100];
// X position for each climber column
const CLIMBER_X = [230, 570];
const BASE_X = 400;
const PLAYER_RADIUS = 14;
const PLATFORM_W = 90;

export class CanvasRenderer {
  private ctx: CanvasRenderingContext2D;

  constructor(private canvas: HTMLCanvasElement) {
    this.ctx = canvas.getContext("2d")!;
  }

  render(snap: SnapshotPayload, myId: string): void {
    this.clear();
    this.drawTower();
    for (const p of snap.players) {
      this.drawPlayer(p, p.id === myId);
    }
    if (snap.phase === "won") {
      this.drawWinOverlay();
    }
  }

  private clear(): void {
    const { ctx, canvas } = this;
    ctx.fillStyle = "#1a1a2e";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
  }

  private drawTower(): void {
    const { ctx } = this;

    for (const colX of CLIMBER_X) {
      // Ladder line
      ctx.strokeStyle = "#334";
      ctx.lineWidth = 3;
      ctx.beginPath();
      ctx.moveTo(colX, PLATFORM_Y[0]!);
      ctx.lineTo(colX, PLATFORM_Y[NUM_PLATFORMS - 1]!);
      ctx.stroke();

      // Platform bars
      for (let i = 0; i < NUM_PLATFORMS; i++) {
        const y = PLATFORM_Y[i]!;
        const isTop = i === NUM_PLATFORMS - 1;
        ctx.strokeStyle = isTop ? "#ffd700" : "#556";
        ctx.lineWidth = isTop ? 4 : 2;
        ctx.beginPath();
        ctx.moveTo(colX - PLATFORM_W / 2, y);
        ctx.lineTo(colX + PLATFORM_W / 2, y);
        ctx.stroke();
      }

      // "TOP" label
      ctx.fillStyle = "#ffd700";
      ctx.font = "bold 11px monospace";
      ctx.textAlign = "center";
      ctx.fillText("TOP", colX, PLATFORM_Y[3]! - 8);
    }

    // Column labels
    ctx.fillStyle = "#778";
    ctx.font = "11px monospace";
    ctx.textAlign = "center";
    ctx.fillText("CLIMBER 1", CLIMBER_X[0]!, PLATFORM_Y[0]! + 20);
    ctx.fillText("CLIMBER 2", CLIMBER_X[1]!, PLATFORM_Y[0]! + 20);
    ctx.fillText("BASE", BASE_X, PLATFORM_Y[0]! + 20);
  }

  private drawPlayer(p: PlayerState, isMe: boolean): void {
    const { ctx } = this;
    let x: number;
    let y: number;

    if (p.role === "base") {
      x = BASE_X;
      y = PLATFORM_Y[0]!;
    } else {
      x = CLIMBER_X[p.climberIndex] ?? BASE_X;
      y = PLATFORM_Y[p.platform] ?? PLATFORM_Y[0]!;
    }

    // Player circle
    ctx.beginPath();
    ctx.arc(x, y, PLAYER_RADIUS, 0, Math.PI * 2);
    ctx.fillStyle = p.color;
    ctx.fill();
    if (isMe) {
      ctx.strokeStyle = "#fff";
      ctx.lineWidth = 3;
      ctx.stroke();
    }

    // Name
    ctx.fillStyle = "#fff";
    ctx.font = "11px monospace";
    ctx.textAlign = "center";
    ctx.fillText(p.name, x, y - PLAYER_RADIUS - 4);

    // Role badge
    const badge = p.role === "base" ? "BASE" : `C${p.climberIndex + 1}`;
    ctx.fillStyle = "#aaa";
    ctx.font = "9px monospace";
    ctx.fillText(badge, x, y + PLAYER_RADIUS + 12);

    // Tool indicator
    if (p.hasTool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "bold 14px monospace";
      ctx.fillText("⚙", x + PLAYER_RADIUS + 2, y - PLAYER_RADIUS + 4);
    }
  }

  private drawWinOverlay(): void {
    const { ctx, canvas } = this;
    ctx.fillStyle = "rgba(0,0,0,0.65)";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = "#ffd700";
    ctx.font = "bold 40px monospace";
    ctx.textAlign = "center";
    ctx.fillText("REPAIR COMPLETE!", canvas.width / 2, canvas.height / 2 - 20);
    ctx.font = "18px monospace";
    ctx.fillStyle = "#ccc";
    ctx.fillText("Refresh to play again", canvas.width / 2, canvas.height / 2 + 30);
  }
}
