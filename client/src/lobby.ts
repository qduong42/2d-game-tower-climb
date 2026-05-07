export interface LobbyResult {
  roomCode: string;
  name: string;
  color: string;
  isPrivate: boolean;
}

const COLORS = ["#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6"];
const LS_NAME = "tc_name";
const LS_COLOR = "tc_color";
const LS_ROOM = "tc_room";

function randomCode(): string {
  return Array.from({ length: 4 }, () => "ABCDEFGHJKLMNPQRSTUVWXYZ"[Math.floor(Math.random() * 23)!]).join("");
}

function savedCode(): string {
  const fromUrl = location.pathname.startsWith("/r/")
    ? location.pathname.slice(3).trim().toUpperCase()
    : "";
  return fromUrl || localStorage.getItem(LS_ROOM) || "";
}

/** Minimal room entry returned by GET /rooms */
interface OpenRoom {
  code: string;
  players: number;
}

export function showLobby(container: HTMLElement): Promise<LobbyResult> {
  return new Promise((resolve) => {
    const initCode  = savedCode();
    const initName  = localStorage.getItem(LS_NAME) || "";
    const initColor = localStorage.getItem(LS_COLOR) || COLORS[0]!;

    container.innerHTML = `
      <div style="max-width:360px;margin:auto;padding:2rem;font-family:monospace;color:#e0e0e0">
        <h1 style="margin-bottom:1rem;color:#ffd700;letter-spacing:0.05em">Turbine Repair</h1>

        <label style="display:block;margin-bottom:0.25rem">Room code</label>
        <div style="display:flex;gap:0.5rem;margin-bottom:0.5rem">
          <input id="room-code" value="${initCode}" placeholder="ABCD"
            style="flex:1;padding:0.5rem;font:inherit;background:#111;color:#e0e0e0;border:1px solid #444" />
          <button id="create-btn" style="padding:0.5rem 0.75rem;background:#333;color:#ffd700;border:1px solid #555;font:inherit;cursor:pointer">
            New
          </button>
        </div>
        <div style="display:flex;align-items:center;gap:0.5rem;margin-bottom:1rem">
          <input type="checkbox" id="private-check" style="width:1rem;height:1rem;accent-color:#ffd700;cursor:pointer" />
          <label for="private-check" style="cursor:pointer;font-size:0.9rem;color:#aaa">Private room</label>
        </div>

        <label style="display:block;margin-bottom:0.25rem">Your name</label>
        <input id="player-name" value="${initName}" placeholder="alice" required
          style="display:block;width:100%;box-sizing:border-box;margin-bottom:0.25rem;padding:0.5rem;font:inherit;background:#111;color:#e0e0e0;border:1px solid #444" />
        <span id="name-error" style="color:#e74c3c;font-size:0.8rem;margin-bottom:0.75rem;display:block;min-height:1.2em">&nbsp;</span>

        <label style="display:block;margin-bottom:0.25rem">Colour</label>
        <div id="color-picker" style="display:flex;gap:0.5rem;margin-bottom:1.25rem">
          ${COLORS.map((c) =>
            `<button data-color="${c}" style="background:${c};width:2rem;height:2rem;border:${c === initColor ? "3px solid #ffd700" : "3px solid transparent"};cursor:pointer;border-radius:3px"></button>`
          ).join("")}
        </div>

        <button id="join-btn" style="width:100%;padding:0.75rem;background:#1a6fa8;color:#fff;border:none;font:inherit;cursor:pointer;letter-spacing:0.05em">
          Join / Create
        </button>

        <hr style="border-color:#333;margin:1.5rem 0" />

        <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.75rem">
          <span style="color:#ffd700;font-size:0.9rem;letter-spacing:0.05em">OPEN ROOMS</span>
          <button id="refresh-btn" style="padding:0.3rem 0.6rem;background:#333;color:#ffd700;border:1px solid #555;font:inherit;font-size:0.8rem;cursor:pointer">
            Refresh
          </button>
        </div>
        <div id="rooms-list" style="font-size:0.85rem;color:#aaa">Loading...</div>
      </div>
    `;

    // ---- helpers ----

    function validateName(): string | null {
      const nameInput = container.querySelector<HTMLInputElement>("#player-name")!;
      const name = nameInput.value.trim();
      if (!name) {
        const err = container.querySelector<HTMLSpanElement>("#name-error")!;
        err.textContent = "Name is required";
        nameInput.style.border = "1px solid #e74c3c";
        nameInput.focus();
        return null;
      }
      return name;
    }

    function doJoin(code: string, isPrivate: boolean): void {
      const name = validateName();
      if (!name) return;
      localStorage.setItem(LS_NAME, name);
      localStorage.setItem(LS_COLOR, selectedColor);
      localStorage.setItem(LS_ROOM, code);
      history.replaceState(null, "", `/r/${code}`);
      container.innerHTML = "";
      resolve({ roomCode: code, name, color: selectedColor, isPrivate });
    }

    // ---- room code input ----

    const codeInput = container.querySelector<HTMLInputElement>("#room-code")!;
    codeInput.addEventListener("input", () => {
      const pos = codeInput.selectionStart;
      codeInput.value = codeInput.value.toUpperCase();
      codeInput.setSelectionRange(pos, pos);
    });

    container.querySelector("#create-btn")!.addEventListener("click", () => {
      codeInput.value = randomCode();
    });

    // ---- colour picker ----

    let selectedColor = initColor;
    container.querySelectorAll<HTMLButtonElement>("[data-color]").forEach((btn) => {
      btn.addEventListener("click", () => {
        container.querySelectorAll<HTMLButtonElement>("[data-color]").forEach((b) => {
          b.style.border = "3px solid transparent";
        });
        btn.style.border = "3px solid #ffd700";
        selectedColor = btn.dataset.color!;
      });
    });

    // ---- join button ----

    container.querySelector("#join-btn")!.addEventListener("click", () => {
      const code = codeInput.value.trim().toUpperCase();
      if (!code) {
        alert("Enter an existing room code or hit New to generate one!");
        codeInput.focus();
        return;
      }
      const isPrivate = (container.querySelector<HTMLInputElement>("#private-check")!).checked;
      doJoin(code, isPrivate);
    });

    // ---- room browser ----

    function renderRooms(rooms: OpenRoom[]): void {
      const listEl = container.querySelector<HTMLDivElement>("#rooms-list")!;
      if (rooms.length === 0) {
        listEl.innerHTML = `<span style="color:#555">No open rooms available</span>`;
        return;
      }
      listEl.innerHTML = rooms
        .map(
          (r) => `
          <div style="display:flex;justify-content:space-between;align-items:center;padding:0.4rem 0.5rem;margin-bottom:0.3rem;background:#111;border:1px solid #333">
            <span style="color:#e0e0e0;letter-spacing:0.08em">${r.code}</span>
            <span style="color:#aaa;font-size:0.8rem">${r.players}/3 players</span>
            <button data-join-code="${r.code}"
              style="padding:0.25rem 0.6rem;background:#1a6fa8;color:#fff;border:none;font:inherit;font-size:0.8rem;cursor:pointer">
              Join
            </button>
          </div>`
        )
        .join("");

      listEl.querySelectorAll<HTMLButtonElement>("[data-join-code]").forEach((btn) => {
        btn.addEventListener("click", () => {
          const code = btn.dataset.joinCode!;
          codeInput.value = code;
          // Joining a listed room is always public (it's already created).
          doJoin(code, false);
        });
      });
    }

    async function fetchRooms(): Promise<void> {
      const listEl = container.querySelector<HTMLDivElement>("#rooms-list");
      if (!listEl) return; // lobby was already dismissed
      listEl.textContent = "Loading...";
      try {
        const res = await fetch("/rooms");
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const rooms = (await res.json()) as OpenRoom[];
        if (!container.querySelector("#rooms-list")) return; // dismissed while fetching
        renderRooms(rooms);
      } catch {
        if (container.querySelector("#rooms-list")) {
          container.querySelector<HTMLDivElement>("#rooms-list")!.textContent =
            "Could not load rooms";
        }
      }
    }

    container.querySelector("#refresh-btn")!.addEventListener("click", () => fetchRooms());
    void fetchRooms();
  });
}
