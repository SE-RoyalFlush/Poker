import { Injectable, InjectionToken, Inject } from '@angular/core';
import { Subject, BehaviorSubject, Observable } from 'rxjs';
import { WsMessage } from '../models';

export const WS_URL = 'ws://localhost:8080/ws';

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
 * Usage:
 *   this.wsService.connect();
 *   this.wsService.messages$.subscribe(msg => { ... });
 *   this.wsService.sendMessage('JOIN_ROOM', { roomCode: 'RF-1234' });
 *   this.wsService.disconnect();
 */
@Injectable({ providedIn: 'root' })
export class WebSocketService {
  private socket: WebSocket | null = null;

  // Subject (not BehaviorSubject) — messages are events with no meaningful last value.
  private readonly messagesSubject = new Subject<WsMessage>();
  public readonly messages$: Observable<WsMessage> = this.messagesSubject.asObservable();

  // BehaviorSubject — new subscribers immediately receive the current connection state.
  private readonly connectedSubject = new BehaviorSubject<boolean>(false);
  public readonly connected$: Observable<boolean> = this.connectedSubject.asObservable();

  constructor(@Inject(WS_FACTORY) private wsFactory: (url: string) => WebSocket) {}

  /**
   * Open the WebSocket connection to the game server.
   * Calling connect() while already OPEN is a no-op.
   * @param url WebSocket endpoint (defaults to WS_URL)
   */
  connect(url: string = WS_URL): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return;
    }

    this.socket = this.wsFactory(url);

    this.socket.onopen = () => {
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
      if (!event.wasClean) {
        console.warn(`[WebSocketService] Connection dropped (code: ${event.code})`);
      }
    };
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
   * Close the WebSocket connection cleanly.
   * Safe to call when no connection exists.
   */
  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }
}
