import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { GameState } from '../models/game-state.model';

const INITIAL_GAME_STATE: GameState = {
  phase: 'waiting',
  communityCards: [],
  seats: [],
  pot: 0,
  currentBet: 0,
  currentUserId: -1,
};

@Injectable({ providedIn: 'root' })
export class GameStateService {
  private readonly stateSubject = new BehaviorSubject<GameState>(INITIAL_GAME_STATE);
  readonly gameState$: Observable<GameState> = this.stateSubject.asObservable();

  getSnapshot(): GameState {
    return this.stateSubject.value;
  }

  patchState(partial: Partial<GameState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partial });
  }
}
