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
});
