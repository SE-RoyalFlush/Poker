import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import orderBy from 'lodash-es/orderBy';

import { LeaderboardService, LeaderboardEntry } from '../../core/services/leaderboard.service';
import { ToastNotificationService } from '../../core/services/toast.service';

type SortField = 'totalEarnings' | 'winRate' | 'wins';
const SORT_CYCLE: SortField[] = ['totalEarnings', 'winRate', 'wins'];

@Component({
  selector: 'app-leaderboard',
  standalone: true,
  imports: [CommonModule, RouterLink, MatProgressSpinnerModule, MatButtonModule],
  templateUrl: './leaderboard.html',
  styleUrl: './leaderboard.scss',
})
export class LeaderboardComponent implements OnInit {
  entries: LeaderboardEntry[] = [];
  sortedEntries: LeaderboardEntry[] = [];
  activeSortField: SortField = 'totalEarnings';
  isLoading = false;
  hasError = false;

  constructor(
    private leaderboardService: LeaderboardService,
    private toast: ToastNotificationService,
  ) {}

  ngOnInit(): void {
    this.loadLeaderboard();
  }

  private loadLeaderboard(): void {
    this.isLoading = true;
    this.hasError = false;

    this.leaderboardService.getLeaderboard().subscribe({
      next: (entries) => {
        this.entries = entries;
        this.sortedEntries = this.applySorting(entries, this.activeSortField);
        this.isLoading = false;
      },
      error: () => {
        this.hasError = true;
        this.isLoading = false;
        this.toast.show('Failed to load leaderboard.', 'error');
      },
    });
  }

  onSortClick(field: SortField): void {
    if (this.activeSortField === field) {
      const idx = SORT_CYCLE.indexOf(field);
      this.activeSortField = SORT_CYCLE[(idx + 1) % SORT_CYCLE.length];
    } else {
      this.activeSortField = field;
    }
    this.sortedEntries = this.applySorting(this.entries, this.activeSortField);
  }

  private applySorting(entries: LeaderboardEntry[], field: SortField): LeaderboardEntry[] {
    return orderBy(entries, [field], ['desc']);
  }

  rankClass(index: number): string {
    if (index === 0) return 'rf-lb__row--gold';
    if (index === 1) return 'rf-lb__row--silver';
    if (index === 2) return 'rf-lb__row--bronze';
    return '';
  }

  isSortActive(field: SortField): boolean {
    return this.activeSortField === field;
  }
}
