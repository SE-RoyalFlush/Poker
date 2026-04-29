import { TestBed } from '@angular/core/testing';
import { take } from 'rxjs/operators';

import { SoundEffectsService } from './sound-effects.service';

const MUTE_KEY = 'rfx_sound_muted';

describe('SoundEffectsService', () => {
  let service: SoundEffectsService;
  let mockAudio: { play: jasmine.Spy; load: jasmine.Spy; currentTime: number };
  let originalAudio: typeof Audio;

  beforeEach(() => {
    localStorage.removeItem(MUTE_KEY);

    originalAudio = window.Audio;
    mockAudio = { play: jasmine.createSpy('play').and.returnValue(Promise.resolve()), load: jasmine.createSpy('load'), currentTime: 0 };
    (window as unknown as { Audio: unknown }).Audio = jasmine.createSpy('Audio').and.returnValue(mockAudio);

    TestBed.configureTestingModule({});
    service = TestBed.inject(SoundEffectsService);
  });

  afterEach(() => {
    (window as unknown as { Audio: unknown }).Audio = originalAudio;
    localStorage.removeItem(MUTE_KEY);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  // ── isMuted$ ──────────────────────────────────────────────────────────────

  describe('isMuted$', () => {
    it('emits false by default', (done) => {
      service.isMuted$.pipe(take(1)).subscribe(muted => {
        expect(muted).toBeFalse();
        done();
      });
    });

    it('emits true after toggleMute()', (done) => {
      service.toggleMute();
      service.isMuted$.pipe(take(1)).subscribe(muted => {
        expect(muted).toBeTrue();
        done();
      });
    });

    it('toggles back to false on second call', (done) => {
      service.toggleMute();
      service.toggleMute();
      service.isMuted$.pipe(take(1)).subscribe(muted => {
        expect(muted).toBeFalse();
        done();
      });
    });
  });

  // ── localStorage persistence ──────────────────────────────────────────────

  describe('localStorage persistence', () => {
    it('persists muted=true to localStorage', () => {
      service.toggleMute();
      expect(localStorage.getItem(MUTE_KEY)).toBe('true');
    });

    it('persists muted=false to localStorage after two toggles', () => {
      service.toggleMute();
      service.toggleMute();
      expect(localStorage.getItem(MUTE_KEY)).toBe('false');
    });

    it('restores muted=true from localStorage on construction', (done) => {
      localStorage.setItem(MUTE_KEY, 'true');
      const fresh = new SoundEffectsService();
      fresh.isMuted$.pipe(take(1)).subscribe(muted => {
        expect(muted).toBeTrue();
        done();
      });
    });

    it('restores muted=false from localStorage on construction', (done) => {
      localStorage.setItem(MUTE_KEY, 'false');
      const fresh = new SoundEffectsService();
      fresh.isMuted$.pipe(take(1)).subscribe(muted => {
        expect(muted).toBeFalse();
        done();
      });
    });
  });

  // ── play methods ──────────────────────────────────────────────────────────

  describe('playChipsClink()', () => {
    it('calls audio.play() when not muted', () => {
      service.playChipsClink();
      expect(mockAudio.play).toHaveBeenCalled();
    });

    it('does not call audio.play() when muted', () => {
      service.toggleMute();
      service.playChipsClink();
      expect(mockAudio.play).not.toHaveBeenCalled();
    });
  });

  describe('playCardFlip()', () => {
    it('calls audio.play() when not muted', () => {
      service.playCardFlip();
      expect(mockAudio.play).toHaveBeenCalled();
    });

    it('does not call audio.play() when muted', () => {
      service.toggleMute();
      service.playCardFlip();
      expect(mockAudio.play).not.toHaveBeenCalled();
    });
  });

  describe('playWinFanfare()', () => {
    it('calls audio.play() when not muted', () => {
      service.playWinFanfare();
      expect(mockAudio.play).toHaveBeenCalled();
    });

    it('does not call audio.play() when muted', () => {
      service.toggleMute();
      service.playWinFanfare();
      expect(mockAudio.play).not.toHaveBeenCalled();
    });
  });
});
