import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

// Angular Material
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { AuthService, CreateRoomPayload, Room, RoomService, User } from '../../core/services';

@Component({
  selector: 'app-dashboard',
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
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class DashboardComponent implements OnInit {

  // ── Auth ───────────────────────────────────────────────────────
  currentUser: User | null = null;

  // ── Create Room ────────────────────────────────────────────────
  createForm!:      FormGroup;
  createLoading   = false;
  createError     = '';
  createdRoomCode = '';          // shown after successful creation

  // ── Join Room ──────────────────────────────────────────────────
  joinForm!:            FormGroup;
  joinLoading         = false;
  joinError           = '';
  showJoinPassword    = false;   // revealed when backend returns 403 requiring password
  showJoinPasswordText = false;

  // ── Live Rooms ─────────────────────────────────────────────────
  liveRooms:       Room[] = [];
  liveRoomsLoading = false;

  constructor(
    private fb:          FormBuilder,
    private authService: AuthService,
    private roomService: RoomService,
    private router:      Router,
  ) {}

  ngOnInit(): void {
    this.currentUser = this.authService.getCurrentUser();
    this.buildForms();
    this.loadLiveRooms();
  }

  // ── Form builders ──────────────────────────────────────────────
  private buildForms(): void {
    this.createForm = this.fb.group({
      roomName:     ['', [Validators.required, Validators.maxLength(40)]],
      maxPlayers:   [9,  [Validators.required, Validators.min(2), Validators.max(9)]],
      smallBlind:   [1,  [Validators.required, Validators.min(0.25)]],
      bigBlind:     [2,  [Validators.required, Validators.min(0.50)]],
      isPrivate:    [false],
      roomPassword: [''],
    });

    // Dynamically add/remove password validation when isPrivate toggles
    this.createForm.get('isPrivate')?.valueChanges.subscribe((isPrivate: boolean) => {
      const pwCtrl = this.createForm.get('roomPassword')!;
      if (isPrivate) {
        pwCtrl.setValidators([Validators.required, Validators.minLength(4)]);
      } else {
        pwCtrl.clearValidators();
        pwCtrl.setValue('');
      }
      pwCtrl.updateValueAndValidity();
    });

    this.joinForm = this.fb.group({
      roomCode:     ['', [Validators.required, Validators.pattern(/^RF-[A-Z0-9]{4}$/i)]],
      joinPassword: [''],
    });
  }

  // ── Live rooms ─────────────────────────────────────────────────
  loadLiveRooms(): void {
    this.liveRoomsLoading = true;
    this.roomService.getLiveRooms().subscribe({
      next: (rooms) => {
        this.liveRooms        = rooms;
        this.liveRoomsLoading = false;
      },
      error: () => {
        this.liveRoomsLoading = false;
      },
    });
  }

  // ── Create Room ────────────────────────────────────────────────
  onCreateRoom(): void {
    this.createError     = '';
    this.createdRoomCode = '';

    if (this.createForm.invalid) {
      this.createForm.markAllAsTouched();
      return;
    }

    this.createLoading = true;
    const payload: CreateRoomPayload = this.createForm.value;

    this.roomService.createRoom(payload).subscribe({
      next: (room) => {
        this.createLoading   = false;
        this.createdRoomCode = room.code;
        this.loadLiveRooms(); // refresh list
      },
      error: (err) => {
        this.createLoading = false;
        const apiError     = err?.error;

        if (apiError?.field) {
          this.createForm.get(apiError.field)?.setErrors({ serverError: apiError.message });
        } else {
          this.createError = apiError?.message ?? 'Could not create room. Please try again.';
        }
      },
    });
  }

  copyRoomCode(): void {
    navigator.clipboard.writeText(this.createdRoomCode).catch(() => {});
  }

  // ── Join Room ──────────────────────────────────────────────────
  onJoinRoom(): void {
    this.joinError = '';

    if (this.joinForm.invalid) {
      this.joinForm.markAllAsTouched();
      return;
    }

    const { roomCode, joinPassword } = this.joinForm.value;
    this.joinLoading = true;

    this.roomService.joinRoom(roomCode.toUpperCase(), joinPassword || undefined).subscribe({
      next: (room) => {
        this.joinLoading = false;
        this.router.navigate(['/room', room.id]);
      },
      error: (err) => {
        this.joinLoading = false;
        const apiError   = err?.error;

        // 403 = room requires a password
        if (err?.status === 403 && !this.showJoinPassword) {
          this.showJoinPassword = true;
          this.joinForm.get('joinPassword')?.setValidators([Validators.required]);
          this.joinForm.get('joinPassword')?.updateValueAndValidity();
          this.joinError = 'This room is password-protected. Enter the password to join.';
          return;
        }

        if (apiError?.field === 'roomCode') {
          this.joinForm.get('roomCode')?.setErrors({ serverError: apiError.message });
        } else if (apiError?.field === 'joinPassword') {
          this.joinForm.get('joinPassword')?.setErrors({ serverError: apiError.message });
        } else {
          this.joinError = apiError?.message ?? 'Could not join room. Check your code and try again.';
        }
      },
    });
  }

  // Click on a live room row → pre-fill code and attempt join
  quickJoin(room: Room): void {
    this.joinForm.patchValue({ roomCode: room.code });
    if (room.isPrivate) {
      this.showJoinPassword = true;
    } else {
      this.onJoinRoom();
    }
  }

  // ── Logout ─────────────────────────────────────────────────────
  onLogout(): void {
    this.authService.logout();
    this.router.navigate(['/']);
  }
}