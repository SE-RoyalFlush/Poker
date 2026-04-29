import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { BehaviorSubject } from 'rxjs';

import { Table } from './table';
import { GameStateService } from '../../core/services/game-state.service';
import { AuthService } from '../../core/services/auth.service';
import { GameState, GamePhase, PlayerSeat, Card, WinnerInfo } from '../../core/models/game-state.model';

// ── Helpers ──────────────────────────────────────────────────────────────────

function makeState(overrides: Partial<GameState> = {}): GameState {
  return {
    phase: 'waiting',
    communityCards: [],
    seats: [],
    pot: 0,
    currentBet: 0,
    currentUserId: 1,
    ...overrides,
  };
}

function buildSeats(count: number, currentUserId: number): PlayerSeat[] {
  return Array.from({ length: count }, (_, i) => ({
    seatIndex: i,
    playerId: i + 1,
    username: i + 1 === currentUserId ? 'ace' : `player-${i + 1}`,
    chipCount: 1000,
    holeCards: [],
    isCurrentUser: i + 1 === currentUserId,
    isActive: false,
    isDealer: i === 0,
  }));
}

const FIVE_CARDS: Card[] = [
  { rank: 'A', suit: 'spades' },
  { rank: 'K', suit: 'hearts' },
  { rank: 'Q', suit: 'diamonds' },
  { rank: 'J', suit: 'clubs' },
  { rank: '10', suit: 'spades' },
];

// ── Suite ─────────────────────────────────────────────────────────────────────

