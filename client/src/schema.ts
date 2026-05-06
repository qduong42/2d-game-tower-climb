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

// --- Server → Client ---

export interface WelcomePayload {
  yourId: string;
  roomCode: string;
  tickRate: number;
}

export interface PlayerState {
  id: string;
  x: number;
  y: number;
  color: string;
  name: string;
}

export interface SnapshotPayload {
  tick: number;
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
