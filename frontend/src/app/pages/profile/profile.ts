import { Component, DestroyRef, inject, OnInit } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';

import { StatsService, UserStats } from '../../core/services/stats.service';
import { ToastNotificationService } from '../../core/services/toast.service';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, RouterLink, MatProgressSpinnerModule, MatButtonModule],
  templateUrl: './profile.html',
  styleUrl: './profile.scss',
})
export class ProfileComponent implements OnInit {
  stats: UserStats | null = null;
  isLoading = false;
  hasError = false;
  userId = '';

  private readonly destroyRef = inject(DestroyRef);

  constructor(
    private route: ActivatedRoute,
    private statsService: StatsService,
    private toast: ToastNotificationService,
  ) {}

  ngOnInit(): void {
    this.userId = this.route.snapshot.paramMap.get('id') ?? '';
    this.loadStats();
  }

  private loadStats(): void {
    if (!this.userId) {
      this.hasError = true;
      return;
    }

    this.isLoading = true;
    this.hasError = false;

    this.statsService.getUserStats(this.userId).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (stats) => {
        this.stats = stats;
        this.isLoading = false;
      },
      error: () => {
        this.hasError = true;
        this.isLoading = false;
        this.toast.show('Failed to load player stats.', 'error');
      },
    });
  }
}
