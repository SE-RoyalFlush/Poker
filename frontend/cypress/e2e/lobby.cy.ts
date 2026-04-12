/**
 * Lobby E2E tests.
 *
 * NOTE: Cypress 14 cannot intercept WebSocket (ws://) traffic natively,
 * so tests that verify PLAYER_JOINED/PLAYER_LEFT/PLAYER_READY/CHAT_MESSAGE
 * behaviour are covered by Jasmine unit tests in lobby.spec.ts using
 * MockWebSocket + WS_FACTORY.
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

    cy.visit('/lobby?code=AB12CD');
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

  // --- Ready button ---
  it('should render the ready button with "Not Ready" initially', () => {
    cy.get('.rf-lobby__ready-btn').should('exist').and('contain', 'Not Ready');
  });

  it('should not have active class on ready button initially', () => {
    cy.get('.rf-lobby__ready-btn').should('not.have.class', 'rf-lobby__ready-btn--active');
  });

  it('should toggle ready button text and class on click', () => {
    cy.get('.rf-lobby__ready-btn').click();
    cy.get('.rf-lobby__ready-btn')
      .should('contain', 'Ready')
      .and('have.class', 'rf-lobby__ready-btn--active');

    cy.get('.rf-lobby__ready-btn').click();
    cy.get('.rf-lobby__ready-btn')
      .should('contain', 'Not Ready')
      .and('not.have.class', 'rf-lobby__ready-btn--active');
  });

  // --- Chat section ---
  it('should render the chat section', () => {
    cy.get('.rf-lobby__chat').should('exist');
  });

  it('should render chat message input', () => {
    cy.get('.rf-lobby__chat-input-row input[matInput]').should('exist');
  });

  it('should render the Send button', () => {
    cy.get('.rf-lobby__chat-send').should('exist').and('contain', 'Send');
  });

  it('should have Send button disabled when input is empty', () => {
    cy.get('.rf-lobby__chat-send').should('be.disabled');
  });

  it('should enable Send button when input has text', () => {
    cy.get('.rf-lobby__chat-input-row input[matInput]').type('Hello');
    cy.get('.rf-lobby__chat-send').should('not.be.disabled');
  });

  it('should show empty state message in chat initially', () => {
    cy.get('.rf-lobby__chat-messages').should('contain', 'No messages yet');
  });
});
