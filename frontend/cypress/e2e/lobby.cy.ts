/**
 * Lobby E2E tests.
 *
 * NOTE: Cypress 14 cannot intercept WebSocket (ws://) traffic natively,
 * so tests that verify PLAYER_JOINED/PLAYER_LEFT behaviour are covered by
 * the Jasmine unit tests in lobby.spec.ts using MockWebSocket + WS_FACTORY.
 * These E2E tests focus on initial page render and static structure.
 */
describe('Lobby Page', () => {
  beforeEach(() => {
    // The WS connection to localhost:8080 will fail in CI/test — suppress only that error.
    cy.on('uncaught:exception', (err) => {
      const errorText = `${err.message}\n${err.stack ?? ''}`;
      const isExpectedWebSocketFailure =
        errorText.includes('WebSocket') &&
        (errorText.includes('ws://localhost:8080') ||
          errorText.includes('localhost:8080')) &&
        (errorText.includes('failed') ||
          errorText.includes('closed before the connection is established') ||
          errorText.includes('ECONNREFUSED'));

      if (isExpectedWebSocketFailure) {
        return false;
      }
    });

    cy.intercept('GET', '**/api/me', { ID: 1, username: 'ace' }).as('getMe');
    cy.intercept('GET', '**/api/csrf', { csrfToken: 'test-csrf' }).as('getCsrf');

    cy.visit('/lobby/AB12CD');
  });

  it('should display the room code in the heading', () => {
    cy.get('h1').should('contain', 'AB12CD');
  });

  it('should show 0 players initially', () => {
    cy.contains('0 player').should('exist');
  });

  it('should render an empty player list', () => {
    cy.get('.rf-lobby__player-list').should('exist');
    cy.get('.rf-lobby__player-card').should('not.exist');
  });
});

/**
 * E2E tests for the Lobby page — Ready Button & Chat UI (Issue #23).
 *
 * WebSocket connections are stubbed so tests run without a live backend.
 */
describe('Lobby - Ready Button & Chat UI', () => {
  beforeEach(() => {
    // Install the WebSocket stub via onBeforeLoad so it's in place before
    // Angular bootstraps — stubs set after cy.visit() apply to the replaced
    // window and won't intercept the app's WS constructor calls.
    cy.visit('/lobby/AB12CD', {
      onBeforeLoad(win) {
        const fakeWs = {
          readyState: WebSocket.OPEN,
          send: cy.stub().as('wsSend'),
          close: cy.stub(),
          onmessage: null as ((ev: MessageEvent) => void) | null,
          onerror: null as ((ev: Event) => void) | null,
        };
        cy.stub(win, 'WebSocket').returns(fakeWs);
      },
    });
  });

  // ── Ready button ─────────────────────────────────────────────
  it('renders the ready button', () => {
    cy.get('[data-cy="ready-btn"]').should('be.visible');
  });

  it('shows "Not Ready" label initially', () => {
    cy.get('[data-cy="ready-btn"]').should('contain.text', 'Not Ready');
  });

  it('toggles to Ready state on click', () => {
    cy.get('[data-cy="ready-btn"]').click();
    cy.get('[data-cy="ready-btn"]')
      .should('contain.text', 'Ready')
      .and('have.class', 'rf-ready-btn--active');
  });

  it('toggles back to Not Ready on second click', () => {
    cy.get('[data-cy="ready-btn"]').click().click();
    cy.get('[data-cy="ready-btn"]')
      .should('contain.text', 'Not Ready')
      .and('not.have.class', 'rf-ready-btn--active');
  });

  it('sends PLAYER_READY event over WebSocket when toggled', () => {
    cy.get('[data-cy="ready-btn"]').click();
    cy.get('@wsSend').should('have.been.calledOnce');
    cy.get('@wsSend').invoke('args', 0, 0).then((arg: string) => {
      const msg = JSON.parse(arg);
      expect(msg.type).to.eq('PLAYER_READY');
      expect(msg.payload.isReady).to.eq(true);
    });
  });

  // ── Chat input ───────────────────────────────────────────────
  it('renders the chat input and send button', () => {
    cy.get('[data-cy="chat-input"]').should('be.visible');
    cy.get('[data-cy="send-btn"]').should('be.visible');
  });

  it('clears the input after clicking Send', () => {
    cy.get('[data-cy="chat-input"]').type('Hello table!');
    cy.get('[data-cy="send-btn"]').click();
    cy.get('[data-cy="chat-input"]').should('have.value', '');
  });

  it('clears the input after pressing Enter', () => {
    cy.get('[data-cy="chat-input"]').type('All in!{enter}');
    cy.get('[data-cy="chat-input"]').should('have.value', '');
  });

  it('does not send an empty message', () => {
    cy.get('[data-cy="send-btn"]').click();
    cy.get('@wsSend').should('not.have.been.called');
  });

  it('appends the message to the chat list after sending', () => {
    cy.get('[data-cy="chat-input"]').type('Good luck everyone!');
    cy.get('[data-cy="send-btn"]').click();
    cy.get('.rf-chat__text').should('contain.text', 'Good luck everyone!');
  });
});
