import { ComponentFixture, TestBed } from '@angular/core/testing';
import { convertToParamMap, ActivatedRoute, Router } from '@angular/router';

import { Lobby } from './lobby';
import { WebSocketService, WS_FACTORY } from '../../core/services/websocket.service';
import { Player } from '../../core/models';
import { MockWebSocket } from '../../../../test-utils/mock-websocket';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
const ALICE: Player = { id: 1, username: 'alice', isHost: true };
const BOB: Player   = { id: 2, username: 'bob',   isHost: false };

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe('Lobby', () => {
  let fixture: ComponentFixture<Lobby>;
  let component: Lobby;
  let wsService: WebSocketService;
  let mockSocket: MockWebSocket;
  let routerSpy: jasmine.SpyObj<Router>;

  beforeEach(async () => {
    routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    await TestBed.configureTestingModule({
      imports: [Lobby],
      providers: [
        WebSocketService,
        { provide: Router, useValue: routerSpy },
        {
          provide: WS_FACTORY,
          useValue: (url: string) => {
            mockSocket = new MockWebSocket(url);
            return mockSocket as unknown as WebSocket;
          },
        },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: convertToParamMap({ code: 'AB12CD' }),
            },
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(Lobby);
    component = fixture.componentInstance;
    wsService = TestBed.inject(WebSocketService);

    // Triggers ngOnInit → wsService.connect() → captures mockSocket
    fixture.detectChanges();
  });

  afterEach(() => {
    component.ngOnDestroy();
  });

  // --- Creation ---
  it('should be created', () => {
    expect(component).toBeTruthy();
  });

  // --- Room code ---
  it('should read roomCode from route params', () => {
    expect(component.roomCode).toBe('AB12CD');
  });

  // --- JOIN_ROOM ---
  it('should send JOIN_ROOM when WebSocket opens', () => {
    mockSocket.simulateOpen();

    expect(mockSocket.send).toHaveBeenCalledOnceWith(
      JSON.stringify({ type: 'JOIN_ROOM', payload: { roomCode: 'AB12CD' } })
    );
  });

  // --- PLAYER_JOINED ---
  it('should replace the roster from ROOM_STATE', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({
      type: 'ROOM_STATE',
      payload: { roomCode: 'AB12CD', players: [ALICE, BOB] },
    });

    expect(component.players.length).toBe(2);
    expect(component.sortedPlayers[0].username).toBe('alice');
  });

  it('should add a player on PLAYER_JOINED', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });

    expect(component.players.length).toBe(1);
    expect(component.players[0].username).toBe('bob');
  });

  it('should deduplicate players on duplicate PLAYER_JOINED (uniqBy id)', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });

    expect(component.players.length).toBe(1);
  });

  // --- PLAYER_LEFT ---
  it('should remove a player on PLAYER_LEFT', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });
    mockSocket.simulateMessage({ type: 'PLAYER_LEFT', payload: { id: BOB.id } });

    expect(component.players.length).toBe(0);
  });

  // --- sortedPlayers (host first) ---
  it('should sort host to the top via sortedPlayers', () => {
    mockSocket.simulateOpen();
    // Add non-host first, then host
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: ALICE });

    expect(component.sortedPlayers[0].isHost).toBeTrue();
    expect(component.sortedPlayers[0].username).toBe('alice');
  });

  // --- ngOnDestroy ---
  it('should call wsService.disconnect() on destroy', () => {
    const disconnectSpy = spyOn(wsService, 'disconnect').and.callThrough();
    component.ngOnDestroy();

    expect(disconnectSpy).toHaveBeenCalled();
  });

  // --- Template: room code ---
  it('should render roomCode in the heading', () => {
    const h1: HTMLElement = fixture.nativeElement.querySelector('h1');
    expect(h1.textContent).toContain('AB12CD');
  });

  // --- Template: player count ---
  it('should update player count in the template', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });
    fixture.detectChanges();

    const p: HTMLElement = fixture.nativeElement.querySelector('header p');
    expect(p.textContent).toContain('1 player');
  });

  // --- Template: player list item ---
  it('should render player username in the list', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });
    fixture.detectChanges();

    const item: HTMLElement = fixture.nativeElement.querySelector('.rf-lobby__username');
    expect(item.textContent?.trim()).toBe('bob');
  });

  // --- Template: host badge ---
  it('should render a Host badge for the host player', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: ALICE });
    fixture.detectChanges();

    const badge: HTMLElement = fixture.nativeElement.querySelector('.rf-lobby__host-badge');
    expect(badge).toBeTruthy();
    expect(badge.textContent?.trim()).toBe('Host');
  });
});

describe('Lobby (missing code param)', () => {
  it('should redirect to /dashboard when code route param is missing', async () => {
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    await TestBed.configureTestingModule({
      imports: [Lobby],
      providers: [
        WebSocketService,
        { provide: Router, useValue: routerSpy },
        {
          provide: WS_FACTORY,
          useValue: (url: string) => new MockWebSocket(url) as unknown as WebSocket,
        },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({}) } },
        },
      ],
    }).compileComponents();

    const f = TestBed.createComponent(Lobby);
    f.detectChanges();

    expect(routerSpy.navigate).toHaveBeenCalledWith(['/dashboard']);

    TestBed.resetTestingModule();
  });
});
