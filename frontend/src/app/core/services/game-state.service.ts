import { Injectable, OnDestroy } from '@angular/core';
import { BehaviorSubject, Observable, Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { Router } from '@angular/router';
import { WebSocketService } from './websocket.service';
import { SoundEffectsService } from './sound-effects.service';
import { GameState, GamePhase, Card, PlayerSeat, WinnerInfo } from '../models/game-state.model';
import {
  WsMessage,
  GameServerMessageType,
  GameStartedPayload,
  CardsDealtPayload,
  PlayerActionPayload,
  PhaseChangePayload,
  GameOverPayload,
} from '../models/ws-message.model';

const INITIAL_GAME_STATE: GameState = {
  phase: 'waiting',
  communityCards: [],
  seats: [],
  pot: 0,
  currentBet: 0,
  currentUserId: -1,
};

@Injectable({ providedIn: 'root' })
export class GameStateService implements OnDestroy {
  private readonly destroy$ = new Subject<void>();

  private readonly playersSubject = new BehaviorSubject<PlayerSeat[]>([]);
  private readonly communityCardsSubject = new BehaviorSubject<Card[]>([]);
  private readonly holeCardsSubject = new BehaviorSubject<Card[]>([]);
  private readonly potSubject = new BehaviorSubject<number>(0);
  private readonly activePlayerIdSubject = new BehaviorSubject<number>(-1);
  private readonly phaseSubject = new BehaviorSubject<GamePhase>('waiting');

  readonly players$: Observable<PlayerSeat[]> = this.playersSubject.asObservable();
  readonly communityCards$: Observable<Card[]> = this.communityCardsSubject.asObservable();
  readonly holeCards$: Observable<Card[]> = this.holeCardsSubject.asObservable();
  readonly pot$: Observable<number> = this.potSubject.asObservable();
  readonly activePlayerId$: Observable<number> = this.activePlayerIdSubject.asObservable();
  readonly phase$: Observable<GamePhase> = this.phaseSubject.asObservable();

  private readonly winnerSubject = new BehaviorSubject<WinnerInfo | null>(null);
  readonly winner$: Observable<WinnerInfo | null> = this.winnerSubject.asObservable();

  private readonly stateSubject = new BehaviorSubject<GameState>(INITIAL_GAME_STATE);
  readonly gameState$: Observable<GameState> = this.stateSubject.asObservable();

  constructor(
    private readonly wsService: WebSocketService,
    private readonly router: Router,
    private readonly soundEffects: SoundEffectsService,
  ) {
    this.wsService.messages$
      .pipe(takeUntil(this.destroy$))
      .subscribe(msg => this.handleMessage(msg));
  }

  private handleMessage(msg: WsMessage): void {
    switch (msg.type) {
      case GameServerMessageType.GAME_STARTED:
        this.onGameStarted(msg.payload as GameStartedPayload);
        break;
      case GameServerMessageType.CARDS_DEALT:
        this.onCardsDealt(msg.payload as CardsDealtPayload);
        break;
      case GameServerMessageType.PLAYER_ACTION:
        this.onPlayerAction(msg.payload as PlayerActionPayload);
        break;
      case GameServerMessageType.PHASE_CHANGE:
        this.onPhaseChange(msg.payload as PhaseChangePayload);
        break;
      case GameServerMessageType.GAME_OVER:
        this.onGameOver(msg.payload as GameOverPayload);
        break;
    }
  }

  private onGameStarted(payload: GameStartedPayload): void {
    this.winnerSubject.next(null);
    this.playersSubject.next(payload.seats);
    this.phaseSubject.next(payload.phase);
    this.potSubject.next(payload.pot);
    this.activePlayerIdSubject.next(payload.activePlayerId);
    this.communityCardsSubject.next([]);
    this.holeCardsSubject.next([]);
    this.syncState({
      phase: payload.phase,
      seats: payload.seats,
      pot: payload.pot,
      currentBet: payload.currentBet,
      currentUserId: payload.currentUserId,
      communityCards: [],
    });
    this.router.navigate(['/table', payload.tableId]);
  }

  private onCardsDealt(payload: CardsDealtPayload): void {
    this.playersSubject.next(payload.seats);
    this.holeCardsSubject.next(payload.holeCards);
    this.syncState({ seats: payload.seats });
    this.soundEffects.playCardFlip();
  }

  private onPlayerAction(payload: PlayerActionPayload): void {
    this.playersSubject.next(payload.seats);
    this.potSubject.next(payload.pot);
    this.activePlayerIdSubject.next(payload.activePlayerId);
    this.syncState({
      seats: payload.seats,
      pot: payload.pot,
      currentBet: payload.currentBet,
    });
  }

  private onPhaseChange(payload: PhaseChangePayload): void {
    this.phaseSubject.next(payload.phase);
    this.communityCardsSubject.next(payload.communityCards);
    this.potSubject.next(payload.pot);
    this.activePlayerIdSubject.next(payload.activePlayerId);
    this.syncState({
      phase: payload.phase,
      communityCards: payload.communityCards,
      pot: payload.pot,
      currentBet: payload.currentBet,
    });
  }

  private onGameOver(payload: GameOverPayload): void {
    this.playersSubject.next(payload.seats);
    this.potSubject.next(payload.pot);
    this.phaseSubject.next('showdown');
    this.syncState({
      seats: payload.seats,
      pot: payload.pot,
      phase: 'showdown',
    });
    const winnerIds = payload.winnerIds ?? [payload.winnerId];
    const isSplit = winnerIds.length > 1;
    const winnerSeats = winnerIds.map(id => payload.seats.find(s => s.playerId === id)).filter(Boolean) as import('../models/game-state.model').PlayerSeat[];
    if (winnerSeats.length > 0) {
      this.winnerSubject.next({
        username: isSplit ? 'Split pot' : winnerSeats[0].username,
        usernames: winnerSeats.map(s => s.username),
        pot: payload.pot,
        isSplit,
      });
    } else {
      this.winnerSubject.next(null);
    }
  }

  private syncState(partial: Partial<GameState>): void {
    this.stateSubject.next({ ...this.stateSubject.value, ...partial });
  }

  /**
   * Returns the active player using _.find(players, { playerId: activePlayerId }).
   */
  getActivePlayer(): PlayerSeat | undefined {
    const activeId = this.activePlayerIdSubject.value;
    return this.playersSubject.value.find(s => s.playerId === activeId);
  }

  sendAction(type: string, amount?: number): void {
    const payload = amount !== undefined ? { amount } : {};
    this.wsService.sendMessage(type, payload);
  }

  getSnapshot(): GameState {
    return this.stateSubject.value;
  }

  patchState(partial: Partial<GameState>): void {
    this.syncState(partial);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
