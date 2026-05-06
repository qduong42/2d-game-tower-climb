export class MenuOverlay {
  private el: HTMLElement;
  private visible = false;

  constructor(private container: HTMLElement, private onLeave: () => void) {
    this.el = document.createElement("div");
    this.el.style.cssText = `
      display:none;position:fixed;inset:0;background:rgba(0,0,0,0.7);
      justify-content:center;align-items:center;z-index:100;
    `;
    this.el.innerHTML = `
      <div style="background:#222;padding:2rem;min-width:200px;text-align:center">
        <p id="menu-room-code" style="margin-bottom:1rem;font-size:1.2rem"></p>
        <button id="copy-btn" style="display:block;width:100%;margin-bottom:0.5rem;padding:0.5rem;background:#555;color:#fff;border:none;font:inherit;cursor:pointer">
          Copy invite link
        </button>
        <button id="leave-btn" style="display:block;width:100%;padding:0.5rem;background:#c0392b;color:#fff;border:none;font:inherit;cursor:pointer">
          Leave room
        </button>
      </div>
    `;
    container.appendChild(this.el);

    this.el.querySelector("#leave-btn")!.addEventListener("click", () => {
      this.hide();
      this.onLeave();
    });

    this.el.querySelector("#copy-btn")!.addEventListener("click", () => {
      navigator.clipboard.writeText(location.href).catch(() => {});
    });

    window.addEventListener("keydown", (e) => {
      if (e.key === "Escape") this.toggle();
    });
  }

  show(roomCode: string): void {
    (this.el.querySelector("#menu-room-code") as HTMLElement).textContent = `Room: ${roomCode}`;
    this.el.style.display = "flex";
    this.visible = true;
  }

  hide(): void {
    this.el.style.display = "none";
    this.visible = false;
  }

  toggle(): void {
    if (this.visible) this.hide(); else this.show("");
  }
}
