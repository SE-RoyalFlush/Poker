import { Component, DestroyRef, OnInit, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { CreateRoomPayload, RoomService } from '../../../core/services';

@Component({
  selector: 'app-create-room',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './create-room.component.html',
  styleUrl: './create-room.component.scss',
})
export class CreateRoomComponent implements OnInit {
  createForm!: FormGroup;
  createLoading = false;
  createError = '';

  private readonly destroyRef  = inject(DestroyRef);
  private fb          = inject(FormBuilder);
  private roomService = inject(RoomService);
  private router      = inject(Router);

  ngOnInit(): void {
    this.createForm = this.fb.group({
      roomName:     ['', [Validators.required, Validators.maxLength(40)]],
      maxPlayers:   [9,  [Validators.required, Validators.min(2), Validators.max(9)]],
      smallBlind:   [1,  [Validators.required, Validators.min(0.25)]],
      bigBlind:     [2,  [Validators.required, Validators.min(0.50)]],
      isPrivate:    [false],
      roomPassword: [''],
    });

    this.createForm.get('isPrivate')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe((isPrivate: boolean) => {
      const pwCtrl = this.createForm.get('roomPassword')!;
      if (isPrivate) {
        pwCtrl.setValidators([Validators.required, Validators.minLength(4)]);
      } else {
        pwCtrl.clearValidators();
        pwCtrl.setValue('');
      }
      pwCtrl.updateValueAndValidity();
    });
  }

  onCreateRoom(): void {
    this.createError = '';

    if (this.createForm.invalid) {
      this.createForm.markAllAsTouched();
      return;
    }

    this.createLoading = true;
    const payload: CreateRoomPayload = this.createForm.value;

    this.roomService.createRoom(payload).subscribe({
      next: (room) => {
        this.createLoading = false;
        this.router.navigate(['/lobby', room.code]);
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
}
