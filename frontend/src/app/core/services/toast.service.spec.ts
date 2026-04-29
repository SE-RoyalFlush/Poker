import { TestBed, fakeAsync, tick } from '@angular/core/testing';
import { MatSnackBar, MatSnackBarRef, TextOnlySnackBar } from '@angular/material/snack-bar';
import { Subject } from 'rxjs';
import { ToastNotificationService, Toast, ToastType } from './toast.service';

function makeSnackBarRefStub(): {
  ref: MatSnackBarRef<TextOnlySnackBar>;
  dismissed$: Subject<void>;
} {
  const dismissed$ = new Subject<void>();
  const ref = {
    afterDismissed: () => dismissed$.asObservable(),
    dismiss: jasmine.createSpy('dismiss'),
  } as unknown as MatSnackBarRef<TextOnlySnackBar>;
  return { ref, dismissed$ };
}

describe('ToastNotificationService', () => {
  let service: ToastNotificationService;
  let snackBarSpy: jasmine.SpyObj<MatSnackBar>;
  let stubDismissed$: Subject<void>;
  let stubRef: MatSnackBarRef<TextOnlySnackBar>;

  beforeEach(() => {
    const { ref, dismissed$ } = makeSnackBarRefStub();
    stubRef = ref;
    stubDismissed$ = dismissed$;

    snackBarSpy = jasmine.createSpyObj<MatSnackBar>('MatSnackBar', ['open']);
    snackBarSpy.open.and.returnValue(stubRef);

    TestBed.configureTestingModule({
      providers: [
        ToastNotificationService,
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    });

    service = TestBed.inject(ToastNotificationService);
  });

  // ── Creation ─────────────────────────────────────────────────────
  it('should create', () => {
    expect(service).toBeTruthy();
  });

  it('should start with an empty toasts array', () => {
    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));
    expect(toasts.length).toBe(0);
  });

  // ── show() ────────────────────────────────────────────────────────
  it('should add a toast to toasts$ when show() is called', () => {
    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));

    service.show('Hello world', 'info');

    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Hello world');
    expect(toasts[0].type).toBe('info');
  });

  it('should return a non-empty id from show()', () => {
    const id = service.show('msg', 'success');
    expect(id).toBeTruthy();
  });

  it('should open MatSnackBar with rf-toast panel class', () => {
    service.show('msg', 'info');
    expect(snackBarSpy.open).toHaveBeenCalledWith(
      'msg', 'Dismiss',
      jasmine.objectContaining({ panelClass: jasmine.arrayContaining(['rf-toast']) })
    );
  });

  it('should include rf-toast--success panel class for success type', () => {
    service.show('ok', 'success');
    const config = snackBarSpy.open.calls.mostRecent().args[2] as { panelClass: string[] };
    expect(config.panelClass).toContain('rf-toast--success');
  });

  it('should include rf-toast--error panel class for error type', () => {
    service.show('fail', 'error');
    const config = snackBarSpy.open.calls.mostRecent().args[2] as { panelClass: string[] };
    expect(config.panelClass).toContain('rf-toast--error');
  });

  it('should include rf-toast--warning panel class for warning type', () => {
    service.show('warn', 'warning');
    const config = snackBarSpy.open.calls.mostRecent().args[2] as { panelClass: string[] };
    expect(config.panelClass).toContain('rf-toast--warning');
  });

  it('should include rf-toast--info panel class for info type', () => {
    service.show('info', 'info');
    const config = snackBarSpy.open.calls.mostRecent().args[2] as { panelClass: string[] };
    expect(config.panelClass).toContain('rf-toast--info');
  });

  it('should use default duration of 3500ms when not specified', () => {
    service.show('msg', 'info');
    const config = snackBarSpy.open.calls.mostRecent().args[2] as { duration: number };
    expect(config.duration).toBe(3500);
  });

  it('should use custom duration when provided', () => {
    service.show('msg', 'info', 1000);
    const config = snackBarSpy.open.calls.mostRecent().args[2] as { duration: number };
    expect(config.duration).toBe(1000);
  });

  it('should use horizontalPosition right and verticalPosition top', () => {
    service.show('msg', 'info');
    const config = snackBarSpy.open.calls.mostRecent().args[2] as {
      horizontalPosition: string;
      verticalPosition: string;
    };
    expect(config.horizontalPosition).toBe('right');
    expect(config.verticalPosition).toBe('top');
  });

  // ── auto-dismiss ──────────────────────────────────────────────────
  it('should remove toast from toasts$ after snackbar dismisses', () => {
    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));

    service.show('auto-dismiss me', 'info');
    expect(toasts.length).toBe(1);

    stubDismissed$.next();
    expect(toasts.length).toBe(0);
  });

  // ── dismiss(id) ───────────────────────────────────────────────────
  it('should remove only the matching toast by id when dismiss() is called', () => {
    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));

    const id = service.show('first', 'info');
    expect(toasts.length).toBe(1);

    service.dismiss(id);
    expect(toasts.length).toBe(0);
  });

  it('should call MatSnackBarRef.dismiss() when dismiss() is called', () => {
    const id = service.show('msg', 'success');
    service.dismiss(id);
    expect((stubRef.dismiss as jasmine.Spy).calls.count()).toBe(1);
  });

  it('should default type to info when not specified', () => {
    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));

    service.show('default type');
    expect(toasts[0].type).toBe('info');
  });

  // ── multiple toasts ───────────────────────────────────────────────
  it('should accumulate multiple toasts in toasts$', () => {
    const { ref: ref2, dismissed$: d2 } = makeSnackBarRefStub();
    snackBarSpy.open.and.returnValues(stubRef, ref2);

    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));

    service.show('first', 'info');
    service.show('second', 'success');

    expect(toasts.length).toBe(2);

    stubDismissed$.next();
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('second');
  });

  // ── toast shape ───────────────────────────────────────────────────
  it('should include all required Toast fields', () => {
    let toasts: Toast[] = [];
    service.toasts$.subscribe(t => (toasts = t));

    service.show('check shape', 'error', 2000);
    const t = toasts[0];

    expect(t.id).toBeTruthy();
    expect(t.message).toBe('check shape');
    expect(t.type).toBe('error');
    expect(t.duration).toBe(2000);
  });
});
