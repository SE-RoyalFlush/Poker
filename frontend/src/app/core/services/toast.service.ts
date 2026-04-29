import { Injectable } from '@angular/core';
import { MatSnackBar, MatSnackBarRef, TextOnlySnackBar } from '@angular/material/snack-bar';
import { BehaviorSubject, Observable } from 'rxjs';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
  duration: number;
}

const TYPE_PANEL_CLASS: Record<ToastType, string | null> = {
  success: 'rf-toast--success',
  error:   'rf-toast--error',
  info:    'rf-toast--info',
  warning: 'rf-toast--warning',
};

@Injectable({ providedIn: 'root' })
export class ToastNotificationService {
  private readonly toastsSubject = new BehaviorSubject<Toast[]>([]);
  readonly toasts$: Observable<Toast[]> = this.toastsSubject.asObservable();

  private activeRef: MatSnackBarRef<TextOnlySnackBar> | null = null;

  constructor(private readonly snackBar: MatSnackBar) {}

  show(message: string, type: ToastType = 'info', duration = 3500): string {
    const id = Math.random().toString(36).slice(2, 9);
    const toast: Toast = { id, message, type, duration };

    this.toastsSubject.next([...this.toastsSubject.value, toast]);

    const panelClass: string[] = ['rf-toast'];
    const typeClass = TYPE_PANEL_CLASS[type];
    if (typeClass) panelClass.push(typeClass);

    this.activeRef = this.snackBar.open(message, 'Dismiss', {
      duration,
      panelClass,
      horizontalPosition: 'right',
      verticalPosition: 'top',
    });

    this.activeRef.afterDismissed().subscribe(() => this.removeToast(id));

    return id;
  }

  dismiss(id: string): void {
    this.removeToast(id);
    this.activeRef?.dismiss();
    this.activeRef = null;
  }

  private removeToast(id: string): void {
    this.toastsSubject.next(this.toastsSubject.value.filter(t => t.id !== id));
  }
}
