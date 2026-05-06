import type { PlayerState, SnapshotPayload } from "./schema";

const NUM_PLATFORMS = 4;
const PLATFORM_Y = [490, 360, 230, 100]; // platform 0–3, bottom to top
const PLAYER_RADIUS = 14;
const PLATFORM_W = 100;
const COL_X = 400; // climber column centered for single-column view

export class CanvasRenderer {
  private ctx: CanvasRenderingContext2D;

  constructor(private canvas: HTMLCanvasElement) {
    this.ctx = canvas.getContext("2d")!;
  }

  render(snap: SnapshotPayload, myId: string): void {
    this.clear();

    if (snap.phase === "waiting") return; // status text handled by main.ts

    const me = snap.players.find(p => p.id === myId);
    if (!me) return;

    if (me.role === "base") {
      this.renderBaseView(snap, me);
    } else {
      this.renderClimberView(snap, me);
    }

    if (snap.phase === "won") {
      this.drawWinOverlay();
    }
  }

  // ── Climber view: own column centered ─────────────────────────────────────

  private renderClimberView(snap: SnapshotPayload, me: PlayerState): void {
    const { ctx, canvas } = this;

    this.drawColumn(COL_X);

    // Draw all players on this column
    for (const p of snap.players) {
      const x = this.playerX(p, me);
      const y = PLATFORM_Y[p.platform] ?? PLATFORM_Y[0]!;
      if (x !== null) {
        this.drawPlayerCircle(p, x, y, p.id === me.id);
      }
    }

    // Role header
    ctx.fillStyle = "#ffd700";
    ctx.font = "bold 14px monospace";
    ctx.textAlign = "center";
    ctx.fillText(me.climberIndex === 0 ? "CLIMBER — MID" : "CLIMBER — TOP", canvas.width / 2, 30);

    // Floor label for own position
    ctx.fillStyle = "#aaa";
    ctx.font = "11px monospace";
    ctx.textAlign = "left";
    for (let i = 0; i < NUM_PLATFORMS; i++) {
      const label = i === 0 ? "Ground" : i === NUM_PLATFORMS - 1 ? "Top ★" : `Floor ${i}`;
      ctx.fillText(label, COL_X + PLATFORM_W / 2 + 10, PLATFORM_Y[i]! + 4);
    }

    // Corner: base status only — climbers cannot see each other
    const base = snap.players.find(p => p.role === "base");
    if (base) this.drawCornerStatus(base, canvas.width - 160, 10);
  }

  // Returns the canvas X for a player in a climber's view.
  // Own player → COL_X. Base operator (when on same platform) → offset left.
  // Other climbers are not shown — each climber's column is their own view.
  private playerX(p: PlayerState, me: PlayerState): number | null {
    if (p.id === me.id) return COL_X;
    if (p.role === "base" && p.platform === me.platform) return COL_X - 34;
    return null;
  }

  // ── Base view: mission-control panel ─────────────────────────────────────

  private renderBaseView(snap: SnapshotPayload, me: PlayerState): void {
    const { ctx, canvas } = this;
    const W = canvas.width;

    // Header
    ctx.fillStyle = "#ffd700";
    ctx.font = "bold 18px monospace";
    ctx.textAlign = "center";
    ctx.fillText("BASE OPERATOR", W / 2, 48);

    // Own avatar + tool
    ctx.beginPath();
    ctx.arc(W / 2, 110, 20, 0, Math.PI * 2);
    ctx.fillStyle = me.color;
    ctx.fill();
    ctx.strokeStyle = "#fff";
    ctx.lineWidth = 3;
    ctx.stroke();

    ctx.fillStyle = "#fff";
    ctx.font = "12px monospace";
    ctx.textAlign = "center";
    ctx.fillText(me.name, W / 2, 143);

    if (me.hasTool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "bold 13px monospace";
      ctx.fillText("⚙ HOLDING TOOL", W / 2, 162);
    } else {
      ctx.fillStyle = "#778";
      ctx.font = "13px monospace";
      ctx.fillText("tool passed", W / 2, 162);
    }

    // Divider
    ctx.strokeStyle = "#334";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(60, 180);
    ctx.lineTo(W - 60, 180);
    ctx.stroke();

    // Climber status cards
    const climbers = snap.players
      .filter(p => p.role === "climber")
      .sort((a, b) => a.climberIndex - b.climberIndex);

    climbers.forEach((c, i) => {
      const cardX = 120 + i * 320;
      this.drawClimberCard(c, cardX, 210);
    });

    // Instruction
    const climberAtGround = snap.players.some(p => p.role === "climber" && p.platform === 0);
    if (me.hasTool && climberAtGround) {
      ctx.fillStyle = "#2ecc71";
      ctx.font = "bold 13px monospace";
      ctx.textAlign = "center";
      ctx.fillText("Press SPACE to pass tool →", W / 2, 490);
    } else if (me.hasTool) {
      ctx.fillStyle = "#aaa";
      ctx.font = "13px monospace";
      ctx.textAlign = "center";
      ctx.fillText("Waiting for a climber to reach ground level…", W / 2, 490);
    }
  }

