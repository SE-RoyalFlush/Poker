import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { EMPTY, TimeoutError, catchError, finalize, switchMap, timeout } from 'rxjs';
import { MatSnackBar } from '@angular/material/snack-bar';

// Angular Material
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { AuthService, RoomService } from '../../core/services';

export type ActivePanel = 'login' | 'register' | null;

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './home.component.html',
  styleUrl: './home.component.scss',
})
export class HomeComponent implements OnInit {

  // ── Panel state ────────────────────────────────────────────────
  activePanel: ActivePanel = null;

  // ── Password visibility ────────────────────────────────────────
  showLoginPassword    = false;
  showRegisterPassword = false;

  // ── Loading flags ──────────────────────────────────────────────
  loginLoading    = false;
  registerLoading = false;
  joinLoading     = false;

  // ── Server-level error messages ────────────────────────────────
  loginError    = '';
  registerError = '';
  joinError     = '';

  // ── Forms ──────────────────────────────────────────────────────
  loginForm!:    FormGroup;
  registerForm!: FormGroup;
  joinForm!:     FormGroup;

  constructor(
    private fb:          FormBuilder,
    private authService: AuthService,
    private roomService: RoomService,
    private router:      Router,
    private snackBar:    MatSnackBar,
  ) {}

  ngOnInit(): void {
    this.buildForms();
  }

  // ── Form builders ──────────────────────────────────────────────
  private buildForms(): void {
    this.loginForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(20)]],
      password: ['', [Validators.required]],
    });

    this.registerForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(20)]],
      password: ['', [Validators.required, Validators.minLength(8)]],
    });

    this.joinForm = this.fb.group({
      roomCode: ['', [
        Validators.required,
        Validators.pattern(/^RF-[A-Z0-9]{4}$/i),
      ]],
    });
  }

  // ── Panel controls ─────────────────────────────────────────────
  openLogin():    void { this.activePanel = 'login';    this.loginError = ''; }
  openRegister(): void { this.activePanel = 'register'; this.registerError = ''; }

  closePanel(event?: MouseEvent): void {
    // Only close if called directly or via backdrop click
    this.activePanel = null;
  }

  scrollToJoin(): void {
    document.getElementById('join-room')?.scrollIntoView({ behavior: 'smooth' });
  }

  // ── Login ──────────────────────────────────────────────────────
  onLogin(): void {
    this.loginError = '';
    if (this.loginForm.invalid) {
      this.loginForm.markAllAsTouched();
      return;
    }

    this.loginLoading = true;
    const { username, password } = this.loginForm.value;

    this.authService.login({ username, password }).subscribe({
      next: () => {
        this.loginLoading = false;
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.loginLoading = false;
        const apiError = err?.error;

        // Map field-level backend errors onto form controls
        if (apiError?.field === 'username') {
          this.loginForm.get('username')?.setErrors({ serverError: apiError.message });
        } else if (apiError?.field === 'password') {
          this.loginForm.get('password')?.setErrors({ serverError: apiError.message });
        } else {
          // Generic fallback (401, network error, etc.)
          this.loginError = apiError?.message ?? 'Invalid credentials. Please try again.';
        }
      },
    });
  }

  // ── Register ───────────────────────────────────────────────────
  onRegister(): void {
    this.registerError = '';
    if (this.registerForm.invalid) {
      this.registerForm.markAllAsTouched();
      return;
    }

    this.registerLoading = true;
    const { username, password } = this.registerForm.value;

    this.authService.register({ username, password }).pipe(
      timeout(8000),
      switchMap(() => this.authService.login({ username, password }).pipe(
        timeout(8000),
        catchError(() => {
          this.registerError = 'Account created. Please log in.';
          this.snackBar.open('Account created successfully. Please log in.', 'Dismiss', {
            duration: 3500,
            panelClass: ['rf-toast', 'rf-toast--success'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
          this.activePanel = null;
          this.registerForm.reset();
          void this.router.navigate(['/login']);
          return EMPTY;
        })
      )),
      finalize(() => {
        this.registerLoading = false;
      })
    ).subscribe({
      next: () => {
        this.snackBar.open('User created successfully. Logged in.', 'Dismiss', {
          duration: 3000,
          panelClass: ['rf-toast', 'rf-toast--success'],
          horizontalPosition: 'right',
          verticalPosition: 'top',
        });
        this.activePanel = null;
        this.registerForm.reset();
        void this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        if (err instanceof TimeoutError) {
          this.registerError = 'Request timed out. Please try again.';
          this.snackBar.open('Request timed out. Please try again.', 'Dismiss', {
            duration: 3500,
            panelClass: ['rf-toast', 'rf-toast--error'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
          return;
        }

        const apiError = err?.error;

        if (apiError?.field === 'username') {
          this.registerForm.get('username')?.setErrors({ serverError: apiError.message });
        } else if (err?.status === 409) {
          this.registerError = 'Username taken';
          this.snackBar.open('Username already exists.', 'Dismiss', {
            duration: 3500,
            panelClass: ['rf-toast', 'rf-toast--error'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
        } else {
          this.registerError = apiError?.message ?? 'Registration failed. Please try again.';
          this.snackBar.open(this.registerError, 'Dismiss', {
            duration: 3500,
            panelClass: ['rf-toast', 'rf-toast--error'],
            horizontalPosition: 'right',
            verticalPosition: 'top',
          });
        }
      },
    });
  }

  // ── Join Room (unauthenticated preview) ────────────────────────
  onJoinRoom(): void {
    this.joinError = '';
    if (this.joinForm.invalid) {
      this.joinForm.markAllAsTouched();
      return;
    }

    const { roomCode } = this.joinForm.value;

    // If the user is not authenticated, send them to login first,
    // then handle the room redirect post-auth via a query param.
    if (!this.authService.isAuthenticated()) {
      this.openLogin();
      return;
    }

    this.joinLoading = true;
    this.roomService.joinRoom(roomCode.toUpperCase()).subscribe({
      next: (room) => {
        this.joinLoading = false;
        this.router.navigate(['/room', room.id]);
      },
      error: (err) => {
        this.joinLoading = false;
        const apiError = err?.error;

        if (apiError?.field === 'roomCode') {
          this.joinForm.get('roomCode')?.setErrors({ serverError: apiError.message });
        } else {
          this.joinError = apiError?.message ?? 'Room not found. Check the code and try again.';
        }
      },
    });
  }
}