/**
 * Represents a WebSocket message envelope used across the application.
 * All messages sent to or received from the server use this shape.
 *
 * Example:
 *   { type: 'JOIN_ROOM', payload: { roomCode: 'RF-1234' } }
 *   { type: 'PLAYER_JOINED', payload: { username: 'alice' } }
 */
export interface WsMessage<T = unknown> {
  type: string;
  payload: T;
}
