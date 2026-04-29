import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { GameControlsComponent, GameAction } from './game-controls.component';
import { SoundEffectsService } from '../../../core/services/sound-effects.service';

describe('GameControlsComponent', () => {
  let component: GameControlsComponent;
  let fixture: ComponentFixture<GameControlsComponent>;
  let soundSpy: jasmine.SpyObj<SoundEffectsService>;

  beforeEach(async () => {
    soundSpy = jasmine.createSpyObj<SoundEffectsService>('SoundEffectsService', [
      'playChipsClink',
      'playCardFlip',
      'playWinFanfare',
      'toggleMute',
    ]);

    await TestBed.configureTestingModule({
      imports: [GameControlsComponent],
      providers: [
        provideNoopAnimations(),
        { provide: SoundEffectsService, useValue: soundSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(GameControlsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('when isActivePlayer is false', () => {
    beforeEach(() => {
      component.isActivePlayer = false;
      fixture.detectChanges();
    });

    it('should disable Check button', () => {
      const btn = fixture.nativeElement.querySelector('[data-testid="btn-check"]');
      expect(btn.disabled).toBeTrue();
    });

    it('should disable Call button', () => {
      const btn = fixture.nativeElement.querySelector('[data-testid="btn-call"]');
      expect(btn.disabled).toBeTrue();
    });

    it('should disable Raise button', () => {
      const btn = fixture.nativeElement.querySelector('[data-testid="btn-raise"]');
      expect(btn.disabled).toBeTrue();
    });

    it('should disable Fold button', () => {
      const btn = fixture.nativeElement.querySelector('[data-testid="btn-fold"]');
      expect(btn.disabled).toBeTrue();
    });

    it('should not emit action on Check click', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onCheck();
      expect(emitted.length).toBe(0);
    });

    it('should not emit action on Fold click', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onFold();
      expect(emitted.length).toBe(0);
    });
  });

  describe('when isActivePlayer is true', () => {
    beforeEach(() => {
      component.isActivePlayer = true;
      component.callAmount = 10;
      component.maxRaise = 200;
      component.ngOnChanges();
      fixture.detectChanges();
    });

    it('should enable all action buttons', () => {
      ['btn-check', 'btn-call', 'btn-raise', 'btn-fold'].forEach((testId) => {
        const btn = fixture.nativeElement.querySelector(`[data-testid="${testId}"]`);
        expect(btn.disabled).toBeFalse();
      });
    });

    it('should emit CHECK action on Check click', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onCheck();
      expect(emitted).toEqual([{ type: 'CHECK' }]);
    });

    it('should emit CALL with callAmount on Call click', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onCall();
      expect(emitted).toEqual([{ type: 'CALL', amount: 10 }]);
    });

    it('should call playChipsClink() on CALL', () => {
      component.onCall();
      expect(soundSpy.playChipsClink).toHaveBeenCalled();
    });

    it('should emit FOLD action on Fold click', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onFold();
      expect(emitted).toEqual([{ type: 'FOLD' }]);
    });

    it('should toggle raise panel on Raise click', () => {
      expect(component.showRaiseInput).toBeFalse();
      component.onRaiseToggle();
      expect(component.showRaiseInput).toBeTrue();
      component.onRaiseToggle();
      expect(component.showRaiseInput).toBeFalse();
    });

    it('should show raise panel in DOM after toggle', () => {
      component.onRaiseToggle();
      fixture.detectChanges();
      const panel = fixture.nativeElement.querySelector('[data-testid="raise-panel"]');
      expect(panel).toBeTruthy();
    });

    it('should emit RAISE with amount when raise form submitted with valid amount', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onRaiseToggle();
      component.raiseForm.patchValue({ amount: 50 });
      component.onRaiseSubmit();
      expect(emitted).toEqual([{ type: 'RAISE', amount: 50 }]);
    });

    it('should call playChipsClink() on valid RAISE', () => {
      component.onRaiseToggle();
      component.raiseForm.patchValue({ amount: 50 });
      component.onRaiseSubmit();
      expect(soundSpy.playChipsClink).toHaveBeenCalled();
    });

    it('should not emit RAISE when raise amount is below minimum', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onRaiseToggle();
      component.raiseForm.patchValue({ amount: 5 });
      component.onRaiseSubmit();
      expect(emitted.length).toBe(0);
    });

    it('should not emit RAISE when raise amount exceeds maximum', () => {
      const emitted: GameAction[] = [];
      component.action.subscribe((a) => emitted.push(a));
      component.onRaiseToggle();
      component.raiseForm.patchValue({ amount: 300 });
      component.onRaiseSubmit();
      expect(emitted.length).toBe(0);
    });

    it('should close raise panel after successful raise', () => {
      component.onRaiseToggle();
      component.raiseForm.patchValue({ amount: 50 });
      component.onRaiseSubmit();
      expect(component.showRaiseInput).toBeFalse();
    });
  });

  describe('raise validation', () => {
    beforeEach(() => {
      component.isActivePlayer = true;
      component.callAmount = 20;
      component.maxRaise = 100;
      component.ngOnChanges();
      fixture.detectChanges();
    });

    it('should use callAmount as minimum raise', () => {
      expect(component.raiseMin).toBe(20);
    });

    it('should mark form invalid when amount equals zero', () => {
      component.raiseForm.patchValue({ amount: 0 });
      expect(component.raiseForm.invalid).toBeTrue();
    });

    it('should mark form valid when amount equals callAmount', () => {
      component.raiseForm.patchValue({ amount: 20 });
      expect(component.raiseForm.valid).toBeTrue();
    });

    it('should mark form valid when amount is between min and max', () => {
      component.raiseForm.patchValue({ amount: 60 });
      expect(component.raiseForm.valid).toBeTrue();
    });
  });
});