  private drawClimberCard(p: PlayerState, x: number, y: number): void {
    const { ctx } = this;
    const W = 260, H = 260;

    // Card background
    ctx.fillStyle = "#1e2040";
    ctx.strokeStyle = "#445";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(x, y, W, H, 6);
    ctx.fill();
    ctx.stroke();

    // Climber label
    ctx.fillStyle = "#ffd700";
    ctx.font = "bold 12px monospace";
    ctx.textAlign = "center";
    ctx.fillText(p.climberIndex === 0 ? "CLIMBER — MID" : "CLIMBER — TOP", x + W / 2, y + 22);

    // Avatar
    ctx.beginPath();
    ctx.arc(x + W / 2, y + 60, 16, 0, Math.PI * 2);
    ctx.fillStyle = p.color;
    ctx.fill();
    ctx.fillStyle = "#fff";
    ctx.font = "11px monospace";
    ctx.textAlign = "center";
    ctx.fillText(p.name, x + W / 2, y + 90);

    if (p.hasTool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "12px monospace";
      ctx.fillText("⚙ has tool", x + W / 2, y + 108);
    }

    // Mini tower showing position
    const towerX = x + W / 2;
    const platformH = 36;
    const towerTop = y + 125;
    for (let i = NUM_PLATFORMS - 1; i >= 0; i--) {
      const py = towerTop + (NUM_PLATFORMS - 1 - i) * platformH;
      const isHere = p.platform === i;
      ctx.fillStyle = isHere ? p.color + "44" : "#ffffff0a";
      ctx.strokeStyle = isHere ? p.color : "#334";
      ctx.lineWidth = isHere ? 2 : 1;
      ctx.beginPath();
      ctx.rect(towerX - 60, py, 120, platformH - 4);
      ctx.fill();
      ctx.stroke();

      const label = i === 0 ? "Ground" : i === NUM_PLATFORMS - 1 ? "Top ★" : `Floor ${i}`;
      ctx.fillStyle = isHere ? "#fff" : "#556";
      ctx.font = isHere ? "bold 10px monospace" : "10px monospace";
      ctx.textAlign = "center";
      ctx.fillText(label, towerX, py + 18);

      if (isHere) {
        ctx.beginPath();
        ctx.arc(towerX - 40, py + 16, 6, 0, Math.PI * 2);
        ctx.fillStyle = p.color;
        ctx.fill();
      }
    }
  }

  // ── Shared drawing helpers ────────────────────────────────────────────────

  private drawColumn(x: number): void {
    const { ctx } = this;

    // Ladder line
    ctx.strokeStyle = "#334";
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(x, PLATFORM_Y[0]!);
    ctx.lineTo(x, PLATFORM_Y[NUM_PLATFORMS - 1]!);
    ctx.stroke();

    // Platforms
    for (let i = 0; i < NUM_PLATFORMS; i++) {
      const y = PLATFORM_Y[i]!;
      const isTop = i === NUM_PLATFORMS - 1;
      ctx.strokeStyle = isTop ? "#ffd700" : "#556";
      ctx.lineWidth = isTop ? 4 : 2;
      ctx.beginPath();
      ctx.moveTo(x - PLATFORM_W / 2, y);
      ctx.lineTo(x + PLATFORM_W / 2, y);
      ctx.stroke();
    }

    // TOP label
    ctx.fillStyle = "#ffd700";
    ctx.font = "bold 11px monospace";
    ctx.textAlign = "center";
    ctx.fillText("TOP", x, PLATFORM_Y[NUM_PLATFORMS - 1]! - 8);
  }

  private drawPlayerCircle(p: PlayerState, x: number, y: number, isMe: boolean): void {
    const { ctx } = this;
    ctx.beginPath();
    ctx.arc(x, y, PLAYER_RADIUS, 0, Math.PI * 2);
    ctx.fillStyle = p.color;
    ctx.fill();
    if (isMe) {
      ctx.strokeStyle = "#fff";
      ctx.lineWidth = 3;
      ctx.stroke();
    }
    ctx.fillStyle = "#fff";
    ctx.font = "11px monospace";
    ctx.textAlign = "center";
    ctx.fillText(p.name, x, y - PLAYER_RADIUS - 4);
    ctx.fillStyle = "#aaa";
    ctx.font = "9px monospace";
    ctx.fillText(p.role === "base" ? "BASE" : p.climberIndex === 0 ? "MID" : "TOP", x, y + PLAYER_RADIUS + 12);
    if (p.hasTool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "bold 14px monospace";
      ctx.fillText("⚙", x + PLAYER_RADIUS + 2, y - PLAYER_RADIUS + 4);
    }
  }

  private drawCornerStatus(p: PlayerState, x: number, y: number): void {
    const { ctx } = this;
    ctx.fillStyle = "#1e2040cc";
    ctx.beginPath();
    ctx.roundRect(x, y, 150, 48, 4);
    ctx.fill();

    ctx.beginPath();
    ctx.arc(x + 20, y + 24, 10, 0, Math.PI * 2);
    ctx.fillStyle = p.color;
    ctx.fill();

    ctx.fillStyle = "#fff";
    ctx.font = "10px monospace";
    ctx.textAlign = "left";
    const label = p.role === "base" ? "BASE" : p.climberIndex === 0 ? "MID" : "TOP";
    const floor = p.platform === 0 ? "Ground" : p.platform === NUM_PLATFORMS - 1 ? "Top" : `Floor ${p.platform}`;
    ctx.fillText(`${label}: ${p.name}`, x + 36, y + 18);
    ctx.fillStyle = "#aaa";
    ctx.fillText(`${floor}${p.hasTool ? " ⚙" : ""}`, x + 36, y + 34);
  }

  private clear(): void {
    const { ctx, canvas } = this;
    ctx.fillStyle = "#1a1a2e";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
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
