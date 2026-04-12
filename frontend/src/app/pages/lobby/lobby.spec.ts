import { ComponentFixture, TestBed } from '@angular/core/testing';
import { convertToParamMap, ActivatedRoute, Router } from '@angular/router';
import { BehaviorSubject } from 'rxjs';

import { Lobby } from './lobby';
import { WebSocketService, WS_FACTORY } from '../../core/services/websocket.service';
import { AuthService } from '../../core/services/auth.service';
import { Player } from '../../core/models';
import { MockWebSocket } from '../../../../test-utils/mock-websocket';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
const ALICE: Player = { id: 1, username: 'alice', isHost: true,  isReady: false };
const BOB: Player   = { id: 2, username: 'bob',   isHost: false, isReady: false };

function mockAuthService(username = 'alice') {
  return {
    currentUser$: new BehaviorSubject({ id: 1, username }),
  };
}

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
        { provide: AuthService, useValue: mockAuthService() },
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
              queryParamMap: convertToParamMap({ code: 'AB12CD' }),
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
  it('should read roomCode from query params', () => {
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

  // --- Ready button: initial state ---
  it('should initialize isReady to false', () => {
    expect(component.isReady).toBeFalse();
  });

  it('should render ready button with "Not Ready" text initially', () => {
    fixture.detectChanges();
    const btn: HTMLElement = fixture.nativeElement.querySelector('.rf-lobby__ready-btn');
    expect(btn.textContent?.trim()).toBe('Not Ready');
  });

  // --- Ready button: toggle ---
  it('should toggle isReady on toggleReady()', () => {
    component.toggleReady();
    expect(component.isReady).toBeTrue();
    component.toggleReady();
    expect(component.isReady).toBeFalse();
  });

  it('should send TOGGLE_READY message on toggleReady()', () => {
    mockSocket.simulateOpen();
    component.toggleReady();

    const calls = mockSocket.send.calls.all().map(c => JSON.parse(c.args[0] as string));
    const toggleCall = calls.find(m => m.type === 'TOGGLE_READY');
    expect(toggleCall).toBeTruthy();
    expect(toggleCall.payload.isReady).toBeTrue();
  });

  it('should apply active class to ready button when isReady is true', () => {
    component.isReady = true;
    fixture.detectChanges();
    const btn: HTMLElement = fixture.nativeElement.querySelector('.rf-lobby__ready-btn');
    expect(btn.classList).toContain('rf-lobby__ready-btn--active');
  });

  // --- PLAYER_READY event ---
  it('should update player isReady on PLAYER_READY event', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: BOB });
    mockSocket.simulateMessage({ type: 'PLAYER_READY', payload: { id: BOB.id, isReady: true } });

    expect(component.players[0].isReady).toBeTrue();
  });

  it('should render ready badge for a ready player', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: { ...BOB, isReady: true } });
    fixture.detectChanges();

    const badge: HTMLElement = fixture.nativeElement.querySelector('.rf-lobby__ready-badge');
    expect(badge).toBeTruthy();
  });

  it('should apply avatar ready class for a ready player', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'PLAYER_JOINED', payload: { ...BOB, isReady: true } });
    fixture.detectChanges();

    const avatar: HTMLElement = fixture.nativeElement.querySelector('.rf-lobby__avatar');
    expect(avatar.classList).toContain('rf-lobby__avatar--ready');
  });

  // --- Chat: initial state ---
  it('should initialize chatMessages as empty array', () => {
    expect(component.chatMessages.length).toBe(0);
  });

  // --- Chat: CHAT_MESSAGE event ---
  it('should append a message on CHAT_MESSAGE event', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({
      type: 'CHAT_MESSAGE',
      payload: { username: 'alice', text: 'Hello!' },
    });

    expect(component.chatMessages.length).toBe(1);
    expect(component.chatMessages[0].text).toBe('Hello!');
  });

  it('should accumulate multiple chat messages', () => {
    mockSocket.simulateOpen();
    mockSocket.simulateMessage({ type: 'CHAT_MESSAGE', payload: { username: 'alice', text: 'Hi' } });
    mockSocket.simulateMessage({ type: 'CHAT_MESSAGE', payload: { username: 'bob', text: 'Hey' } });

    expect(component.chatMessages.length).toBe(2);
  });

  // --- Chat: sendChat ---
  it('should send CHAT_MESSAGE and clear input on sendChat()', () => {
    mockSocket.simulateOpen();
    component.chatInput = 'Hello world';
    component.sendChat();

    const calls = mockSocket.send.calls.all().map(c => JSON.parse(c.args[0] as string));
    const chatCall = calls.find(m => m.type === 'CHAT_MESSAGE');
    expect(chatCall).toBeTruthy();
    expect(chatCall.payload.text).toBe('Hello world');
    expect(component.chatInput).toBe('');
  });

  it('should not send CHAT_MESSAGE when input is empty', () => {
    mockSocket.simulateOpen();
    component.chatInput = '   ';
    component.sendChat();

    const calls = mockSocket.send.calls.all().map(c => JSON.parse(c.args[0] as string));
    const chatCall = calls.find(m => m.type === 'CHAT_MESSAGE');
    expect(chatCall).toBeUndefined();
  });

  // --- Chat: Enter key ---
  it('should call sendChat() when Enter key is pressed', () => {
    spyOn(component, 'sendChat');
    const event = new KeyboardEvent('keydown', { key: 'Enter' });
    component.onChatKeydown(event);
    expect(component.sendChat).toHaveBeenCalled();
  });

  it('should not call sendChat() when Shift+Enter is pressed', () => {
    spyOn(component, 'sendChat');
    const event = new KeyboardEvent('keydown', { key: 'Enter', shiftKey: true });
    component.onChatKeydown(event);
    expect(component.sendChat).not.toHaveBeenCalled();
  });
});

describe('Lobby (missing code param)', () => {
  it('should redirect to /dashboard when code query param is missing', async () => {
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    await TestBed.configureTestingModule({
      imports: [Lobby],
      providers: [
        WebSocketService,
        { provide: Router, useValue: routerSpy },
        { provide: AuthService, useValue: mockAuthService() },
        {
          provide: WS_FACTORY,
          useValue: (url: string) => new MockWebSocket(url) as unknown as WebSocket,
        },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: convertToParamMap({}) } },
        },
      ],
    }).compileComponents();

    const f = TestBed.createComponent(Lobby);
    f.detectChanges();

    expect(routerSpy.navigate).toHaveBeenCalledWith(['/dashboard']);

    TestBed.resetTestingModule();
  });
});
