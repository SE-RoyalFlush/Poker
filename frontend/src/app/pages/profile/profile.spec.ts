import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, ActivatedRoute } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { of, throwError } from 'rxjs';

import { ProfileComponent } from './profile';
import { StatsService, UserStats } from '../../core/services/stats.service';
import { ToastNotificationService } from '../../core/services/toast.service';

const MOCK_STATS: UserStats = {
  handsPlayed: 100,
  wins: 60,
  losses: 40,
  winRate: 60,
  totalEarnings: 350.75,
};

function fakeRoute(id: string): Partial<ActivatedRoute> {
  return {
    snapshot: { paramMap: { get: (_k: string) => id } } as any,
  };
}

describe('ProfileComponent', () => {
  let component: ProfileComponent;
  let fixture: ComponentFixture<ProfileComponent>;
  let statsSpy: jasmine.SpyObj<StatsService>;
  let toastSpy: jasmine.SpyObj<ToastNotificationService>;

  async function createComponent(userId: string, statsObs: jasmine.SpyObj<StatsService>['getUserStats']): Promise<void> {
    await TestBed.configureTestingModule({
      imports: [ProfileComponent],
      providers: [
        provideRouter([]),
        provideNoopAnimations(),
        { provide: StatsService, useValue: statsSpy },
        { provide: ToastNotificationService, useValue: toastSpy },
        { provide: ActivatedRoute, useValue: fakeRoute(userId) },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ProfileComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  beforeEach(() => {
    statsSpy = jasmine.createSpyObj<StatsService>('StatsService', ['getUserStats']);
    toastSpy = jasmine.createSpyObj<ToastNotificationService>('ToastNotificationService', ['show']);
  });

  describe('loading state', () => {
    it('shows spinner when isLoading is true', async () => {
      statsSpy.getUserStats.and.returnValue(of(MOCK_STATS));
      await createComponent('5', statsSpy.getUserStats);

      component.isLoading = true;
      component.stats = null;
      component.hasError = false;
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.rf-profile__loading')).toBeTruthy();
    });
  });

  describe('success state', () => {
    beforeEach(async () => {
      statsSpy.getUserStats.and.returnValue(of(MOCK_STATS));
      await createComponent('5', statsSpy.getUserStats);
    });

    it('renders all stat cards', () => {
      const cards = fixture.nativeElement.querySelectorAll('.rf-stat-card');
      expect(cards.length).toBe(5);
    });

    it('does not show error state', () => {
      expect(fixture.nativeElement.querySelector('.rf-profile__empty')).toBeNull();
    });

    it('sets stats from service', () => {
      expect(component.stats).toEqual(MOCK_STATS);
    });

    it('calls service with correct userId', () => {
      expect(statsSpy.getUserStats).toHaveBeenCalledWith('5');
    });
  });

  describe('error state', () => {
    beforeEach(async () => {
      statsSpy.getUserStats.and.returnValue(throwError(() => new Error('Network error')));
      await createComponent('7', statsSpy.getUserStats);
    });

    it('shows error state', () => {
      expect(fixture.nativeElement.querySelector('.rf-profile__empty')).toBeTruthy();
    });

    it('shows error toast', () => {
      expect(toastSpy.show).toHaveBeenCalledWith('Failed to load player stats.', 'error');
    });

    it('sets hasError to true', () => {
      expect(component.hasError).toBeTrue();
    });
  });
});
