export interface LobbyResult {
  roomCode: string;
  name: string;
  color: string;
}

const COLORS = ["#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6"];

function randomCode(): string {
  return Array.from({ length: 4 }, () => "ABCDEFGHJKLMNPQRSTUVWXYZ"[Math.floor(Math.random() * 23)!]).join("");
}

export function showLobby(container: HTMLElement): Promise<LobbyResult> {
  return new Promise((resolve) => {
    const codeFromUrl = location.pathname.startsWith("/r/")
      ? location.pathname.slice(3).trim().toUpperCase()
      : "";

    container.innerHTML = `
      <div style="max-width:320px;margin:auto;padding:2rem">
        <h1 style="margin-bottom:1rem">Tower Climb</h1>
        <label>Room code
          <div style="display:flex;gap:0.5rem;margin:0.25rem 0 1rem">
            <input id="room-code" value="${codeFromUrl}" placeholder="ABCD"
              style="flex:1;padding:0.5rem;font:inherit;background:#333;color:#fff;border:1px solid #666" />
            <button id="create-btn" style="padding:0.5rem 0.75rem;background:#555;color:#fff;border:none;font:inherit;cursor:pointer">
              New
            </button>
          </div>
        </label>
        <label>Your name
          <input id="player-name" placeholder="alice" required
            style="display:block;width:100%;margin:0.25rem 0 0.25rem;padding:0.5rem;font:inherit;background:#333;color:#fff;border:1px solid #666" />
          <span id="name-error" style="display:none;color:#e74c3c;font-size:0.8rem;margin-bottom:0.75rem;display:block">&nbsp;</span>
        </label>
        <label>Colour
          <div id="color-picker" style="display:flex;gap:0.5rem;margin:0.25rem 0 1rem">
            ${COLORS.map((c, i) =>
              `<button data-color="${c}" style="background:${c};width:2rem;height:2rem;border:${i === 0 ? "3px solid #fff" : "3px solid transparent"};cursor:pointer"></button>`
            ).join("")}
          </div>
        </label>
        <button id="join-btn" style="width:100%;padding:0.75rem;background:#3498db;color:#fff;border:none;font:inherit;cursor:pointer">
          Join
        </button>
      </div>
    `;

    const codeInput = container.querySelector<HTMLInputElement>("#room-code")!;
    codeInput.addEventListener("input", () => {
      const pos = codeInput.selectionStart;
      codeInput.value = codeInput.value.toUpperCase();
      codeInput.setSelectionRange(pos, pos);
    });

    container.querySelector("#create-btn")!.addEventListener("click", () => {
      codeInput.value = randomCode();
    });

    let selectedColor = COLORS[0]!;
    container.querySelectorAll<HTMLButtonElement>("[data-color]").forEach((btn) => {
      btn.addEventListener("click", () => {
        container.querySelectorAll<HTMLButtonElement>("[data-color]").forEach((b) => {
          b.style.border = "3px solid transparent";
        });
        btn.style.border = "3px solid #fff";
        selectedColor = btn.dataset.color!;
      });
    });

    container.querySelector("#join-btn")!.addEventListener("click", () => {
      const code = codeInput.value.trim().toUpperCase();
      if (!code) {
        alert("Enter existing room code or hit NEW to generate a room code!");
        codeInput.focus();
        return;
      }

      const nameInput = container.querySelector<HTMLInputElement>("#player-name")!;
      const name = nameInput.value.trim();
      if (!name) {
        const err = container.querySelector<HTMLSpanElement>("#name-error")!;
        err.textContent = "Name is required";
        err.style.display = "block";
        nameInput.style.border = "1px solid #e74c3c";
        nameInput.focus();
        return;
      }

      container.innerHTML = "";
      resolve({ roomCode: code, name, color: selectedColor });
    });
  });
}
