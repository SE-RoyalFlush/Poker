import { CommonModule } from '@angular/common';
import { ChangeDetectionStrategy, Component, OnInit, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService, CreateRoomPayload, Room, RoomService, User } from '../../core/services';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
  changeDetection: ChangeDetectionStrategy.Default,
})
export class DashboardComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly authService = inject(AuthService);
  private readonly roomService = inject(RoomService);
  private readonly router = inject(Router);

  currentUser: User | null = null;

  readonly createForm = this.fb.group({
    roomName: ['', [Validators.required, Validators.maxLength(40)]],
    maxPlayers: [6, [Validators.required, Validators.min(2), Validators.max(9)]],
    smallBlind: [1, [Validators.required, Validators.min(0.25)]],
    bigBlind: [2, [Validators.required, Validators.min(0.5)]],
    isPrivate: [false],
    roomPassword: [''],
  });

  readonly joinForm = this.fb.group({
    roomCode: ['', [Validators.required, Validators.pattern(/^RF-[A-Z0-9]{4}$/i)]],
    joinPassword: [''],
  });

  createLoading = false;
  createError = '';
  createdRoomCode = '';

  joinLoading = false;
  joinError = '';
  showJoinPassword = false;
  showJoinPasswordText = false;

  liveRooms: Room[] = [];
  liveRoomsLoading = false;

  ngOnInit(): void {
    this.currentUser = this.authService.getCurrentUser();
    this.bindPrivateRoomValidation();
    this.loadLiveRooms();
  }

  loadLiveRooms(): void {
    this.liveRoomsLoading = true;
    this.roomService.getLiveRooms().subscribe({
      next: (rooms) => {
        this.liveRooms = rooms;
        this.liveRoomsLoading = false;
      },
      error: () => {
        this.liveRoomsLoading = false;
      },
    });
  }

  onCreateRoom(): void {
    this.createError = '';
    this.createdRoomCode = '';

    if (this.createForm.invalid) {
      this.createForm.markAllAsTouched();
      return;
    }

    this.createLoading = true;
    const payload = this.createForm.getRawValue() as CreateRoomPayload;

    this.roomService.createRoom(payload).subscribe({
      next: (room) => {
        this.createLoading = false;
        this.createdRoomCode = room.code;
        this.loadLiveRooms();
      },
      error: (err) => {
        this.createLoading = false;
        const apiError = err?.error;
        if (apiError?.field) {
          this.createForm.get(apiError.field)?.setErrors({ serverError: apiError.message });
          return;
        }

        this.createError = apiError?.message ?? 'Could not create room. Please try again.';
      },
    });
  }

  onJoinRoom(): void {
    this.joinError = '';

    if (this.joinForm.invalid) {
      this.joinForm.markAllAsTouched();
      return;
    }

    const roomCode = this.joinForm.get('roomCode')?.value?.toUpperCase() ?? '';
    const joinPassword = this.joinForm.get('joinPassword')?.value || undefined;

    this.joinLoading = true;
    this.roomService.joinRoom(roomCode, joinPassword).subscribe({
      next: (room) => {
        this.joinLoading = false;
        this.router.navigate(['/table', room.id]);
      },
      error: (err) => {
        this.joinLoading = false;
        const apiError = err?.error;

        if (err?.status === 403 && !this.showJoinPassword) {
          this.showJoinPassword = true;
          const joinPasswordControl = this.joinForm.get('joinPassword');
          joinPasswordControl?.setValidators([Validators.required]);
          joinPasswordControl?.updateValueAndValidity();
          this.joinError = 'This room is password-protected. Enter the password to join.';
          return;
        }

        if (apiError?.field) {
          this.joinForm.get(apiError.field)?.setErrors({ serverError: apiError.message });
          return;
        }

        this.joinError = apiError?.message ?? 'Could not join room. Check your code and try again.';
      },
    });
  }

  quickJoin(room: Room): void {
    this.joinForm.patchValue({ roomCode: room.code });
    if (room.isPrivate) {
      this.showJoinPassword = true;
      const joinPasswordControl = this.joinForm.get('joinPassword');
      joinPasswordControl?.setValidators([Validators.required]);
      joinPasswordControl?.updateValueAndValidity();
      return;
    }

    this.onJoinRoom();
  }

  copyRoomCode(): void {
    if (!this.createdRoomCode) {
      return;
    }

    navigator.clipboard.writeText(this.createdRoomCode).catch(() => {});
  }

  onLogout(): void {
    const result = this.authService.logout();
    if (result && typeof result.subscribe === 'function') {
      result.subscribe({
        complete: () => {
          this.router.navigate(['/']);
        },
      });
      return;
    }

    this.router.navigate(['/']);
  }

  private bindPrivateRoomValidation(): void {
    this.createForm.get('isPrivate')?.valueChanges.subscribe((isPrivate) => {
      const passwordControl = this.createForm.get('roomPassword');
      if (!passwordControl) {
        return;
      }

      if (isPrivate) {
        passwordControl.setValidators([Validators.required, Validators.minLength(4)]);
      } else {
        passwordControl.clearValidators();
        passwordControl.setValue('');
      }

      passwordControl.updateValueAndValidity();
    });
  }
}
