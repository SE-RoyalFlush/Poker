import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

const MUTE_KEY = 'rfx_sound_muted';
const SOUND_NAMES = ['chips-clink', 'card-flip', 'win-fanfare'] as const;
type SoundName = typeof SOUND_NAMES[number];

@Injectable({ providedIn: 'root' })
export class SoundEffectsService {
  private readonly mutedSubject: BehaviorSubject<boolean>;
  readonly isMuted$: Observable<boolean>;
  private readonly sounds = new Map<SoundName, HTMLAudioElement>();

  constructor() {
    const stored = localStorage.getItem(MUTE_KEY);
    this.mutedSubject = new BehaviorSubject<boolean>(stored === 'true');
    this.isMuted$ = this.mutedSubject.asObservable();

    document.addEventListener('click', () => this.preload(), { once: true });
  }

  toggleMute(): void {
    const next = !this.mutedSubject.value;
    this.mutedSubject.next(next);
    localStorage.setItem(MUTE_KEY, String(next));
  }

  playChipsClink(): void { this.play('chips-clink'); }
  playCardFlip(): void { this.play('card-flip'); }
  playWinFanfare(): void { this.play('win-fanfare'); }

  private preload(): void {
    SOUND_NAMES.forEach(name => {
      if (!this.sounds.has(name)) {
        const audio = new Audio(`assets/sounds/${name}.mp3`);
        audio.load();
        this.sounds.set(name, audio);
      }
    });
  }

  private play(name: SoundName): void {
    if (this.mutedSubject.value) return;
    let audio = this.sounds.get(name);
    if (!audio) {
      audio = new Audio(`assets/sounds/${name}.mp3`);
      this.sounds.set(name, audio);
    }
    audio.currentTime = 0;
    audio.play().catch(() => undefined);
  }
}
