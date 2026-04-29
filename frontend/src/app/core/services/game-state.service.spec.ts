import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { Subject } from 'rxjs';
import { take } from 'rxjs/operators';

import { GameStateService } from './game-state.service';
import { WebSocketService } from './websocket.service';
import { SoundEffectsService } from './sound-effects.service';
import { WsMessage } from '../models/ws-message.model';
import { PlayerSeat, Card, GamePhase } from '../models/game-state.model';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeSeat(overrides: Partial<PlayerSeat> = {}): PlayerSeat {
  return {
    seatIndex: 0,
    playerId: 1,
    username: 'alice',
    chipCount: 1000,
    holeCards: [],
    isCurrentUser: true,
    isActive: false,
    isDealer: false,
    ...overrides,
  };
}

const TWO_CARDS: Card[] = [
  { rank: 'A', suit: 'spades' },
  { rank: 'K', suit: 'hearts' },
];

const THREE_CARDS: Card[] = [
  { rank: '2', suit: 'clubs' },
  { rank: '7', suit: 'diamonds' },
  { rank: 'J', suit: 'spades' },
];

// ---------------------------------------------------------------------------
// Suite
// ---------------------------------------------------------------------------

describe('GameStateService', () => {
  let service: GameStateService;
  let messagesSubject: Subject<WsMessage>;
  let wsSpy: jasmine.SpyObj<WebSocketService>;
  let routerSpy: jasmine.SpyObj<Router>;
  let soundSpy: jasmine.SpyObj<SoundEffectsService>;

  beforeEach(() => {
    messagesSubject = new Subject<WsMessage>();

    wsSpy = jasmine.createSpyObj<WebSocketService>('WebSocketService', [
      'sendMessage',
      'connect',
      'disconnect',
    ]);
    (wsSpy as unknown as { messages$: unknown }).messages$ = messagesSubject.asObservable();

    routerSpy = jasmine.createSpyObj<Router>('Router', ['navigate']);

    soundSpy = jasmine.createSpyObj<SoundEffectsService>('SoundEffectsService', [
      'playChipsClink',
      'playCardFlip',
      'playWinFanfare',
      'toggleMute',
    ]);

    TestBed.configureTestingModule({
      providers: [
        GameStateService,
        { provide: WebSocketService, useValue: wsSpy },
        { provide: Router, useValue: routerSpy },
        { provide: SoundEffectsService, useValue: soundSpy },
      ],
    });

    service = TestBed.inject(GameStateService);
  });

  afterEach(() => {
    service.ngOnDestroy();
  });

  // ── Creation ──────────────────────────────────────────────────────────────

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should start with empty players$', (done) => {
    service.players$.pipe(take(1)).subscribe(players => {
      expect(players).toEqual([]);
      done();
    });
  });

  it('should start with phase$ = waiting', (done) => {
    service.phase$.pipe(take(1)).subscribe(phase => {
      expect(phase).toBe('waiting');
      done();
    });
  });

  it('should start with pot$ = 0', (done) => {
    service.pot$.pipe(take(1)).subscribe(pot => {
      expect(pot).toBe(0);
      done();
    });
  });

  it('should start with activePlayerId$ = -1', (done) => {
    service.activePlayerId$.pipe(take(1)).subscribe(id => {
      expect(id).toBe(-1);
      done();
    });
  });

  // ── GAME_STARTED ──────────────────────────────────────────────────────────

  describe('GAME_STARTED event', () => {
    const seats = [makeSeat({ playerId: 1 }), makeSeat({ playerId: 2, isCurrentUser: false })];

    const gameStartedMsg: WsMessage = {
      type: 'GAME_STARTED',
      payload: {
        tableId: 'table-99',
        seats,
        phase: 'pre-flop' as GamePhase,
        pot: 200,
        currentBet: 10,
        activePlayerId: 1,
        currentUserId: 1,
      },
    };

    beforeEach(() => {
      messagesSubject.next(gameStartedMsg);
    });

    it('should update players$ with seats from payload', (done) => {
      service.players$.pipe(take(1)).subscribe(players => {
        expect(players).toEqual(seats);
        done();
      });
    });

    it('should update phase$ to pre-flop', (done) => {
      service.phase$.pipe(take(1)).subscribe(phase => {
        expect(phase).toBe('pre-flop');
        done();
      });
    });

    it('should update pot$', (done) => {
      service.pot$.pipe(take(1)).subscribe(pot => {
        expect(pot).toBe(200);
        done();
      });
    });

    it('should update activePlayerId$', (done) => {
      service.activePlayerId$.pipe(take(1)).subscribe(id => {
        expect(id).toBe(1);
        done();
      });
    });

    it('should clear communityCards$', (done) => {
      service.communityCards$.pipe(take(1)).subscribe(cards => {
        expect(cards).toEqual([]);
        done();
      });
    });

    it('should navigate to /table/:tableId', () => {
      expect(routerSpy.navigate).toHaveBeenCalledOnceWith(['/table', 'table-99']);
    });

    it('should update gameState$ with combined state', (done) => {
      service.gameState$.pipe(take(1)).subscribe(state => {
        expect(state.phase).toBe('pre-flop');
        expect(state.pot).toBe(200);
        expect(state.currentBet).toBe(10);
        expect(state.seats).toEqual(seats);
        done();
      });
    });
  });

  // ── CARDS_DEALT ──────────────────────────────────────────────────────────

  describe('CARDS_DEALT event', () => {
    const updatedSeats = [makeSeat({ playerId: 1, holeCards: TWO_CARDS })];

    const cardsDealtMsg: WsMessage = {
      type: 'CARDS_DEALT',
      payload: { seats: updatedSeats, holeCards: TWO_CARDS },
    };

    beforeEach(() => {
      messagesSubject.next(cardsDealtMsg);
    });

    it('should update players$ with seats containing hole cards', (done) => {
      service.players$.pipe(take(1)).subscribe(players => {
        expect(players).toEqual(updatedSeats);
        done();
      });
    });

    it('should update holeCards$ with current user hole cards', (done) => {
      service.holeCards$.pipe(take(1)).subscribe(cards => {
        expect(cards).toEqual(TWO_CARDS);
        done();
      });
    });

    it('should sync seats into gameState$', (done) => {
      service.gameState$.pipe(take(1)).subscribe(state => {
        expect(state.seats).toEqual(updatedSeats);
        done();
      });
    });

    it('should call playCardFlip() on SoundEffectsService', () => {
      expect(soundSpy.playCardFlip).toHaveBeenCalled();
    });
  });

  // ── PLAYER_ACTION ──────────────────────────────────────────────────────

  describe('PLAYER_ACTION event', () => {
    const seats = [makeSeat({ playerId: 1 }), makeSeat({ playerId: 2, isActive: true, isCurrentUser: false })];

    const playerActionMsg: WsMessage = {
      type: 'PLAYER_ACTION',
      payload: {
        playerId: 1,
        action: 'CALL',
        amount: 20,
        pot: 300,
        currentBet: 20,
        seats,
        activePlayerId: 2,
      },
    };

    beforeEach(() => {
      messagesSubject.next(playerActionMsg);
    });

    it('should update players$ after an action', (done) => {
      service.players$.pipe(take(1)).subscribe(players => {
        expect(players).toEqual(seats);
        done();
      });
    });

    it('should update pot$', (done) => {
      service.pot$.pipe(take(1)).subscribe(pot => {
        expect(pot).toBe(300);
        done();
      });
    });

    it('should update activePlayerId$ to next player', (done) => {
      service.activePlayerId$.pipe(take(1)).subscribe(id => {
        expect(id).toBe(2);
        done();
      });
    });

    it('should update gameState$ currentBet', (done) => {
      service.gameState$.pipe(take(1)).subscribe(state => {
        expect(state.currentBet).toBe(20);
        done();
      });
    });
  });

  // ── PHASE_CHANGE ──────────────────────────────────────────────────────────

  describe('PHASE_CHANGE event', () => {
    const phaseChangeMsg: WsMessage = {
      type: 'PHASE_CHANGE',
      payload: {
        phase: 'flop' as GamePhase,
        communityCards: THREE_CARDS,
        pot: 400,
        currentBet: 0,
        activePlayerId: 3,
      },
    };

    beforeEach(() => {
      messagesSubject.next(phaseChangeMsg);
    });

    it('should update phase$ to flop', (done) => {
      service.phase$.pipe(take(1)).subscribe(phase => {
        expect(phase).toBe('flop');
        done();
      });
    });

    it('should update communityCards$', (done) => {
      service.communityCards$.pipe(take(1)).subscribe(cards => {
        expect(cards).toEqual(THREE_CARDS);
        done();
      });
    });

    it('should update pot$', (done) => {
      service.pot$.pipe(take(1)).subscribe(pot => {
        expect(pot).toBe(400);
        done();
      });
    });

    it('should update activePlayerId$', (done) => {
      service.activePlayerId$.pipe(take(1)).subscribe(id => {
        expect(id).toBe(3);
        done();
      });
    });

    it('should sync phase and communityCards into gameState$', (done) => {
      service.gameState$.pipe(take(1)).subscribe(state => {
        expect(state.phase).toBe('flop');
        expect(state.communityCards).toEqual(THREE_CARDS);
        done();
      });
    });
  });

  // ── GAME_OVER ──────────────────────────────────────────────────────────────

  describe('GAME_OVER event', () => {
    const finalSeats = [
      makeSeat({ playerId: 1, username: 'alice', chipCount: 2000 }),
      makeSeat({ playerId: 2, username: 'bob', chipCount: 0, isCurrentUser: false }),
    ];

    const gameOverMsg: WsMessage = {
      type: 'GAME_OVER',
      payload: { winnerId: 1, pot: 500, seats: finalSeats },
    };

    beforeEach(() => {
      messagesSubject.next(gameOverMsg);
    });

    it('should update players$ with final seat state', (done) => {
      service.players$.pipe(take(1)).subscribe(players => {
        expect(players).toEqual(finalSeats);
        done();
      });
    });

    it('should set phase$ to showdown', (done) => {
      service.phase$.pipe(take(1)).subscribe(phase => {
        expect(phase).toBe('showdown');
        done();
      });
    });

    it('should set pot$ to 500', (done) => {
      service.pot$.pipe(take(1)).subscribe(pot => {
        expect(pot).toBe(500);
        done();
      });
    });

    it('should sync showdown phase into gameState$', (done) => {
      service.gameState$.pipe(take(1)).subscribe(state => {
        expect(state.phase).toBe('showdown');
        done();
      });
    });

    it('should emit winner$ with correct username and pot', (done) => {
      service.winner$.pipe(take(1)).subscribe(winner => {
        expect(winner).not.toBeNull();
        expect(winner?.username).toBe('alice');
        expect(winner?.pot).toBe(500);
        done();
      });
    });

    it('should emit winner$ as null when winnerId has no matching seat', (done) => {
      messagesSubject.next({
        type: 'GAME_OVER',
        payload: { winnerId: 99, pot: 100, seats: finalSeats },
      });
      service.winner$.pipe(take(1)).subscribe(winner => {
        expect(winner).toBeNull();
        done();
      });
    });
  });

  describe('GAME_STARTED resets winner$', () => {
    it('clears winner$ when a new game starts', (done) => {
      const seats = [makeSeat({ playerId: 1 })];
      messagesSubject.next({
        type: 'GAME_OVER',
        payload: { winnerId: 1, pot: 300, seats },
      });
      messagesSubject.next({
        type: 'GAME_STARTED',
        payload: {
          tableId: 't1',
          seats,
          phase: 'pre-flop',
          pot: 0,
          currentBet: 0,
          activePlayerId: 1,
          currentUserId: 1,
        },
      });
      service.winner$.pipe(take(1)).subscribe(winner => {
        expect(winner).toBeNull();
        done();
      });
    });
  });

  // ── getActivePlayer() — _.find() ──────────────────────────────────────────

  describe('getActivePlayer()', () => {
    it('returns undefined when no players', () => {
      expect(service.getActivePlayer()).toBeUndefined();
    });

    it('returns undefined when activePlayerId has no match', () => {
      messagesSubject.next({
        type: 'GAME_STARTED',
        payload: {
          tableId: 't1',
          seats: [makeSeat({ playerId: 1 })],
          phase: 'pre-flop',
          pot: 0,
          currentBet: 0,
          activePlayerId: 99,
          currentUserId: 1,
        },
      });
      expect(service.getActivePlayer()).toBeUndefined();
    });

    it('uses _.find to locate the active player by playerId', () => {
      const seat1 = makeSeat({ playerId: 1 });
      const seat2 = makeSeat({ playerId: 2, username: 'bob', isCurrentUser: false });

      messagesSubject.next({
        type: 'GAME_STARTED',
        payload: {
          tableId: 't1',
          seats: [seat1, seat2],
          phase: 'pre-flop',
          pot: 0,
          currentBet: 0,
          activePlayerId: 2,
          currentUserId: 1,
        },
      });

      const active = service.getActivePlayer();
      expect(active?.username).toBe('bob');
      expect(active?.playerId).toBe(2);
    });
  });

  // ── sendAction() ──────────────────────────────────────────────────────────

  describe('sendAction()', () => {
    it('sends CHECK with empty payload', () => {
      service.sendAction('CHECK');
      expect(wsSpy.sendMessage).toHaveBeenCalledOnceWith('CHECK', {});
    });

    it('sends CALL with empty payload', () => {
      service.sendAction('CALL');
      expect(wsSpy.sendMessage).toHaveBeenCalledOnceWith('CALL', {});
    });

    it('sends FOLD with empty payload', () => {
      service.sendAction('FOLD');
      expect(wsSpy.sendMessage).toHaveBeenCalledOnceWith('FOLD', {});
    });

    it('sends RAISE with amount payload', () => {
      service.sendAction('RAISE', 150);
      expect(wsSpy.sendMessage).toHaveBeenCalledOnceWith('RAISE', { amount: 150 });
    });
  });

  // ── getSnapshot / patchState ──────────────────────────────────────────────

  describe('getSnapshot()', () => {
    it('returns the current game state', () => {
      const snap = service.getSnapshot();
      expect(snap.phase).toBe('waiting');
    });
  });

  describe('patchState()', () => {
    it('merges partial state into gameState$', (done) => {
      service.patchState({ pot: 500 });
      service.gameState$.pipe(take(1)).subscribe(state => {
        expect(state.pot).toBe(500);
        done();
      });
    });
  });

  // ── Lifecycle ────────────────────────────────────────────────────────────

  describe('ngOnDestroy()', () => {
    it('stops reacting to WS messages after destroy', () => {
      service.ngOnDestroy();
      messagesSubject.next({
        type: 'GAME_STARTED',
        payload: {
          tableId: 'x',
          seats: [],
          phase: 'pre-flop',
          pot: 999,
          currentBet: 0,
          activePlayerId: 1,
          currentUserId: 1,
        },
      });
      // pot should remain 0 since service was destroyed before the message
      expect(service.getSnapshot().pot).toBe(0);
    });
  });
});
