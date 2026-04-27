import { Player } from './player.model';

/**
 * Shared wire format for all WebSocket messages.
 * Every message sent to or received from the server uses this envelope.
 *
 *   { type: 'JOIN_ROOM', payload: { roomCode: 'AB12CD' } }
 *   { type: 'PLAYER_JOINED', payload: { id: 1, username: 'alice', ... } }
 */
export interface WsMessage<T = unknown> {
  type: string;
  payload: T;
}

/** Client → Server event type constants. */
export const ClientMessageType = {
  JOIN_ROOM: 'JOIN_ROOM',
  LEAVE_ROOM: 'LEAVE_ROOM',
  TOGGLE_READY: 'TOGGLE_READY',
} as const;

export type ClientMessageType = (typeof ClientMessageType)[keyof typeof ClientMessageType];

/** Server → Client event type constants. */
export const ServerMessageType = {
  PLAYER_JOINED: 'PLAYER_JOINED',
  PLAYER_LEFT: 'PLAYER_LEFT',
  PLAYER_UPDATE: 'PLAYER_UPDATE',
  ROOM_STATE: 'ROOM_STATE',
  ERROR: 'ERROR',
} as const;

export type ServerMessageType = (typeof ServerMessageType)[keyof typeof ServerMessageType];

// ---------------------------------------------------------------------------
// Payload schemas
// ---------------------------------------------------------------------------

/** Sent by the client to join a room. */
export interface JoinRoomPayload {
  roomCode: string;
}

/** Broadcast by the server when a player leaves or disconnects. */
export interface PlayerLeftPayload {
  id: number;
}

/** Sent to a newly joined client with the current room roster. */
export interface RoomStatePayload {
  roomCode: string;
  players: Player[];
}

/** Sent by the server when a client message cannot be processed. */
export interface ErrorPayload {
  code: string;
  message: string;
}

// ---------------------------------------------------------------------------
// Typed message aliases
// ---------------------------------------------------------------------------

export type JoinRoomMessage = WsMessage<JoinRoomPayload>;
export type LeaveRoomMessage = WsMessage<Record<string, never>>;
export type ToggleReadyMessage = WsMessage<Record<string, never>>;
export type PlayerJoinedMessage = WsMessage<Player>;
export type PlayerLeftMessage = WsMessage<PlayerLeftPayload>;
export type PlayerUpdateMessage = WsMessage<Player>;
export type RoomStateMessage = WsMessage<RoomStatePayload>;
export type ErrorMessage = WsMessage<ErrorPayload>;
