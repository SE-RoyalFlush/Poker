/**
 * Player model representing a participant in a lobby or game room.
 * Received from the server via PLAYER_JOINED WebSocket messages.
 */
export interface Player {
  id: number;
  username: string;
  isHost: boolean;
  isReady: boolean;
}
