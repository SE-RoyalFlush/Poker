export type GamePhase = 'waiting' | 'pre-flop' | 'flop' | 'turn' | 'river' | 'showdown';

export interface Card {
  rank: string;
  suit: string;
}

export interface PlayerSeat {
  seatIndex: number;
  playerId: number;
  username: string;
  chipCount: number;
  holeCards: Card[];
  isCurrentUser: boolean;
  isActive: boolean;
  isDealer: boolean;
}

export interface GameState {
  phase: GamePhase;
  communityCards: Card[];
  seats: PlayerSeat[];
  pot: number;
  currentBet: number;
  currentUserId: number;
}

export interface WinnerInfo {
  username: string;   // primary winner; "Split pot" when isSplit is true
  usernames: string[]; // all winners (len > 1 for a split pot)
  pot: number;
  isSplit: boolean;
}
