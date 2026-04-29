import { Component, Input, Output, EventEmitter, OnChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { SoundEffectsService } from '../../../core/services/sound-effects.service';

export type ActionType = 'CHECK' | 'CALL' | 'RAISE' | 'FOLD';

export interface GameAction {
  type: ActionType;
  amount?: number;
}

@Component({
  selector: 'app-game-controls',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatInputModule,
    MatFormFieldModule,
  ],
  templateUrl: './game-controls.component.html',
  styleUrl: './game-controls.component.scss',
})
export class GameControlsComponent implements OnChanges {
  @Input() isActivePlayer = false;
  @Input() callAmount = 0;
  @Input() maxRaise = 0;
  @Output() action = new EventEmitter<GameAction>();

  showRaiseInput = false;
  raiseForm: FormGroup;

  constructor(private fb: FormBuilder, private soundEffects: SoundEffectsService) {
    this.raiseForm = this.fb.group({
      amount: [null, [Validators.required, Validators.min(0)]],
    });
  }

  ngOnChanges(): void {
    this.showRaiseInput = false;
    this.updateRaiseValidators();
  }

  private updateRaiseValidators(): void {
    const min = this.callAmount > 0 ? this.callAmount : 1;
    const validators = [Validators.required, Validators.min(min)];
    if (this.maxRaise > 0) {
      validators.push(Validators.max(this.maxRaise));
    }
    this.raiseForm.get('amount')?.setValidators(validators);
    this.raiseForm.get('amount')?.updateValueAndValidity();
  }

  onCheck(): void {
    if (!this.isActivePlayer) return;
    this.action.emit({ type: 'CHECK' });
  }

  onCall(): void {
    if (!this.isActivePlayer) return;
    this.soundEffects.playChipsClink();
    this.action.emit({ type: 'CALL', amount: this.callAmount });
  }

  onFold(): void {
    if (!this.isActivePlayer) return;
    this.action.emit({ type: 'FOLD' });
  }

  onRaiseToggle(): void {
    if (!this.isActivePlayer) return;
    this.showRaiseInput = !this.showRaiseInput;
    if (this.showRaiseInput) {
      this.raiseForm.reset({ amount: this.callAmount > 0 ? this.callAmount : 1 });
    }
  }

  onRaiseSubmit(): void {
    if (!this.isActivePlayer || this.raiseForm.invalid) return;
    this.soundEffects.playChipsClink();
    this.action.emit({ type: 'RAISE', amount: this.raiseForm.value.amount });
    this.showRaiseInput = false;
    this.raiseForm.reset();
  }

  get raiseMin(): number {
    return this.callAmount > 0 ? this.callAmount : 1;
  }
}
