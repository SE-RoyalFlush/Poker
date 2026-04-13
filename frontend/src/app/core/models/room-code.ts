import { ValidatorFn, Validators } from '@angular/forms';

export const ROOM_CODE_LENGTH = 6;
export const ROOM_CODE_EXAMPLE = 'AB12CD';
export const ROOM_CODE_PATTERN = /^[A-Z0-9]{6}$/i;

export function roomCodeValidator(): ValidatorFn {
  return Validators.pattern(ROOM_CODE_PATTERN);
}

export function normalizeRoomCode(code: string): string {
  return code.trim().toUpperCase();
}
