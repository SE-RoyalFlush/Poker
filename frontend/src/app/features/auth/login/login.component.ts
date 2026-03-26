import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { EMPTY, TimeoutError, catchError, finalize, timeout } from 'rxjs';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { AuthService } from '../../../core/services';
import { LoginCredentials } from '../../../core/models';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss',
})
export class LoginComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);
  private readonly snackBar = inject(MatSnackBar);

  readonly loginForm = this.formBuilder.nonNullable.group({
    username: ['', [Validators.required, Validators.minLength(3)]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  isSubmitting = false;
  submitError: string | null = null;

  get usernameControl() {
    return this.loginForm.controls.username;
  }

  get passwordControl() {
    return this.loginForm.controls.password;
  }

  onSubmit(): void {
    if (this.loginForm.invalid || this.isSubmitting) {
      this.loginForm.markAllAsTouched();
      return;
    }

    this.isSubmitting = true;
    this.submitError = null;

    const formValue = this.loginForm.getRawValue();
    const payload: LoginCredentials = {
      username: formValue.username,
      password: formValue.password,
    };

    this.authService
      .login(payload)
      .pipe(
        timeout(8000),
        catchError((error: unknown) => {
          if (error instanceof TimeoutError) {
            this.submitError = 'Request timed out. Please try again.';
            this.snackBar.open(this.submitError, 'Dismiss', {
              duration: 3500,
              panelClass: ['rf-toast', 'rf-toast--error'],
              horizontalPosition: 'right',
              verticalPosition: 'top',
            });
            return EMPTY;
          }

          if (error instanceof HttpErrorResponse && error.status === 401) {
            this.submitError = 'Invalid username or password.';
            this.snackBar.open('Invalid username or password.', 'Dismiss', {
              duration: 3500,
              panelClass: ['rf-toast', 'rf-toast--error'],
              horizontalPosition: 'right',
              verticalPosition: 'top',
            });
            return EMPTY;
          }

          this.submitError = 'Server error. Please try again later.';
          this.snackBar.open(this.submitError, 'Dismiss', {
            duration: 3500,
            panelClass: ['rf-toast', 'rf-toast--error'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
          return EMPTY;
        }),
        finalize(() => (this.isSubmitting = false))
      )
      .subscribe({
        next: () => {
          void this.router.navigate(['/home']);
        },
      });
  }
}