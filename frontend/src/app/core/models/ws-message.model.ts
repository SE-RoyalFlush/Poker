import { Player } from './player.model';
import { Card, GamePhase, PlayerSeat } from './game-state.model';

/**
 * Shared wire format for all WebSocket messages.
 * Every message sent to or received from the server uses this envelope.
 *
 *   { type: 'JOIN_ROOM', payload: { roomCode: 'AB12CD' } }
 *   { type: 'PLAYER_JOINED', payload: { id: 1, username: 'alice', ... } }
 */
/** Client → Server event type constants. */
export const ClientMessageType = {
  JOIN_ROOM: 'JOIN_ROOM',
  LEAVE_ROOM: 'LEAVE_ROOM',
  TOGGLE_READY: 'TOGGLE_READY',
  START_GAME: 'START_GAME',
} as const;

export type ClientMessageType = (typeof ClientMessageType)[keyof typeof ClientMessageType];

/** Server → Client event type constants. */
export const ServerMessageType = {
  PLAYER_JOINED: 'PLAYER_JOINED',
  PLAYER_LEFT: 'PLAYER_LEFT',
  PLAYER_UPDATE: 'PLAYER_UPDATE',
  ROOM_STATE: 'ROOM_STATE',
  CHAT_MESSAGE: 'CHAT_MESSAGE',
  ERROR: 'ERROR',
} as const;

export type ServerMessageType = (typeof ServerMessageType)[keyof typeof ServerMessageType];

/** Server → Client game event type constants. */
export const GameServerMessageType = {
  GAME_STARTED: 'GAME_STARTED',
  CARDS_DEALT: 'CARDS_DEALT',
  PLAYER_ACTION: 'PLAYER_ACTION',
  PHASE_CHANGE: 'PHASE_CHANGE',
  GAME_OVER: 'GAME_OVER',
} as const;

export type GameServerMessageType = (typeof GameServerMessageType)[keyof typeof GameServerMessageType];

/** Client → Server game action type constants. */
export const GameClientMessageType = {
  CHECK: 'CHECK',
  CALL: 'CALL',
  RAISE: 'RAISE',
  FOLD: 'FOLD',
} as const;

export type GameClientMessageType = (typeof GameClientMessageType)[keyof typeof GameClientMessageType];

export type MessageType = ClientMessageType | ServerMessageType | GameServerMessageType | GameClientMessageType;

export interface WsMessage<TPayload = unknown, TType extends MessageType = MessageType> {
  type: TType;
  payload: TPayload;
}

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

export type JoinRoomMessage = WsMessage<JoinRoomPayload, typeof ClientMessageType.JOIN_ROOM>;
export type LeaveRoomMessage = WsMessage<Record<string, never>, typeof ClientMessageType.LEAVE_ROOM>;
export type ToggleReadyMessage = WsMessage<Record<string, never>, typeof ClientMessageType.TOGGLE_READY>;
export type PlayerJoinedMessage = WsMessage<Player, typeof ServerMessageType.PLAYER_JOINED>;
export type PlayerLeftMessage = WsMessage<PlayerLeftPayload, typeof ServerMessageType.PLAYER_LEFT>;
export type PlayerUpdateMessage = WsMessage<Player, typeof ServerMessageType.PLAYER_UPDATE>;
export type RoomStateMessage = WsMessage<RoomStatePayload, typeof ServerMessageType.ROOM_STATE>;
export type ErrorMessage = WsMessage<ErrorPayload, typeof ServerMessageType.ERROR>;

// ---------------------------------------------------------------------------
// Game event payloads (Server → Client)
// ---------------------------------------------------------------------------

/** Sent when a game round begins; triggers navigation from lobby to table. */
export interface GameStartedPayload {
  tableId: string;
  seats: PlayerSeat[];
  phase: GamePhase;
  pot: number;
  currentBet: number;
  activePlayerId: number;
  currentUserId: number;
}

/** Sent after the deal; contains updated seats with hole cards for the current user. */
export interface CardsDealtPayload {
  seats: PlayerSeat[];
  holeCards: Card[];
}

/** Broadcast after a player acts; contains the action and updated table state. */
export interface PlayerActionPayload {
  playerId: number;
  action: string;
  amount?: number;
  pot: number;
  currentBet: number;
  seats: PlayerSeat[];
  activePlayerId: number;
}

/** Sent when betting moves to the next street; includes new community cards. */
export interface PhaseChangePayload {
  phase: GamePhase;
  communityCards: Card[];
  pot: number;
  currentBet: number;
  activePlayerId: number;
}

/** Sent when the round ends; identifies the winner and final chip counts. */
export interface GameOverPayload {
  winnerId: number;
  pot: number;
  seats: PlayerSeat[];
}

// ---------------------------------------------------------------------------
// Game action payload (Client → Server)
// ---------------------------------------------------------------------------

export interface RaisePayload {
  amount: number;
}

// ---------------------------------------------------------------------------
// Typed game message aliases
// ---------------------------------------------------------------------------

export type GameStartedMessage = WsMessage<GameStartedPayload, typeof GameServerMessageType.GAME_STARTED>;
export type CardsDealtMessage = WsMessage<CardsDealtPayload, typeof GameServerMessageType.CARDS_DEALT>;
export type PlayerActionMessage = WsMessage<PlayerActionPayload, typeof GameServerMessageType.PLAYER_ACTION>;
export type PhaseChangeMessage = WsMessage<PhaseChangePayload, typeof GameServerMessageType.PHASE_CHANGE>;
export type GameOverMessage = WsMessage<GameOverPayload, typeof GameServerMessageType.GAME_OVER>;
export type CheckMessage = WsMessage<Record<string, never>, typeof GameClientMessageType.CHECK>;
export type CallMessage = WsMessage<Record<string, never>, typeof GameClientMessageType.CALL>;
export type RaiseMessage = WsMessage<RaisePayload, typeof GameClientMessageType.RAISE>;
export type FoldMessage = WsMessage<Record<string, never>, typeof GameClientMessageType.FOLD>;
