const mockRoom = {
  id: 'r1',
  code: 'AB12CD',
  name: 'Test Table',
  gameType: 'NLH',
  smallBlind: 1,
  bigBlind: 2,
  maxPlayers: 9,
  currentPlayers: 4,
  isPrivate: false,
  isFull: false,
  seats: 5,
};

describe('Dashboard - Create & Join Room', () => {
  beforeEach(() => {
    // AuthService maps { ID, username } → { id, username }
    cy.intercept('GET', '**/api/me', { ID: 1, username: 'ace' }).as('getMe');
    cy.intercept('GET', '**/api/csrf', { csrfToken: 'test-csrf' }).as('getCsrf');
    cy.intercept('GET', '**/api/rooms*', [mockRoom]).as('getLiveRooms');
    cy.visit('/dashboard');
    cy.wait('@getMe');
  });

  describe('Create Room', () => {
    it('should call API and navigate to /lobby/code on success', () => {
      cy.intercept('POST', '**/api/rooms', mockRoom).as('createRoom');

      cy.get('app-create-room').within(() => {
        cy.get('input[formControlName="roomName"]').type('My Table');
        cy.get('button[type="submit"]').click();
      });

      cy.wait('@createRoom');
      cy.url().should('include', '/lobby/AB12CD');
    });

    it('should show validation error when room name is empty', () => {
      cy.get('app-create-room').within(() => {
        cy.get('input[formControlName="roomName"]').focus().blur();
        cy.get('mat-error').should('contain', 'Room name is required');
      });
    });
  });

  describe('Join Room', () => {
    it('should call API and navigate to /lobby/code on success', () => {
      cy.intercept('POST', '**/api/rooms/join', mockRoom).as('joinRoom');

      cy.get('app-join-room').within(() => {
        cy.get('input[formControlName="roomCode"]').type('AB12CD');
        cy.get('button[type="submit"]').click();
      });

      cy.wait('@joinRoom');
      cy.url().should('include', '/lobby/AB12CD');
    });

    it('should show validation error for invalid code format', () => {
      cy.get('app-join-room').within(() => {
        cy.get('input[formControlName="roomCode"]').type('RF-7742');
        cy.get('input[formControlName="roomCode"]').blur();
        cy.get('mat-error').should('contain', 'AB12CD');
      });
    });

    it('should enforce maxlength of 6 characters on code input', () => {
      cy.get('app-join-room').within(() => {
        cy.get('input[formControlName="roomCode"]').type('ABCDEFGHIJ');
        cy.get('input[formControlName="roomCode"]').should('have.attr', 'maxlength', '6');
      });
    });

    it('should reveal password field on 403 response', () => {
      cy.intercept('POST', '**/api/rooms/join', { statusCode: 403, body: { message: 'Password required.' } }).as('joinFail');

      cy.get('app-join-room').within(() => {
        cy.get('input[formControlName="roomCode"]').type('AB12CD');
        cy.get('button[type="submit"]').click();
        cy.wait('@joinFail');
        cy.get('input[formControlName="joinPassword"]').should('be.visible');
      });
    });
  });
});
