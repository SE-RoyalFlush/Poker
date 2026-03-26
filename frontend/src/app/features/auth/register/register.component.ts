import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { EMPTY, TimeoutError, catchError, finalize, switchMap, timeout } from 'rxjs';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { AuthService } from '../../../core/services';
import { RegisterData } from '../../../core/models';

function passwordMatchValidator(control: AbstractControl): ValidationErrors | null {
  const password = control.get('password')?.value;
  const confirmPassword = control.get('confirmPassword')?.value;

  if (!password || !confirmPassword) {
    return null;
  }

  return password === confirmPassword ? null : { passwordMismatch: true };
}

@Component({
  selector: 'app-register',
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
  templateUrl: './register.component.html',
  styleUrl: './register.component.scss',
})
export class RegisterComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);
  private readonly snackBar = inject(MatSnackBar);

  readonly registerForm = this.formBuilder.nonNullable.group(
    {
      username: ['', [Validators.required, Validators.minLength(3)]],
      password: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', [Validators.required, Validators.minLength(8)]],
    },
    { validators: passwordMatchValidator }
  );

  isSubmitting = false;
  submitError: string | null = null;

  get usernameControl() {
    return this.registerForm.controls.username;
  }

  get passwordControl() {
    return this.registerForm.controls.password;
  }

  get confirmPasswordControl() {
    return this.registerForm.controls.confirmPassword;
  }

  onSubmit(): void {
    if (this.registerForm.invalid || this.isSubmitting) {
      this.registerForm.markAllAsTouched();
      return;
    }

    this.isSubmitting = true;
    this.submitError = null;

    const formValue = this.registerForm.getRawValue();
    const payload: RegisterData = {
      username: formValue.username,
      password: formValue.password,
      confirmPassword: formValue.confirmPassword,
    };

    this.authService
      .register(payload)
      .pipe(
        timeout(8000),
        switchMap(() =>
          this.authService.login({
            username: payload.username,
            password: payload.password,
          }).pipe(
            timeout(8000),
            catchError(() => {
              this.submitError = 'Account created. Please log in.';
              this.snackBar.open('Account created successfully. Please log in.', 'Dismiss', {
                duration: 3500,
                panelClass: ['rf-toast', 'rf-toast--success'],
                horizontalPosition: 'right',
                verticalPosition: 'top',
              });
              void this.router.navigate(['/login']);
              return EMPTY;
            })
          )
        ),
        finalize(() => (this.isSubmitting = false))
      )
      .subscribe({
        next: () => {
          this.snackBar.open('User created successfully. Logged in.', 'Dismiss', {
            duration: 3000,
            panelClass: ['rf-toast', 'rf-toast--success'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
          this.router.navigate(['/dashboard']);
        },
        error: (error: unknown) => {
          if (error instanceof TimeoutError) {
            this.submitError = 'Request timed out. Please try again.';
            this.snackBar.open(this.submitError, 'Dismiss', {
              duration: 3500,
              panelClass: ['rf-toast', 'rf-toast--error'],
              horizontalPosition: 'right',
              verticalPosition: 'top',
            });
            return;
          }

          if (error instanceof HttpErrorResponse && error.status === 409) {
            this.submitError = 'Username taken';
            this.snackBar.open('Username already exists.', 'Dismiss', {
              duration: 3500,
              panelClass: ['rf-toast', 'rf-toast--error'],
              horizontalPosition: 'right',
              verticalPosition: 'top',
            });
            return;
          }

          this.submitError = 'Unable to create account right now. Please try again.';
          this.snackBar.open(this.submitError, 'Dismiss', {
            duration: 3500,
            panelClass: ['rf-toast', 'rf-toast--error'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
        },
      });
  }
}
