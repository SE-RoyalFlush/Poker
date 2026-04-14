import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CardComponent } from './card.component';

describe('CardComponent', () => {
  let component: CardComponent;
  let fixture: ComponentFixture<CardComponent>;

  function compile(rank: string, suit: string, faceDown = false): void {
    component.rank = rank;
    component.suit = suit;
    component.faceDown = faceDown;
    fixture.detectChanges();
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CardComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(CardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  // ── Creation ────────────────────────────────────────────────────
  it('should create', () => {
    expect(component).toBeTruthy();
  });

  // ── Suit symbols ───────────────────────────────────────────────
  describe('suitSymbol getter', () => {
    it('should return ♠ for spades', () => {
      component.suit = 'spades';
      expect(component.suitSymbol).toBe('♠');
    });

    it('should return ♥ for hearts', () => {
      component.suit = 'hearts';
      expect(component.suitSymbol).toBe('♥');
    });

    it('should return ♦ for diamonds', () => {
      component.suit = 'diamonds';
      expect(component.suitSymbol).toBe('♦');
    });

    it('should return ♣ for clubs', () => {
      component.suit = 'clubs';
      expect(component.suitSymbol).toBe('♣');
    });

    it('should be case-insensitive', () => {
      component.suit = 'Hearts';
      expect(component.suitSymbol).toBe('♥');
    });

    it('should return empty string for unknown suit', () => {
      component.suit = 'joker';
      expect(component.suitSymbol).toBe('');
    });
  });

  // ── Colour classification ──────────────────────────────────────
  describe('isRedSuit getter', () => {
    it('should be true for hearts', () => {
      component.suit = 'hearts';
      expect(component.isRedSuit).toBeTrue();
    });

    it('should be true for diamonds', () => {
      component.suit = 'diamonds';
      expect(component.isRedSuit).toBeTrue();
    });

    it('should be false for spades', () => {
      component.suit = 'spades';
      expect(component.isRedSuit).toBeFalse();
    });

    it('should be false for clubs', () => {
      component.suit = 'clubs';
      expect(component.isRedSuit).toBeFalse();
    });
  });

  // ── Face-up rendering ─────────────────────────────────────────
  describe('face-up card', () => {
    const ranks = ['A', '2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K'];
    const suits = [
      { name: 'spades',   symbol: '♠' },
      { name: 'hearts',   symbol: '♥' },
      { name: 'diamonds', symbol: '♦' },
      { name: 'clubs',    symbol: '♣' },
    ];

    suits.forEach(({ name, symbol }) => {
      ranks.forEach(rank => {
        it(`should render ${rank} of ${name}`, () => {
          compile(rank, name);
          const el: HTMLElement = fixture.nativeElement;
          expect(el.textContent).toContain(rank);
          expect(el.textContent).toContain(symbol);
        });
      });
    });

    it('should not show face-down content when faceDown is false', () => {
      compile('A', 'spades', false);
      const back: HTMLElement = fixture.nativeElement.querySelector('.rf-card__back');
      expect(back).toBeNull();
    });

    it('should apply rf-card--red class for hearts', () => {
      compile('K', 'hearts');
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.classList).toContain('rf-card--red');
    });

    it('should apply rf-card--red class for diamonds', () => {
      compile('Q', 'diamonds');
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.classList).toContain('rf-card--red');
    });

    it('should NOT apply rf-card--red class for spades', () => {
      compile('A', 'spades');
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.classList).not.toContain('rf-card--red');
    });

    it('should NOT apply rf-card--red class for clubs', () => {
      compile('2', 'clubs');
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.classList).not.toContain('rf-card--red');
    });
  });

  // ── Face-down rendering ───────────────────────────────────────
  describe('face-down card', () => {
    beforeEach(() => compile('A', 'spades', true));

    it('should apply rf-card--face-down class', () => {
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.classList).toContain('rf-card--face-down');
    });

    it('should render the back pattern element', () => {
      const pattern: HTMLElement = fixture.nativeElement.querySelector('.rf-card__back-pattern');
      expect(pattern).toBeTruthy();
    });

    it('should NOT render rank or suit symbols', () => {
      const center: HTMLElement = fixture.nativeElement.querySelector('.rf-card__center');
      expect(center).toBeNull();
    });

    it('should NOT apply rf-card--red class when face down', () => {
      component.suit = 'hearts';
      fixture.detectChanges();
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.classList).not.toContain('rf-card--red');
    });

    it('should have aria-label "Card face down"', () => {
      const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
      expect(card.getAttribute('aria-label')).toBe('Card face down');
    });
  });

  // ── Aria label (face-up) ──────────────────────────────────────
  it('should have descriptive aria-label for face-up card', () => {
    compile('A', 'spades');
    const card: HTMLElement = fixture.nativeElement.querySelector('.rf-card');
    expect(card.getAttribute('aria-label')).toBe('A of spades');
  });
});
