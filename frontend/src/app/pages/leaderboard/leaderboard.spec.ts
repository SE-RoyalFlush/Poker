import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { of, throwError } from 'rxjs';

import { LeaderboardComponent } from './leaderboard';
import { LeaderboardService, LeaderboardEntry } from '../../core/services/leaderboard.service';
import { ToastNotificationService } from '../../core/services/toast.service';

const mockEntries: LeaderboardEntry[] = [
  { rank: 2, userId: 2, username: 'Bob',   handsPlayed: 80,  wins: 40, winRate: 50.0, totalEarnings: 800  },
  { rank: 1, userId: 1, username: 'Alice', handsPlayed: 100, wins: 60, winRate: 60.0, totalEarnings: 1500 },
  { rank: 3, userId: 3, username: 'Carol', handsPlayed: 50,  wins: 20, winRate: 40.0, totalEarnings: 200  },
];

describe('LeaderboardComponent', () => {
  let component: LeaderboardComponent;
  let fixture: ComponentFixture<LeaderboardComponent>;
  let leaderboardService: jasmine.SpyObj<LeaderboardService>;
  let toastService: jasmine.SpyObj<ToastNotificationService>;

  beforeEach(async () => {
    leaderboardService = jasmine.createSpyObj('LeaderboardService', ['getLeaderboard']);
    toastService = jasmine.createSpyObj('ToastNotificationService', ['show']);
    leaderboardService.getLeaderboard.and.returnValue(of(mockEntries));

    await TestBed.configureTestingModule({
      imports: [
        LeaderboardComponent,
        RouterTestingModule,
        NoopAnimationsModule,
        MatSnackBarModule,
      ],
      providers: [
        { provide: LeaderboardService, useValue: leaderboardService },
        { provide: ToastNotificationService, useValue: toastService },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LeaderboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('lodash sorting', () => {
    it('sorts entries by totalEarnings descending on init', () => {
      expect(component.sortedEntries[0].username).toBe('Alice');
      expect(component.sortedEntries[1].username).toBe('Bob');
      expect(component.sortedEntries[2].username).toBe('Carol');
    });

    it('re-sorts by winRate when sort field changes to winRate', () => {
      component.onSortClick('winRate');
      expect(component.activeSortField).toBe('winRate');
      expect(component.sortedEntries[0].winRate).toBeGreaterThanOrEqual(component.sortedEntries[1].winRate);
    });

    it('re-sorts by wins when sort field changes to wins', () => {
      component.onSortClick('wins');
      expect(component.activeSortField).toBe('wins');
      expect(component.sortedEntries[0].wins).toBeGreaterThanOrEqual(component.sortedEntries[1].wins);
    });

    it('cycles sort field on repeated clicks of the same column', () => {
      // starts at totalEarnings
      expect(component.activeSortField).toBe('totalEarnings');
      component.onSortClick('totalEarnings');
      expect(component.activeSortField).toBe('winRate');
      component.onSortClick('winRate');
      expect(component.activeSortField).toBe('wins');
      component.onSortClick('wins');
      expect(component.activeSortField).toBe('totalEarnings');
    });
  });

  describe('top-3 rank classes', () => {
    it('assigns gold class to index 0', () => {
      expect(component.rankClass(0)).toBe('rf-lb__row--gold');
    });

    it('assigns silver class to index 1', () => {
      expect(component.rankClass(1)).toBe('rf-lb__row--silver');
    });

    it('assigns bronze class to index 2', () => {
      expect(component.rankClass(2)).toBe('rf-lb__row--bronze');
    });

    it('assigns no class beyond index 2', () => {
      expect(component.rankClass(3)).toBe('');
      expect(component.rankClass(9)).toBe('');
    });
  });

  describe('isSortActive', () => {
    it('returns true for the active sort field', () => {
      expect(component.isSortActive('totalEarnings')).toBeTrue();
      expect(component.isSortActive('winRate')).toBeFalse();
    });

    it('updates after sort change', () => {
      component.onSortClick('winRate');
      expect(component.isSortActive('winRate')).toBeTrue();
      expect(component.isSortActive('totalEarnings')).toBeFalse();
    });
  });

  describe('API error handling', () => {
    it('shows toast and sets hasError on API failure', fakeAsync(() => {
      leaderboardService.getLeaderboard.and.returnValue(throwError(() => new Error('Server error')));
      component.ngOnInit();
      tick();

      expect(component.hasError).toBeTrue();
      expect(component.isLoading).toBeFalse();
      expect(toastService.show).toHaveBeenCalledWith('Failed to load leaderboard.', 'error');
    }));
  });

  describe('loading state', () => {
    it('sets isLoading to false after data loads', () => {
      expect(component.isLoading).toBeFalse();
    });

    it('populates sortedEntries after successful load', () => {
      expect(component.sortedEntries.length).toBe(3);
    });
  });

  describe('username links', () => {
    it('generates correct profile link for each entry', () => {
      fixture.detectChanges();
      const links = fixture.nativeElement.querySelectorAll('.rf-lb__username-link');
      expect(links.length).toBe(3);
      // First sorted entry is Alice (userId: '1')
      expect(links[0].getAttribute('href')).toContain('/profile/1');
    });
  });
});
