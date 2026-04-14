import { Injectable, InjectionToken, Inject } from '@angular/core';
import { Subject, BehaviorSubject, Observable } from 'rxjs';
import { WsMessage } from '../models';
import { WS_URL } from '../config/endpoints';

export { WS_URL } from '../config/endpoints';

/**
 * Injectable factory for creating WebSocket instances.
 * Overridden in tests to return a MockWebSocket instead of a real browser WebSocket.
 */
export const WS_FACTORY = new InjectionToken<(url: string) => WebSocket>(
  'WS_FACTORY',
  { providedIn: 'root', factory: () => (url: string) => new WebSocket(url) }
);

/**
 * WebSocketService manages the WebSocket connection to the game server.
 *
 * Session cookies (set by the backend at login) are sent automatically
 * by the browser on the WebSocket handshake — standard browser behavior
 * per RFC 6455 §10.5. No explicit `withCredentials` flag exists for
 * WebSocket (unlike XMLHttpRequest); cookies are always included when
 * the WS origin matches the cookie domain/SameSite policy.
 *
 * On unexpected disconnection the service retries with exponential backoff
 * (base 1 s, doubling each attempt, capped at 30 s) up to MAX_RECONNECT_ATTEMPTS.
 * Calling disconnect() cancels any pending retry.
 *
 * Usage:
 *   this.wsService.connect();
 *   this.wsService.messages$.subscribe(msg => { ... });
 *   this.wsService.sendMessage('JOIN_ROOM', { roomCode: 'AB12CD' });
 *   this.wsService.disconnect();
 */
@Injectable({ providedIn: 'root' })
export class WebSocketService {
  private socket: WebSocket | null = null;
  private manualDisconnect = false;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  private readonly MAX_RECONNECT_ATTEMPTS = 5;
  private readonly BASE_RECONNECT_DELAY_MS = 1000;

  // Subject (not BehaviorSubject) — messages are events with no meaningful last value.
  private readonly messagesSubject = new Subject<WsMessage>();
  public readonly messages$: Observable<WsMessage> = this.messagesSubject.asObservable();

  // BehaviorSubject — new subscribers immediately receive the current connection state.
  private readonly connectedSubject = new BehaviorSubject<boolean>(false);
  public readonly connected$: Observable<boolean> = this.connectedSubject.asObservable();

  constructor(@Inject(WS_FACTORY) private wsFactory: (url: string) => WebSocket) {}

  /**
   * Open the WebSocket connection to the game server.
   * No-op if the socket is already OPEN or CONNECTING.
   * Cancels any pending reconnect and resets the manual-disconnect flag.
   * @param url WebSocket endpoint (defaults to WS_URL)
   */
  connect(url: string = WS_URL): void {
    this.manualDisconnect = false;

    if (
      this.socket &&
      (this.socket.readyState === WebSocket.OPEN ||
        this.socket.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    this.socket = this.wsFactory(url);

    this.socket.onopen = () => {
      this.reconnectAttempts = 0;
      this.connectedSubject.next(true);
    };

    this.socket.onmessage = (event: MessageEvent) => {
      try {
        const message: WsMessage = JSON.parse(event.data as string);
        this.messagesSubject.next(message);
      } catch (e) {
        console.error('[WebSocketService] Failed to parse message:', event.data, e);
      }
    };

    this.socket.onerror = (event: Event) => {
      console.error('[WebSocketService] Connection error:', event);
      this.connectedSubject.next(false);
    };

    this.socket.onclose = (event: CloseEvent) => {
      this.connectedSubject.next(false);
      if (!event.wasClean && !this.manualDisconnect) {
        this.scheduleReconnect(url);
      } else if (event.wasClean) {
        this.reconnectAttempts = 0;
      }
    };
  }

  /**
   * Schedule a reconnect attempt with exponential backoff.
   * Stops after MAX_RECONNECT_ATTEMPTS consecutive failures.
   */
  private scheduleReconnect(url: string): void {
    if (this.reconnectAttempts >= this.MAX_RECONNECT_ATTEMPTS) {
      console.error('[WebSocketService] Max reconnect attempts reached, giving up');
      return;
    }

    const delay = Math.min(
      this.BASE_RECONNECT_DELAY_MS * Math.pow(2, this.reconnectAttempts),
      30000
    );
    this.reconnectAttempts++;

    console.warn(
      `[WebSocketService] Reconnecting in ${delay}ms` +
        ` (attempt ${this.reconnectAttempts}/${this.MAX_RECONNECT_ATTEMPTS})`
    );

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect(url);
    }, delay);
  }

  /**
   * Send a typed message to the server.
   * Serialized as JSON with { type, payload } envelope.
   * Warns (does not throw) when the socket is not open.
   *
   * @param type    Event type string, e.g. 'JOIN_ROOM', 'FOLD'
   * @param payload Arbitrary data to include with the message
   */
  sendMessage(type: string, payload: unknown = {}): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.warn('[WebSocketService] sendMessage called but socket is not open');
      return;
    }
    this.socket.send(JSON.stringify({ type, payload } as WsMessage));
  }

  /**
   * Close the WebSocket connection cleanly and cancel any pending reconnect.
   * Safe to call when no connection exists.
   */
  disconnect(): void {
    this.manualDisconnect = true;
    this.reconnectAttempts = 0;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }
}
