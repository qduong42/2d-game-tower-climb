import type { PlayerState, SnapshotPayload } from "./schema";

const NUM_PLATFORMS = 7;
const MID_MAX_PLATFORM = Math.floor(NUM_PLATFORMS / 2); // = 3, the handoff platform
const PLATFORM_Y = [540, 460, 380, 300, 220, 140, 60]; // platform 0–6, bottom to top
const PLAYER_RADIUS = 14;
const PLATFORM_W = 100;
const COL_X = 400; // climber column centered for single-column view

export class CanvasRenderer {
  private ctx: CanvasRenderingContext2D;

  constructor(private canvas: HTMLCanvasElement) {
    this.ctx = canvas.getContext("2d")!;
  }

  render(snap: SnapshotPayload, myId: string, boundaryHint = ""): void {
    this.clear();

    if (snap.phase === "waiting") return; // status text handled by main.ts

    const me = snap.players.find(p => p.id === myId);
    if (!me) return;

    if (me.role === "base") {
      this.renderBaseView(snap, me);
    } else {
      this.renderClimberView(snap, me, boundaryHint);
    }

    if (snap.phase === "won") {
      this.drawWinOverlay();
    }
  }

  // ── Climber view: own column centered ─────────────────────────────────────

  private renderClimberView(snap: SnapshotPayload, me: PlayerState, boundaryHint = ""): void {
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

    // Tool display — placeholder for sprite
    if (me.tool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "bold 13px monospace";
      ctx.textAlign = "center";
      ctx.fillText(`⚙ ${me.tool.toUpperCase()}`, canvas.width / 2, 50);
    }

    // Floor label for own position
    ctx.font = "11px monospace";
    ctx.textAlign = "left";
    for (let i = 0; i < NUM_PLATFORMS; i++) {
      const isHandoff = i === MID_MAX_PLATFORM;
      const label = i === 0 ? "Ground" : i === NUM_PLATFORMS - 1 ? "Top ★" : isHandoff ? "Handoff" : `Floor ${i}`;
      ctx.fillStyle = isHandoff ? "#e67e22" : "#aaa";
      ctx.fillText(label, COL_X + PLATFORM_W / 2 + 10, PLATFORM_Y[i]! + 4);
    }

    // SPACE hints — UP: MID at handoff passes to TOP; DOWN: MID at ground returns to BASE
    if (me.climberIndex === 0 && me.tool) {
      if (me.platform === MID_MAX_PLATFORM) {
        const top = snap.players.find(p => p.role === "climber" && p.climberIndex === 1);
        if (top && top.platform === MID_MAX_PLATFORM) {
          ctx.fillStyle = "#2ecc71";
          ctx.font = "bold 13px monospace";
          ctx.textAlign = "center";
          ctx.fillText("Press SPACE to pass tool to TOP ↑", canvas.width / 2, canvas.height - 20);
        } else {
          ctx.fillStyle = "#aaa";
          ctx.font = "13px monospace";
          ctx.textAlign = "center";
          ctx.fillText("Waiting for TOP climber to reach this level…", canvas.width / 2, canvas.height - 20);
        }
      } else if (me.platform === 0) {
        ctx.fillStyle = "#e67e22";
        ctx.font = "bold 13px monospace";
        ctx.textAlign = "center";
        ctx.fillText("Press SPACE to return tool to BASE ↓", canvas.width / 2, canvas.height - 20);
      }
    }

    // Boundary hint (shown when player presses the blocked direction)
    if (boundaryHint) {
      ctx.fillStyle = "#e74c3c";
      ctx.font = "bold 13px monospace";
      ctx.textAlign = "center";
      ctx.fillText(boundaryHint, canvas.width / 2, 76);
    }

    // SPACE hint — DOWN: TOP at handoff returns tool to MID
    if (me.climberIndex === 1 && me.tool && me.platform === MID_MAX_PLATFORM) {
      const mid = snap.players.find(p => p.role === "climber" && p.climberIndex === 0);
      if (mid && mid.platform === MID_MAX_PLATFORM) {
        ctx.fillStyle = "#e67e22";
        ctx.font = "bold 13px monospace";
        ctx.textAlign = "center";
        ctx.fillText("Press SPACE to pass tool back to MID ↓", canvas.width / 2, canvas.height - 20);
      } else {
        ctx.fillStyle = "#aaa";
        ctx.font = "13px monospace";
        ctx.textAlign = "center";
        ctx.fillText("Waiting for MID climber to reach handoff floor…", canvas.width / 2, canvas.height - 20);
      }
    }
  }

