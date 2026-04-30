/**
 * Responsive layout E2E tests (Issue #95).
 * Tests at standard mobile breakpoints: 320px, 375px, 414px, 768px.
 */

const VIEWPORTS: Array<{ label: string; width: number; height: number }> = [
  { label: 'iPhone SE', width: 320, height: 568 },
  { label: 'iPhone 14', width: 375, height: 812 },
  { label: 'Pixel', width: 414, height: 896 },
  { label: 'iPad', width: 768, height: 1024 },
];

describe('Responsive — no horizontal overflow', () => {
  VIEWPORTS.forEach(({ label, width, height }) => {
    it(`${label} (${width}px) — leaderboard has no horizontal scrollbar`, () => {
      cy.viewport(width, height);
      cy.visit('/leaderboard');
      cy.get('body').then(($body) => {
        expect($body[0].scrollWidth).to.be.lte($body[0].clientWidth + 1);
      });
    });
  });
});

describe('Responsive — leaderboard hamburger nav', () => {
  it('shows hamburger toggle below 768px and hides nav links', () => {
    cy.viewport(375, 812);
    cy.visit('/leaderboard');
    cy.get('.rf-nav__toggle').should('be.visible');
    cy.get('.rf-nav__links').should('not.be.visible');
  });

  it('opens nav links when hamburger is clicked', () => {
    cy.viewport(375, 812);
    cy.visit('/leaderboard');
    cy.get('.rf-nav__toggle').click();
    cy.get('.rf-nav__links').should('be.visible');
  });

  it('hides hamburger on desktop and shows nav links', () => {
    cy.viewport(1280, 800);
    cy.visit('/leaderboard');
    cy.get('.rf-nav__toggle').should('not.be.visible');
    cy.get('.rf-nav__links').should('be.visible');
  });
});

describe('Responsive — profile stat cards grid', () => {
  VIEWPORTS.forEach(({ label, width, height }) => {
    it(`${label} (${width}px) — stat cards reflow without overflow`, () => {
      cy.viewport(width, height);
      cy.visit('/profile/1');
      cy.get('body').then(($body) => {
        expect($body[0].scrollWidth).to.be.lte($body[0].clientWidth + 1);
      });
    });
  });
});
