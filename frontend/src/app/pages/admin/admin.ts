import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { finalize, timeout } from 'rxjs';

import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { AdminService, AdminUser } from '../../core/services';

@Component({
  selector: 'app-admin',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './admin.html',
  styleUrl: './admin.scss',
})
export class AdminComponent {
  private readonly fb = inject(FormBuilder);

  readonly credentialsForm = this.fb.nonNullable.group({
    username: ['', [Validators.required]],
    password: ['', [Validators.required]],
  });

  users: AdminUser[] = [];
  isAuthenticating = false;
  isLoadingUsers = false;
  hasLoadedUsers = false;
  loginError = '';
  loadError = '';
  deleteError = '';
  deletingUserIds = new Set<number>();

  constructor(
    private adminService: AdminService
  ) {}

  onAdminLogin(): void {
    this.loginError = '';
    this.loadError = '';
    this.deleteError = '';

    if (this.credentialsForm.invalid) {
      this.credentialsForm.markAllAsTouched();
      return;
    }

    this.isAuthenticating = true;
    const { username, password } = this.credentialsForm.getRawValue();

    this.adminService.getUsers(username, password).pipe(
      timeout(8000),
      finalize(() => {
        this.isAuthenticating = false;
      })
    ).subscribe({
      next: (users) => {
        this.users = users;
        this.hasLoadedUsers = true;
      },
      error: (error: unknown) => {
        this.users = [];
        this.hasLoadedUsers = false;
        this.loginError = this.mapAdminError(error, 'Invalid admin credentials.');
      },
    });
  }

  refreshUsers(): void {
    this.loadError = '';

    if (this.credentialsForm.invalid) {
      this.credentialsForm.markAllAsTouched();
      return;
    }

    this.isLoadingUsers = true;
    const { username, password } = this.credentialsForm.getRawValue();

    this.adminService.getUsers(username, password).pipe(
      timeout(8000),
      finalize(() => {
        this.isLoadingUsers = false;
      })
    ).subscribe({
      next: (users) => {
        this.users = users;
        this.hasLoadedUsers = true;
      },
      error: (error: unknown) => {
        this.loadError = this.mapAdminError(error, 'Unable to load users.');
      },
    });
  }

  deleteUser(userId: number): void {
    this.deleteError = '';

    if (this.credentialsForm.invalid) {
      this.credentialsForm.markAllAsTouched();
      return;
    }

    const { username, password } = this.credentialsForm.getRawValue();
    this.deletingUserIds.add(userId);

    this.adminService.deleteUser(username, password, userId).subscribe({
      next: () => {
        this.users = this.users.filter((user) => user.id !== userId);
        this.deletingUserIds.delete(userId);
      },
      error: (error: unknown) => {
        this.deletingUserIds.delete(userId);
        this.deleteError = this.mapAdminError(error, 'Failed to delete user.');
      },
    });
  }

  isDeleting(userId: number): boolean {
    return this.deletingUserIds.has(userId);
  }

  private mapAdminError(error: unknown, fallback: string): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 404) {
        return 'Admin endpoint unavailable. Restart backend server to load new admin routes.';
      }
      if (typeof error.error === 'string' && error.error.trim().length > 0) {
        return error.error;
      }
      return error.error?.message ?? fallback;
    }

    return fallback;
  }
}
