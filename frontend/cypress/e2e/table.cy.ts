/**
 * Table Page E2E tests.
 *
 * NOTE: Cypress 14 cannot intercept WebSocket (ws://) traffic natively.
 * These tests focus on static structure and initial render without a live backend.
 */
describe('Table Page', () => {
  beforeEach(() => {
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

    cy.visit('/table/table-42');
  });

  it('renders the table arena', () => {
    cy.get('[data-cy="table-arena"]').should('exist');
  });

  it('renders the felt surface', () => {
    cy.get('.rf-table__felt').should('exist');
  });

  it('renders exactly 5 community card slots', () => {
    cy.get('[data-cy="community-slot"]').should('have.length', 5);
  });

  it('renders the player zone', () => {
    cy.get('[data-cy="player-zone"]').should('exist');
  });

  it('renders the phase badge', () => {
    cy.get('[data-cy="phase-badge"]').should('exist');
  });

  it('renders the pot display', () => {
    cy.get('[data-cy="pot-display"]').should('exist');
  });

  it('shows no community cards in waiting phase (all slots are empty)', () => {
    cy.get('[data-cy="community-slot"] app-card').should('not.exist');
  });

  it('shows 2 empty hole card placeholders in player zone when no cards dealt', () => {
    cy.get('[data-cy="player-zone"] [data-cy="hole-card-slot-empty"]').should('have.length', 2);
  });
});
