/**
 * MockWebSocket — simulates the browser WebSocket API in tests.
 * Shared between websocket.service.spec.ts and lobby.spec.ts.
 */
export class MockWebSocket {
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
