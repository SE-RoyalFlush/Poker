import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { catchError, of } from 'rxjs';

import { CardComponent } from '../../shared/card/card.component';
import { GameState, Card, PlayerSeat, WinnerInfo, GamePhase } from '../../core/models/game-state.model';
import { API_URL } from '../../core/config/endpoints';

const PHASES: GamePhase[] = ['pre-flop', 'flop', 'turn', 'river', 'showdown'];

const POT_BY_PHASE: Record<string, number> = {
  'pre-flop': 450,
  'flop': 850,
  'turn': 1250,
  'river': 1800,
  'showdown': 1800,
};

const FALLBACK_STATE: GameState = {
  phase: 'pre-flop',
  communityCards: [
    { rank: 'A', suit: 'spades' },
    { rank: 'K', suit: 'hearts' },
    { rank: '7', suit: 'diamonds' },
    { rank: '2', suit: 'clubs' },
    { rank: 'J', suit: 'spades' },
  ],
  seats: [
    {
      seatIndex: 0, playerId: 1, username: 'You', chipCount: 1500,
      holeCards: [{ rank: 'A', suit: 'hearts' }, { rank: 'A', suit: 'clubs' }],
      isCurrentUser: true, isActive: true, isDealer: false,
    },
    {
      seatIndex: 1, playerId: 2, username: 'Alice', chipCount: 2000,
      holeCards: [{ rank: 'K', suit: 'spades' }, { rank: 'Q', suit: 'hearts' }],
      isCurrentUser: false, isActive: false, isDealer: true,
    },
    {
      seatIndex: 2, playerId: 3, username: 'Bob', chipCount: 800,
      holeCards: [{ rank: '9', suit: 'clubs' }, { rank: '8', suit: 'diamonds' }],
      isCurrentUser: false, isActive: false, isDealer: false,
    },
    {
      seatIndex: 3, playerId: 4, username: 'Carol', chipCount: 3200,
      holeCards: [{ rank: '5', suit: 'hearts' }, { rank: '5', suit: 'spades' }],
      isCurrentUser: false, isActive: false, isDealer: false,
    },
  ],
  pot: 450,
  currentBet: 50,
  currentUserId: 1,
};

@Component({
  selector: 'app-table-demo',
  standalone: true,
  imports: [CommonModule, CardComponent],
  templateUrl: './table-demo.html',
  styleUrls: ['../table/table.scss', './table-demo.scss'],
})
export class TableDemo implements OnInit, OnDestroy {
  baseState: GameState = FALLBACK_STATE;
  phase: GamePhase = 'pre-flop';
  winner: WinnerInfo | null = null;
  autoPlay = true;

  private phaseIndex = 0;
  private timer: ReturnType<typeof setInterval> | null = null;

  constructor(private readonly http: HttpClient) {}

  ngOnInit(): void {
    this.http.get<GameState>(`${API_URL}/demo/game-state`).pipe(
      catchError(() => of(FALLBACK_STATE)),
    ).subscribe(state => {
      this.baseState = state;
      this.startAutoPlay();
    });
  }

  ngOnDestroy(): void {
    this.clearTimer();
  }

  get currentState(): GameState {
    return { ...this.baseState, phase: this.phase, pot: POT_BY_PHASE[this.phase] };
  }

  get visibleCommunityCards(): Card[] {
    const all = this.baseState.communityCards ?? [];
    if (this.phase === 'flop')    return all.slice(0, 3);
    if (this.phase === 'turn')    return all.slice(0, 4);
    if (this.phase === 'river' || this.phase === 'showdown') return all.slice(0, 5);
    return [];
  }

  get communityCardSlots(): Array<Card | null> {
    const visible = this.visibleCommunityCards;
    return Array.from({ length: 5 }, (_, i) => visible[i] ?? null);
  }

  get currentUserSeat(): PlayerSeat | undefined {
    return this.baseState.seats.find(s => s.isCurrentUser);
  }

  get opponentSeats(): PlayerSeat[] {
    return this.baseState.seats.filter(s => !s.isCurrentUser);
  }

  get pot(): number {
    return POT_BY_PHASE[this.phase];
  }

  get phaseLabel(): string {
    const labels: Record<string, string> = {
      'pre-flop': 'Pre-Flop',
      'flop': 'Flop',
      'turn': 'Turn',
      'river': 'River',
      'showdown': 'Showdown',
    };
    return labels[this.phase] ?? 'Pre-Flop';
  }

  get isLastPhase(): boolean {
    return this.phaseIndex >= PHASES.length - 1;
  }

  advancePhase(): void {
    if (this.winner) { this.reset(); return; }

    if (this.phaseIndex < PHASES.length - 1) {
      this.phaseIndex++;
      this.phase = PHASES[this.phaseIndex];
    } else {
      this.showWinner();
    }
  }

  reset(): void {
    this.winner = null;
    this.phaseIndex = 0;
    this.phase = 'pre-flop';
    if (this.autoPlay) this.startAutoPlay();
  }

  toggleAutoPlay(): void {
    this.autoPlay = !this.autoPlay;
    if (this.autoPlay) { this.startAutoPlay(); } else { this.clearTimer(); }
  }

  dismissWinner(): void {
    this.reset();
  }

  private showWinner(): void {
    this.clearTimer();
    this.winner = { username: 'You', usernames: ['You'], pot: POT_BY_PHASE['showdown'], isSplit: false };
  }

  private startAutoPlay(): void {
    this.clearTimer();
    this.timer = setInterval(() => this.advancePhase(), 3000);
  }

  private clearTimer(): void {
    if (this.timer) { clearInterval(this.timer); this.timer = null; }
  }
}