  // Returns the canvas X for a player in a climber's view.
  // Own player → COL_X. Everyone else is not shown — each climber's column is their own.
  private playerX(p: PlayerState, me: PlayerState): number | null {
    if (p.id === me.id) return COL_X;
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

    // Tool inventory: show each held tool as a box, highlight the selected one
    const toolEmoji: Record<string, string> = { wrench: "🔧 WRENCH", hammer: "🔨 HAMMER" };
    const toolBoxW = 130, toolBoxH = 36, toolBoxY = 158;
    const totalW = me.heldTools.length * toolBoxW + (me.heldTools.length - 1) * 12;
    let bx = W / 2 - totalW / 2;
    me.heldTools.forEach((t) => {
      const isSelected = t === me.selectedTool;
      ctx.fillStyle = isSelected ? "#2c3e8a" : "#1e2040";
      ctx.strokeStyle = isSelected ? "#ffd700" : "#445";
      ctx.lineWidth = isSelected ? 2 : 1;
      ctx.beginPath();
      ctx.roundRect(bx, toolBoxY, toolBoxW, toolBoxH, 4);
      ctx.fill();
      ctx.stroke();
      ctx.fillStyle = isSelected ? "#ffd700" : "#aaa";
      ctx.font = isSelected ? "bold 11px monospace" : "11px monospace";
      ctx.textAlign = "center";
      ctx.fillText(toolEmoji[t] ?? t.toUpperCase(), bx + toolBoxW / 2, toolBoxY + 23);
      bx += toolBoxW + 12;
    });
    if (me.heldTools.length === 0) {
      ctx.fillStyle = "#556";
      ctx.font = "12px monospace";
      ctx.textAlign = "center";
      ctx.fillText("no tools remaining", W / 2, toolBoxY + 23);
    } else {
      ctx.fillStyle = "#556";
      ctx.font = "10px monospace";
      ctx.textAlign = "center";
      ctx.fillText("←/→ to select", W / 2, toolBoxY + toolBoxH + 14);
    }

    // Divider
    ctx.strokeStyle = "#334";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(60, 220);
    ctx.lineTo(W - 60, 220);
    ctx.stroke();

    // Climber status cards
    const climbers = snap.players
      .filter(p => p.role === "climber")
      .sort((a, b) => a.climberIndex - b.climberIndex);

    climbers.forEach((c, i) => {
      const cardX = 120 + i * 320;
      this.drawClimberCard(c, cardX, 240);
    });

    // Instruction
    const climberAtGround = snap.players.some(p => p.role === "climber" && p.platform === 0);
    if (me.heldTools.length > 0 && climberAtGround) {
      ctx.fillStyle = "#2ecc71";
      ctx.font = "bold 13px monospace";
      ctx.textAlign = "center";
      ctx.fillText(`Press SPACE to pass ${me.selectedTool} →`, W / 2, canvas.height - 20);
    } else if (me.heldTools.length > 0) {
      ctx.fillStyle = "#aaa";
      ctx.font = "13px monospace";
      ctx.textAlign = "center";
      ctx.fillText("Waiting for a climber to reach ground level…", W / 2, canvas.height - 20);
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

    if (p.tool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "12px monospace";
      ctx.fillText(`⚙ ${p.tool}`, x + W / 2, y + 108);
    }

    // Mini tower showing position
    const towerX = x + W / 2;
    const towerTop = y + 125;
    const platformH = Math.floor((H - 130) / NUM_PLATFORMS); // shrink rows to fit card
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

      const label = i === 0 ? "Ground" : i === NUM_PLATFORMS - 1 ? "Top ★" : i === MID_MAX_PLATFORM ? "Handoff" : `Floor ${i}`;
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
      const isHandoff = i === MID_MAX_PLATFORM;
      ctx.strokeStyle = isTop ? "#ffd700" : isHandoff ? "#e67e22" : "#556";
      ctx.lineWidth = isTop ? 4 : isHandoff ? 3 : 2;
      ctx.beginPath();
      ctx.moveTo(x - PLATFORM_W / 2, y);
      ctx.lineTo(x + PLATFORM_W / 2, y);
      ctx.stroke();
    }

    // "TOP" label removed: the floor-label strip already shows "Top ★" beside
    // the top platform, and the centred label at y=52 conflicted with the tool
    // text drawn at y=50 in renderClimberView (issue #17).
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
    if (p.tool) {
      ctx.fillStyle = "#ffd700";
      ctx.font = "bold 14px monospace";
      ctx.fillText("⚙", x + PLAYER_RADIUS + 2, y - PLAYER_RADIUS + 4);
    }
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
