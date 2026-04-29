import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { CardComponent } from '../../shared/card/card.component';
import { AuthService } from '../../core/services/auth.service';
import { GameStateService } from '../../core/services/game-state.service';
import { GameState, Card, PlayerSeat, WinnerInfo } from '../../core/models/game-state.model';
import { GameControlsComponent, GameAction } from '../../features/table/game-controls/game-controls.component';

@Component({
  selector: 'app-table',
  standalone: true,
  imports: [CommonModule, CardComponent, GameControlsComponent],
  templateUrl: './table.html',
  styleUrl: './table.scss',
})
export class Table implements OnInit, OnDestroy {
  private readonly destroy$ = new Subject<void>();

  tableId = '';
  gameState: GameState | null = null;
  winner: WinnerInfo | null = null;

  constructor(
    private readonly route: ActivatedRoute,
    private readonly authService: AuthService,
    private readonly gameStateService: GameStateService,
  ) {}

  ngOnInit(): void {
    this.tableId = this.route.snapshot.paramMap.get('id') ?? '';

    this.gameStateService.gameState$
      .pipe(takeUntil(this.destroy$))
      .subscribe(state => { this.gameState = state; });

    this.gameStateService.winner$
      .pipe(takeUntil(this.destroy$))
      .subscribe(winner => { this.winner = winner; });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  get currentUsername(): string {
    return this.authService.getCurrentUser()?.username ?? '';
  }

  get visibleCommunityCards(): Card[] {
    const phase = this.gameState?.phase;
    const all = this.gameState?.communityCards ?? [];
    if (phase === 'flop')    return all.slice(0, 3);
    if (phase === 'turn')    return all.slice(0, 4);
    if (phase === 'river' || phase === 'showdown') return all.slice(0, 5);
    return [];
  }

  get communityCardSlots(): Array<Card | null> {
    const visible = this.visibleCommunityCards;
    return Array.from({ length: 5 }, (_, i) => visible[i] ?? null);
  }

  get currentUserSeat(): PlayerSeat | undefined {
    return this.gameState?.seats.find(s => s.isCurrentUser);
  }

  get opponentSeats(): PlayerSeat[] {
    return this.gameState?.seats.filter(s => !s.isCurrentUser) ?? [];
  }

  get holeCardSlots(): Array<Card | null> {
    const cards = this.currentUserSeat?.holeCards ?? [];
    return Array.from({ length: 2 }, (_, i) => cards[i] ?? null);
  }

  get isActivePlayer(): boolean {
    return this.currentUserSeat?.isActive ?? false;
  }

  get phasLabel(): string {
    const labels: Record<string, string> = {
      'waiting':  'Waiting',
      'pre-flop': 'Pre-Flop',
      'flop':     'Flop',
      'turn':     'Turn',
      'river':    'River',
      'showdown': 'Showdown',
    };
    return labels[this.gameState?.phase ?? 'waiting'] ?? 'Waiting';
  }

  dismissWinner(): void {
    this.winner = null;
  }

  get callAmount(): number {
    return this.gameState?.currentBet ?? 0;
  }

  get maxRaise(): number {
    return this.currentUserSeat?.chipCount ?? 0;
  }

  onGameAction(action: GameAction): void {
    this.gameStateService.sendAction(action.type, action.amount);
  }
}