describe('Table', () => {
  let component: Table;
  let fixture: ComponentFixture<Table>;
  let stateSubject: BehaviorSubject<GameState>;
  let winnerSubject: BehaviorSubject<WinnerInfo | null>;
  let gssSpy: jasmine.SpyObj<GameStateService>;
  let authSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    stateSubject = new BehaviorSubject<GameState>(makeState());
    winnerSubject = new BehaviorSubject<WinnerInfo | null>(null);

    gssSpy = jasmine.createSpyObj<GameStateService>('GameStateService', ['getSnapshot', 'patchState', 'sendAction']);
    (gssSpy as unknown as { gameState$: unknown }).gameState$ = stateSubject.asObservable();
    (gssSpy as unknown as { winner$: unknown }).winner$ = winnerSubject.asObservable();

    authSpy = jasmine.createSpyObj<AuthService>('AuthService', ['getCurrentUser']);
    authSpy.getCurrentUser.and.returnValue({ id: 1, username: 'ace' });

    await TestBed.configureTestingModule({
      imports: [Table],
      providers: [
        provideNoopAnimations(),
        { provide: GameStateService, useValue: gssSpy },
        { provide: AuthService, useValue: authSpy },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id: 'table-42' }) } },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(Table);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  // ── Creation ──────────────────────────────────────────────────────────────

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should read tableId from route params', () => {
    expect(component.tableId).toBe('table-42');
  });

  it('should unsubscribe from gameState$ on destroy', () => {
    const phaseBefore = component.gameState?.phase;
    component.ngOnDestroy();
    stateSubject.next(makeState({ phase: 'river' }));
    expect(component.gameState?.phase).toBe(phaseBefore);
  });

  // ── Pot display ───────────────────────────────────────────────────────────

  describe('pot display', () => {
    it('renders $0 initially', () => {
      const el: HTMLElement = fixture.nativeElement.querySelector('[data-cy="pot-display"]');
      expect(el.textContent).toContain('0');
    });

    it('updates reactively when pot$ emits new value', () => {
      stateSubject.next(makeState({ pot: 750 }));
      fixture.detectChanges();
      const el: HTMLElement = fixture.nativeElement.querySelector('[data-cy="pot-display"]');
      expect(el.textContent).toContain('750');
    });
  });

  // ── Phase badge ───────────────────────────────────────────────────────────

  describe('phase badge', () => {
    const phases: Array<{ phase: GamePhase; label: string }> = [
      { phase: 'waiting',  label: 'Waiting'   },
      { phase: 'pre-flop', label: 'Pre-Flop'  },
      { phase: 'flop',     label: 'Flop'      },
      { phase: 'turn',     label: 'Turn'       },
      { phase: 'river',    label: 'River'      },
      { phase: 'showdown', label: 'Showdown'  },
    ];

    phases.forEach(({ phase, label }) => {
      it(`shows "${label}" label for phase "${phase}"`, () => {
        stateSubject.next(makeState({ phase }));
        fixture.detectChanges();
        const el: HTMLElement = fixture.nativeElement.querySelector('[data-cy="phase-badge"]');
        expect(el.textContent?.trim()).toBe(label);
      });
    });
  });

  // ── Active seat highlighting ──────────────────────────────────────────────

  describe('active seat highlighting', () => {
    it('applies rf-seat--active class to an active opponent seat', () => {
      const seats = buildSeats(3, 1);
      seats[1].isActive = true;
      stateSubject.next(makeState({ seats }));
      fixture.detectChanges();
      const seatEls: NodeListOf<HTMLElement> = fixture.nativeElement.querySelectorAll('[data-cy="opponent-seat"]');
      expect(seatEls[0].classList).toContain('rf-seat--active');
    });

    it('does not apply rf-seat--active to inactive seats', () => {
      const seats = buildSeats(3, 1);
      stateSubject.next(makeState({ seats }));
      fixture.detectChanges();
      const seatEls: NodeListOf<HTMLElement> = fixture.nativeElement.querySelectorAll('[data-cy="opponent-seat"]');
      seatEls.forEach(el => expect(el.classList).not.toContain('rf-seat--active'));
    });

    it('applies rf-player-zone--active to player zone when current user is active', () => {
      const seats = buildSeats(2, 1);
      seats[0].isActive = true;
      stateSubject.next(makeState({ seats }));
      fixture.detectChanges();
      const zone: HTMLElement = fixture.nativeElement.querySelector('[data-cy="player-zone"]');
      expect(zone.classList).toContain('rf-player-zone--active');
    });

    it('does not apply rf-player-zone--active when current user is inactive', () => {
      const seats = buildSeats(2, 1);
      seats[0].isActive = false;
      stateSubject.next(makeState({ seats }));
      fixture.detectChanges();
      const zone: HTMLElement = fixture.nativeElement.querySelector('[data-cy="player-zone"]');
      expect(zone.classList).not.toContain('rf-player-zone--active');
    });
  });

  // ── Chip counts ───────────────────────────────────────────────────────────

  describe('chip count display', () => {
    it('shows chip count on opponent seats', () => {
      stateSubject.next(makeState({ seats: buildSeats(2, 1) }));
      fixture.detectChanges();
      const chipsEl: HTMLElement = fixture.nativeElement.querySelector('[data-cy="seat-chips"]');
      expect(chipsEl?.textContent).toContain('1000');
    });

    it('updates chip count after each action', () => {
      const seats = buildSeats(2, 1);
      stateSubject.next(makeState({ seats }));
      fixture.detectChanges();

      const updatedSeats = buildSeats(2, 1);
      updatedSeats[1].chipCount = 600;
      stateSubject.next(makeState({ seats: updatedSeats }));
      fixture.detectChanges();

      const chipsEl: HTMLElement = fixture.nativeElement.querySelector('[data-cy="seat-chips"]');
      expect(chipsEl?.textContent).toContain('600');
    });
  });

  // ── Winner overlay ────────────────────────────────────────────────────────

  describe('winner overlay', () => {
    it('does not render overlay when winner is null', () => {
      const overlay = fixture.nativeElement.querySelector('[data-cy="winner-overlay"]');
      expect(overlay).toBeNull();
    });

    it('renders overlay when winner$ emits', () => {
      winnerSubject.next({ username: 'alice', pot: 500 });
      fixture.detectChanges();
      const overlay = fixture.nativeElement.querySelector('[data-cy="winner-overlay"]');
      expect(overlay).not.toBeNull();
    });

    it('shows correct winner name in overlay', () => {
      winnerSubject.next({ username: 'alice', pot: 500 });
      fixture.detectChanges();
      const title: HTMLElement = fixture.nativeElement.querySelector('[data-cy="winner-title"]');
      expect(title.textContent).toContain('alice');
    });

    it('shows correct pot amount in overlay', () => {
      winnerSubject.next({ username: 'alice', pot: 500 });
      fixture.detectChanges();
      const pot: HTMLElement = fixture.nativeElement.querySelector('[data-cy="winner-pot"]');
      expect(pot.textContent).toContain('500');
    });

    it('dismisses overlay on button click', () => {
      winnerSubject.next({ username: 'alice', pot: 500 });
      fixture.detectChanges();
      const btn: HTMLElement = fixture.nativeElement.querySelector('[data-cy="winner-dismiss"]');
      btn.click();
      fixture.detectChanges();
      const overlay = fixture.nativeElement.querySelector('[data-cy="winner-overlay"]');
      expect(overlay).toBeNull();
    });
  });

  // ── Community card slots ──────────────────────────────────────────────────

  describe('communityCardSlots getter', () => {
    it('always returns exactly 5 items', () => {
      expect(component.communityCardSlots.length).toBe(5);
    });

    it('returns 5 nulls in waiting phase', () => {
      stateSubject.next(makeState({ phase: 'waiting', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      expect(component.visibleCommunityCards.length).toBe(0);
      expect(component.communityCardSlots.every(s => s === null)).toBeTrue();
    });

    it('returns 5 nulls in pre-flop phase', () => {
      stateSubject.next(makeState({ phase: 'pre-flop', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      expect(component.visibleCommunityCards.length).toBe(0);
    });

    it('returns 3 cards on flop', () => {
      stateSubject.next(makeState({ phase: 'flop', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      expect(component.visibleCommunityCards.length).toBe(3);
      const slots = component.communityCardSlots;
      expect(slots.filter(s => s !== null).length).toBe(3);
      expect(slots.filter(s => s === null).length).toBe(2);
    });

    it('returns 4 cards on turn', () => {
      stateSubject.next(makeState({ phase: 'turn', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      expect(component.visibleCommunityCards.length).toBe(4);
    });

    it('returns 5 cards on river', () => {
      stateSubject.next(makeState({ phase: 'river', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      expect(component.visibleCommunityCards.length).toBe(5);
      expect(component.communityCardSlots.every(s => s !== null)).toBeTrue();
    });

    it('returns 5 cards on showdown', () => {
      stateSubject.next(makeState({ phase: 'showdown', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      expect(component.visibleCommunityCards.length).toBe(5);
    });
  });

  // ── DOM: community slots ──────────────────────────────────────────────────

  describe('community slot DOM', () => {
    it('renders exactly 5 [data-cy="community-slot"] containers', () => {
      const slots: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="community-slot"]');
      expect(slots.length).toBe(5);
    });

    it('renders 0 app-card inside slots in waiting phase', () => {
      const cards: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="community-slot"] app-card');
      expect(cards.length).toBe(0);
    });

    it('renders 3 app-card inside slots on flop', () => {
      stateSubject.next(makeState({ phase: 'flop', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      const cards: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="community-slot"] app-card');
      expect(cards.length).toBe(3);
    });

    it('renders 5 app-card inside slots on river', () => {
      stateSubject.next(makeState({ phase: 'river', communityCards: FIVE_CARDS }));
      fixture.detectChanges();
      const cards: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="community-slot"] app-card');
      expect(cards.length).toBe(5);
    });
  });

  // ── Opponent seat rendering ───────────────────────────────────────────────

  describe('opponentSeats getter', () => {
    it('returns empty array when no seats', () => {
      expect(component.opponentSeats.length).toBe(0);
    });

    it('excludes the current user seat', () => {
      stateSubject.next(makeState({ seats: buildSeats(4, 1), currentUserId: 1 }));
      fixture.detectChanges();
      expect(component.opponentSeats.length).toBe(3);
    });

    it('returns 8 opponents for a full 9-player table', () => {
      stateSubject.next(makeState({ seats: buildSeats(9, 1), currentUserId: 1 }));
      fixture.detectChanges();
      expect(component.opponentSeats.length).toBe(8);
    });
  });

  describe('opponent seat DOM', () => {
    it('renders 0 opponent seats when empty', () => {
      const seats: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="opponent-seat"]');
      expect(seats.length).toBe(0);
    });

    it('renders 3 opponent seats for 4 players', () => {
      stateSubject.next(makeState({ seats: buildSeats(4, 1) }));
      fixture.detectChanges();
      const seats: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="opponent-seat"]');
      expect(seats.length).toBe(3);
    });

    it('renders 8 opponent seats for 9 players', () => {
      stateSubject.next(makeState({ seats: buildSeats(9, 1) }));
      fixture.detectChanges();
      const seats: NodeList = fixture.nativeElement.querySelectorAll('[data-cy="opponent-seat"]');
      expect(seats.length).toBe(8);
    });

    it('displays username on each seat', () => {
      stateSubject.next(makeState({ seats: buildSeats(2, 1) }));
      fixture.detectChanges();
      const usernameEl: HTMLElement = fixture.nativeElement.querySelector('[data-cy="seat-username"]');
      expect(usernameEl?.textContent?.trim()).toBe('player-2');
    });

    it('displays chip count on each seat', () => {
      stateSubject.next(makeState({ seats: buildSeats(2, 1) }));
      fixture.detectChanges();
      const chipsEl: HTMLElement = fixture.nativeElement.querySelector('[data-cy="seat-chips"]');
      expect(chipsEl?.textContent).toContain('1000');
    });
  });

  // ── Hole card face direction ──────────────────────────────────────────────

  describe('hole card face direction', () => {
    const DEALT_SEATS: PlayerSeat[] = [
      {
        seatIndex: 0, playerId: 1, username: 'ace', chipCount: 1000,
        holeCards: [{ rank: 'A', suit: 'spades' }, { rank: 'K', suit: 'hearts' }],
        isCurrentUser: true, isActive: true, isDealer: false,
      },
      {
        seatIndex: 1, playerId: 2, username: 'bob', chipCount: 800,
        holeCards: [{ rank: '2', suit: 'clubs' }, { rank: '7', suit: 'diamonds' }],
        isCurrentUser: false, isActive: true, isDealer: false,
      },
    ];

    beforeEach(() => {
      stateSubject.next(makeState({ phase: 'pre-flop', seats: DEALT_SEATS, currentUserId: 1 }));
      fixture.detectChanges();
    });

    it('renders 2 app-card in the player zone', () => {
      const playerZone: HTMLElement = fixture.nativeElement.querySelector('[data-cy="player-zone"]');
      const cards = playerZone.querySelectorAll('app-card');
      expect(cards.length).toBe(2);
    });

    it('renders current user hole cards face-up (aria-label contains rank/suit)', () => {
      const playerZone: HTMLElement = fixture.nativeElement.querySelector('[data-cy="player-zone"]');
      const cards = playerZone.querySelectorAll('.rf-card[aria-label]');
      expect(cards.length).toBe(2);
      cards.forEach((card: Element) => {
        expect(card.getAttribute('aria-label')).not.toBe('Card face down');
      });
    });

    it('renders opponent hole cards face-down', () => {
      const opponentSeat: HTMLElement = fixture.nativeElement.querySelector('[data-cy="opponent-seat"]');
      const cards = opponentSeat.querySelectorAll('.rf-card[aria-label]');
      expect(cards.length).toBe(2);
      cards.forEach((card: Element) => {
        expect(card.getAttribute('aria-label')).toBe('Card face down');
      });
    });
  });

  // ── No cards dealt ────────────────────────────────────────────────────────

  describe('empty hole cards', () => {
    it('shows 2 empty slot placeholders in player zone when no cards dealt', () => {
      const seat: PlayerSeat = {
        seatIndex: 0, playerId: 1, username: 'ace', chipCount: 1000,
        holeCards: [], isCurrentUser: true, isActive: true, isDealer: false,
      };
      stateSubject.next(makeState({ seats: [seat], currentUserId: 1 }));
      fixture.detectChanges();

      const playerZone: HTMLElement = fixture.nativeElement.querySelector('[data-cy="player-zone"]');
      const emptySlots = playerZone.querySelectorAll('[data-cy="hole-card-slot-empty"]');
      expect(emptySlots.length).toBe(2);
    });
  });

  // ── currentUserSeat ───────────────────────────────────────────────────────

  describe('currentUserSeat getter', () => {
    it('returns undefined when no seats', () => {
      expect(component.currentUserSeat).toBeUndefined();
    });

    it('returns the seat marked isCurrentUser', () => {
      stateSubject.next(makeState({ seats: buildSeats(3, 1) }));
      fixture.detectChanges();
      expect(component.currentUserSeat?.username).toBe('ace');
    });
  });

  // ── holeCardSlots ─────────────────────────────────────────────────────────

  describe('holeCardSlots getter', () => {
    it('returns 2 nulls when no current user seat', () => {
      expect(component.holeCardSlots).toEqual([null, null]);
    });

    it('returns 2 nulls when hole cards are empty', () => {
      const seat: PlayerSeat = {
        seatIndex: 0, playerId: 1, username: 'ace', chipCount: 1000,
        holeCards: [], isCurrentUser: true, isActive: true, isDealer: false,
      };
      stateSubject.next(makeState({ seats: [seat] }));
      fixture.detectChanges();
      expect(component.holeCardSlots).toEqual([null, null]);
    });

    it('returns the 2 hole cards when dealt', () => {
      const seat: PlayerSeat = {
        seatIndex: 0, playerId: 1, username: 'ace', chipCount: 1000,
        holeCards: [{ rank: 'A', suit: 'spades' }, { rank: 'K', suit: 'hearts' }],
        isCurrentUser: true, isActive: true, isDealer: false,
      };
      stateSubject.next(makeState({ seats: [seat] }));
      fixture.detectChanges();
      expect(component.holeCardSlots[0]).toEqual({ rank: 'A', suit: 'spades' });
      expect(component.holeCardSlots[1]).toEqual({ rank: 'K', suit: 'hearts' });
    });
  });
});
