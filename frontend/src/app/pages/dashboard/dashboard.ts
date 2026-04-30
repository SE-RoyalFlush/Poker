import { Component, DestroyRef, inject, OnInit } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';

import { MatButtonModule } from '@angular/material/button';

import { AuthService, User } from '../../core/services';
import { CreateRoomComponent } from '../../features/rooms/create-room/create-room.component';
import { JoinRoomComponent } from '../../features/rooms/join-room/join-room.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatButtonModule,
    CreateRoomComponent,
    JoinRoomComponent,
  ],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class DashboardComponent implements OnInit {
  currentUser: User | null = null;

  private readonly destroyRef = inject(DestroyRef);

  constructor(
    private authService: AuthService,
    private router:      Router,
  ) {}

  ngOnInit(): void {
    this.currentUser = this.authService.getCurrentUser();
  }

  onLogout(): void {
    this.authService.logout().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      complete: () => {
        this.router.navigate(['/']);
      },
    });
  }
}
