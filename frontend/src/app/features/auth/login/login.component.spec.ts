import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { provideRouter, Router } from '@angular/router';
import { EMPTY, of, Subject, throwError } from 'rxjs';
import { AuthService, ToastNotificationService } from '../../../core/services';
import { LoginComponent } from './login.component';

describe('LoginComponent', () => {
  let fixture: ComponentFixture<LoginComponent>;
  let component: LoginComponent;
  let authServiceSpy: jasmine.SpyObj<AuthService>;
  let toastSpy: jasmine.SpyObj<ToastNotificationService>;
  let router: Router;

  beforeEach(async () => {
    authServiceSpy = jasmine.createSpyObj<AuthService>('AuthService', ['login']);
    toastSpy = jasmine.createSpyObj<ToastNotificationService>('ToastNotificationService', ['show', 'dismiss']);

    await TestBed.configureTestingModule({
      imports: [LoginComponent, NoopAnimationsModule],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authServiceSpy },
        { provide: ToastNotificationService, useValue: toastSpy },
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, 'navigate').and.returnValue(Promise.resolve(true));

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  // ── Form validation ──────────────────────────────────────────────

  it('should initialise with an invalid form', () => {
    expect(component.loginForm.invalid).toBeTrue();
  });

  it('should disable the submit button while the form is invalid', () => {
    const btn = fixture.nativeElement.querySelector('button[type="submit"]') as HTMLButtonElement;
    expect(btn.disabled).toBeTrue();
  });

  it('should enable the submit button when the form is valid', () => {
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    fixture.detectChanges();
    const btn = fixture.nativeElement.querySelector('button[type="submit"]') as HTMLButtonElement;
    expect(btn.disabled).toBeFalse();
  });

  it('should require username (minlength 3)', () => {
    component.loginForm.setValue({ username: 'ab', password: 'password123' });
    expect(component.usernameControl.hasError('minlength')).toBeTrue();
  });

  it('should require password (minlength 8)', () => {
    component.loginForm.setValue({ username: 'player1', password: 'short' });
    expect(component.passwordControl.hasError('minlength')).toBeTrue();
  });

  it('should not call authService.login when form is submitted invalid', () => {
    component.onSubmit();
    expect(authServiceSpy.login).not.toHaveBeenCalled();
  });

  it('should mark all controls as touched when invalid form is submitted', () => {
    component.onSubmit();
    expect(component.usernameControl.touched).toBeTrue();
    expect(component.passwordControl.touched).toBeTrue();
  });

  // ── Successful login ─────────────────────────────────────────────

  it('should navigate to /home on successful login', () => {
    authServiceSpy.login.and.returnValue(of({ id: 1, username: 'player1' }));
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(router.navigate).toHaveBeenCalledWith(['/home']);
  });

  it('should call authService.login with the form values', () => {
    authServiceSpy.login.and.returnValue(of({ id: 1, username: 'player1' }));
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(authServiceSpy.login).toHaveBeenCalledWith({
      username: 'player1',
      password: 'password123',
    });
  });

  it('should clear submitError on a new submission attempt', () => {
    component.submitError = 'old error';
    authServiceSpy.login.and.returnValue(of({ id: 1, username: 'player1' }));
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(component.submitError).toBeNull();
  });

  // ── 401 error ────────────────────────────────────────────────────

  it('should set submitError to invalid credentials message on 401', () => {
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 401, statusText: 'Unauthorized' }))
    );
    component.loginForm.setValue({ username: 'player1', password: 'wrongpassword' });
    component.onSubmit();
    expect(component.submitError).toBe('Invalid username or password.');
  });

  it('should show error toast on 401', () => {
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 401, statusText: 'Unauthorized' }))
    );
    component.loginForm.setValue({ username: 'player1', password: 'wrongpassword' });
    component.onSubmit();
    expect(toastSpy.show).toHaveBeenCalledWith('Invalid username or password.', 'error');
  });

  it('should not navigate on 401', () => {
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 401, statusText: 'Unauthorized' }))
    );
    component.loginForm.setValue({ username: 'player1', password: 'wrongpassword' });
    component.onSubmit();
    expect(router.navigate).not.toHaveBeenCalled();
  });

  // ── Server error (non-401) ────────────────────────────────────────

  it('should set submitError to server error message on 500', () => {
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 500, statusText: 'Internal Server Error' }))
    );
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(component.submitError).toBe('Server error. Please try again later.');
  });

  it('should show error toast on server error', () => {
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 500, statusText: 'Internal Server Error' }))
    );
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(toastSpy.show).toHaveBeenCalledWith('Server error. Please try again later.', 'error');
  });

  // ── Timeout handling ─────────────────────────────────────────────

  it('should set submitError to timeout message on TimeoutError', fakeAsync(() => {
    const pending$ = new Subject<never>();
    authServiceSpy.login.and.returnValue(pending$);
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();

    tick(8001);

    expect(component.submitError).toBe('Request timed out. Please try again.');
  }));

  // ── isSubmitting flag ────────────────────────────────────────────

  it('should set isSubmitting to false after login completes', () => {
    authServiceSpy.login.and.returnValue(of({ id: 1, username: 'player1' }));
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(component.isSubmitting).toBeFalse();
  });

  it('should set isSubmitting to false after login fails', () => {
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 401 }))
    );
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(component.isSubmitting).toBeFalse();
  });

  it('should not call authService.login again while isSubmitting is true', () => {
    const pending$ = new Subject<never>();
    authServiceSpy.login.and.returnValue(pending$);
    component.loginForm.setValue({ username: 'player1', password: 'password123' });
    component.onSubmit();
    expect(component.isSubmitting).toBeTrue();
    component.onSubmit();
    expect(authServiceSpy.login).toHaveBeenCalledTimes(1);
  });
});
