import { TestBed } from '@angular/core/testing';
import { skip, take } from 'rxjs';
import { WebSocketService, WS_FACTORY, WS_URL } from './websocket.service';
import { WsMessage } from '../models';

// ---------------------------------------------------------------------------
// MockWebSocket — simulates the browser WebSocket API in tests
// ---------------------------------------------------------------------------
class MockWebSocket {
  readyState: number = WebSocket.CONNECTING;
  url: string;

  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;

  send = jasmine.createSpy('send');
  close = jasmine.createSpy('close').and.callFake(() => {
    this.readyState = WebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close', { wasClean: true, code: 1000 }));
  });

  constructor(url: string) {
    this.url = url;
  }

  simulateOpen(): void {
    this.readyState = WebSocket.OPEN;
    this.onopen?.(new Event('open'));
  }

  simulateMessage(data: unknown): void {
    this.onmessage?.(new MessageEvent('message', { data: JSON.stringify(data) }));
  }

  simulateError(): void {
    this.readyState = WebSocket.CLOSED;
    this.onerror?.(new Event('error'));
  }

  simulateClose(wasClean = true, code = 1000): void {
    this.readyState = WebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close', { wasClean, code }));
  }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe('WebSocketService', () => {
  let service: WebSocketService;
  let mockSocket: MockWebSocket;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        WebSocketService,
        {
          provide: WS_FACTORY,
          useValue: (url: string) => {
            mockSocket = new MockWebSocket(url);
            return mockSocket as unknown as WebSocket;
          }
        }
      ]
    });

    service = TestBed.inject(WebSocketService);
  });

  afterEach(() => {
    service.disconnect();
  });

  // --- Creation ---
  describe('service creation', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });

    it('should start with connected$ = false', (done) => {
      service.connected$.pipe(take(1)).subscribe(connected => {
        expect(connected).toBeFalse();
        done();
      });
    });
  });

  // --- connect() ---
  describe('connect()', () => {
    it('should create a WebSocket to the default URL', () => {
      service.connect();
      expect(mockSocket.url).toBe(WS_URL);
    });

    it('should create a WebSocket to a custom URL when provided', () => {
      service.connect('ws://custom-host/ws');
      expect(mockSocket.url).toBe('ws://custom-host/ws');
    });

    it('should set connected$ to true when socket opens', (done) => {
      service.connect();

      // skip(1) skips the initial BehaviorSubject value (false); take(1) auto-unsubscribes
      service.connected$.pipe(skip(1), take(1)).subscribe(connected => {
        expect(connected).toBeTrue();
        done();
      });

      mockSocket.simulateOpen();
    });

    it('should not create a new socket when already OPEN', () => {
      service.connect();
      mockSocket.simulateOpen();

      const firstSocket = mockSocket;
      service.connect();

      expect(mockSocket).toBe(firstSocket);
    });

    it('should not create a new socket when already CONNECTING', () => {
      service.connect();
      // socket is CONNECTING — simulateOpen() never called

      const firstSocket = mockSocket;
      service.connect(); // second call while still connecting

      expect(mockSocket).toBe(firstSocket);
    });

    it('should set connected$ to false on socket error', (done) => {
      service.connect();
      mockSocket.simulateOpen();

      // skip(1) skips the current 'true'; take(1) auto-unsubscribes to prevent double-done
      service.connected$.pipe(skip(1), take(1)).subscribe(connected => {
        expect(connected).toBeFalse();
        done();
      });

      mockSocket.simulateError();
    });

    it('should set connected$ to false on clean close', (done) => {
      service.connect();
      mockSocket.simulateOpen();

      service.connected$.pipe(skip(1), take(1)).subscribe(connected => {
        expect(connected).toBeFalse();
        done();
      });

      mockSocket.simulateClose(true);
    });
  });

  // --- messages$ ---
  describe('messages$ (receiving)', () => {
    it('should emit a parsed WsMessage on messages$', (done) => {
      const testMsg: WsMessage = { type: 'PLAYER_JOINED', payload: { username: 'alice' } };

      service.connect();
      mockSocket.simulateOpen();

      service.messages$.pipe(take(1)).subscribe(msg => {
        expect(msg).toEqual(testMsg);
        done();
      });

      mockSocket.simulateMessage(testMsg);
    });

    it('should emit multiple messages in order', (done) => {
      const received: WsMessage[] = [];
      const msg1: WsMessage = { type: 'EVENT_A', payload: {} };
      const msg2: WsMessage = { type: 'EVENT_B', payload: { x: 1 } };

      service.connect();
      mockSocket.simulateOpen();

      service.messages$.pipe(take(2)).subscribe({
        next: msg => received.push(msg),
        complete: () => {
          expect(received[0]).toEqual(msg1);
          expect(received[1]).toEqual(msg2);
          done();
        }
      });

      mockSocket.simulateMessage(msg1);
      mockSocket.simulateMessage(msg2);
    });

    it('should not throw on malformed (non-JSON) messages — logs error instead', () => {
      service.connect();
      mockSocket.simulateOpen();

      const consoleSpy = spyOn(console, 'error');
      mockSocket.onmessage?.(new MessageEvent('message', { data: 'not-json{{' }));

      expect(consoleSpy).toHaveBeenCalled();
    });
  });

  // --- sendMessage() ---
  describe('sendMessage()', () => {
    it('should serialize and send { type, payload } as JSON', () => {
      service.connect();
      mockSocket.simulateOpen();

      service.sendMessage('JOIN_ROOM', { roomCode: 'RF-1234' });

      expect(mockSocket.send).toHaveBeenCalledOnceWith(
        JSON.stringify({ type: 'JOIN_ROOM', payload: { roomCode: 'RF-1234' } })
      );
    });

    it('should default payload to {} when not provided', () => {
      service.connect();
      mockSocket.simulateOpen();

      service.sendMessage('PING');

      expect(mockSocket.send).toHaveBeenCalledOnceWith(
        JSON.stringify({ type: 'PING', payload: {} })
      );
    });

    it('should warn and NOT throw when socket is not connected', () => {
      const warnSpy = spyOn(console, 'warn');

      expect(() => service.sendMessage('JOIN_ROOM', {})).not.toThrow();
      expect(warnSpy).toHaveBeenCalled();
    });

    it('should not call send when socket is in CONNECTING state', () => {
      service.connect();
      // socket is CONNECTING — simulateOpen() never called
      const warnSpy = spyOn(console, 'warn');

      service.sendMessage('JOIN_ROOM', {});

      expect(mockSocket.send).not.toHaveBeenCalled();
      expect(warnSpy).toHaveBeenCalled();
    });
  });

  // --- disconnect() ---
  describe('disconnect()', () => {
    it('should call close() on the socket', () => {
      service.connect();
      mockSocket.simulateOpen();

      service.disconnect();

      expect(mockSocket.close).toHaveBeenCalled();
    });

    it('should set connected$ to false after disconnect', (done) => {
      service.connect();
      mockSocket.simulateOpen();

      service.connected$.pipe(skip(1), take(1)).subscribe(connected => {
        expect(connected).toBeFalse();
        done();
      });

      service.disconnect();
    });

    it('should be safe to call when no connection exists', () => {
      expect(() => service.disconnect()).not.toThrow();
    });
  });

  // --- Auto-reconnect ---
  describe('auto-reconnect', () => {
    beforeEach(() => jasmine.clock().install());
    afterEach(() => jasmine.clock().uninstall());

    it('should reconnect after an unclean close', () => {
      service.connect();
      mockSocket.simulateOpen();

      const firstSocket = mockSocket;
      mockSocket.simulateClose(false, 1006); // unclean drop

      jasmine.clock().tick(1001); // past the 1 s base delay

      expect(mockSocket).not.toBe(firstSocket); // a new socket was created
    });

    it('should NOT reconnect after a clean close', () => {
      service.connect();
      mockSocket.simulateOpen();

      const firstSocket = mockSocket;
      mockSocket.simulateClose(true, 1000); // clean close

      jasmine.clock().tick(5000);

      expect(mockSocket).toBe(firstSocket); // no new socket
    });

    it('should NOT reconnect after explicit disconnect()', () => {
      service.connect();
      mockSocket.simulateOpen();

      const firstSocket = mockSocket;
      service.disconnect(); // sets manualDisconnect = true

      jasmine.clock().tick(5000);

      expect(mockSocket).toBe(firstSocket); // no new socket
    });

    it('should reset reconnect counter when connection opens successfully', () => {
      service.connect();
      // Force a few failed attempts
      service['reconnectAttempts'] = 3;

      mockSocket.simulateOpen(); // successful open resets counter

      expect(service['reconnectAttempts']).toBe(0);
    });

    it('should stop reconnecting after MAX_RECONNECT_ATTEMPTS', () => {
      service.connect();
      mockSocket.simulateOpen();

      // Force attempts to the limit so the next close will be the final one
      service['reconnectAttempts'] = service['MAX_RECONNECT_ATTEMPTS'];

      const lastSocket = mockSocket;
      mockSocket.simulateClose(false, 1006);

      jasmine.clock().tick(30001); // well past any delay

      expect(mockSocket).toBe(lastSocket); // no new socket — gave up
    });
  });

  // --- Session cookie documentation ---
  describe('session cookie behavior', () => {
    it('should connect to ws://localhost:8080/ws (browser sends session cookie automatically)', () => {
      // The browser includes session cookies on the WS handshake per RFC 6455 §10.5.
      // No explicit withCredentials flag exists for WebSocket unlike XMLHttpRequest.
      service.connect();
      expect(mockSocket.url).toBe('ws://localhost:8080/ws');
    });
  });
});
