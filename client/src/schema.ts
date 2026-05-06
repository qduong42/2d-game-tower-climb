// Mirror of internal/schema/messages.go — update both files together.

export enum MsgType {
  Welcome  = "welcome",
  Snapshot = "snapshot",
  Event    = "event",
  Join     = "join",
  Input    = "input",
}

// Envelope wraps every message on the wire.
export interface Envelope {
  type: MsgType;
  payload: unknown;
}

export type Role = "base" | "climber";
export type Phase = "waiting" | "playing" | "won";
export type ToolType = "" | "wrench" | "hammer";

// --- Server → Client ---

export interface WelcomePayload {
  yourId: string;
  roomCode: string;
  tickRate: number;
  color: string; // actual assigned color (may differ if chosen color was taken)
}

export interface PlayerState {
  id: string;
  color: string;
  name: string;
  role: Role;
  climberIndex: number;    // 0 or 1 for climbers; -1 for base operator
  platform: number;        // 0=ground … NumPlatforms-1=top
  tool: ToolType;          // tool carried by this player ("" = none)
  heldTools: ToolType[];   // BASE inventory (empty for climbers)
  selectedTool: ToolType;  // BASE: which tool is queued to pass next
}

export interface SnapshotPayload {
  tick: number;
  phase: Phase;
  players: PlayerState[];
}

export type EventType = "join" | "leave" | "error";

export interface EventPayload {
  eventType: EventType;
  playerId?: string;
  message?: string;
}

// --- Client → Server ---

export interface JoinPayload {
  roomCode: string;
  name: string;
  color: string;
}

export interface InputKeys {
  up: boolean;
  down: boolean;
  left: boolean;
  right: boolean;
  space: boolean;
}

export interface MouseState {
  x: number;
  y: number;
  click: boolean;
}

export interface InputPayload {
  tick: number;
  keys: InputKeys;
  mouse?: MouseState;
}
