import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { Room, RoomService } from '../../../core/services';

@Component({
  selector: 'app-join-room',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './join-room.component.html',
  styleUrl: './join-room.component.scss',
})
export class JoinRoomComponent implements OnInit {
  joinForm!: FormGroup;
  joinLoading         = false;
  joinError           = '';
  showJoinPassword    = false;
  showJoinPasswordText = false;

  liveRooms:       Room[] = [];
  liveRoomsLoading = false;

  private fb          = inject(FormBuilder);
  private roomService = inject(RoomService);
  private router      = inject(Router);

  ngOnInit(): void {
    this.joinForm = this.fb.group({
      roomCode:     ['', [Validators.required, Validators.pattern(/^[A-Z0-9]{6}$/i)]],
      joinPassword: [''],
    });
    this.loadLiveRooms();
  }

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
        this.router.navigate(['/room', room.code]);
      },
      error: (err) => {
        this.joinLoading = false;
        const apiError   = err?.error;

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

  quickJoin(room: Room): void {
    this.joinForm.patchValue({ roomCode: room.code });
    if (room.isPrivate) {
      this.showJoinPassword = true;
    } else {
      this.onJoinRoom();
    }
  }
}
