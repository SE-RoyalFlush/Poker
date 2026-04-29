import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { provideRouter, Router } from '@angular/router';
import { of, throwError } from 'rxjs';
import { AuthService, ToastNotificationService } from '../../../core/services';
import { RegisterComponent } from './register.component';

describe('RegisterComponent', () => {
  let fixture: ComponentFixture<RegisterComponent>;
  let component: RegisterComponent;
  let authServiceSpy: jasmine.SpyObj<AuthService>;
  let toastSpy: jasmine.SpyObj<ToastNotificationService>;
  let router: Router;

  beforeEach(async () => {
    authServiceSpy = jasmine.createSpyObj<AuthService>('AuthService', ['register', 'login']);
    toastSpy = jasmine.createSpyObj<ToastNotificationService>('ToastNotificationService', ['show', 'dismiss']);

    await TestBed.configureTestingModule({
      imports: [RegisterComponent, NoopAnimationsModule],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authServiceSpy },
        { provide: ToastNotificationService, useValue: toastSpy },
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, 'navigate').and.returnValue(Promise.resolve(true));

    fixture = TestBed.createComponent(RegisterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should disable submit button while form is invalid', () => {
    const submitButton = fixture.nativeElement.querySelector('button[type="submit"]') as HTMLButtonElement;
    expect(component.registerForm.invalid).toBeTrue();
    expect(submitButton.disabled).toBeTrue();

    component.registerForm.setValue({
      username: 'player-one',
      password: 'StrongPass1',
      confirmPassword: 'StrongPass1',
    });
    fixture.detectChanges();

    expect(component.registerForm.valid).toBeTrue();
    expect(submitButton.disabled).toBeFalse();
  });

  it('should display username taken message when backend responds with 409', () => {
    authServiceSpy.register.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 409, statusText: 'Conflict' }))
    );

    component.registerForm.setValue({
      username: 'existing-user',
      password: 'StrongPass1',
      confirmPassword: 'StrongPass1',
    });

    component.onSubmit();
    fixture.detectChanges();

    expect(component.submitError).toBe('Username taken');
    const errorNode = fixture.nativeElement.querySelector('.submit-error') as HTMLElement;
    expect(errorNode.textContent).toContain('Username taken');
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('should auto-login and redirect to /dashboard after successful registration', () => {
    authServiceSpy.register.and.returnValue(
      of({
        id: 5,
        username: 'new-player',
      })
    );
    authServiceSpy.login.and.returnValue(
      of({
        id: 5,
        username: 'new-player',
      })
    );

    component.registerForm.setValue({
      username: 'new-player',
      password: 'StrongPass1',
      confirmPassword: 'StrongPass1',
    });

    component.onSubmit();

    expect(authServiceSpy.register).toHaveBeenCalled();
    expect(authServiceSpy.login).toHaveBeenCalledWith({
      username: 'new-player',
      password: 'StrongPass1',
    });
    expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
  });

  it('should redirect to /login with message when auto-login fails after registration', () => {
    authServiceSpy.register.and.returnValue(
      of({
        id: 6,
        username: 'new-player-2',
      })
    );
    authServiceSpy.login.and.returnValue(
      throwError(() => new HttpErrorResponse({ status: 401, statusText: 'Unauthorized' }))
    );

    component.registerForm.setValue({
      username: 'new-player-2',
      password: 'StrongPass1',
      confirmPassword: 'StrongPass1',
    });

    component.onSubmit();

    expect(component.submitError).toBe('Account created. Please log in.');
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });

  it('should call register flow when submit button is clicked with valid form', () => {
    authServiceSpy.register.and.returnValue(
      of({
        id: 9,
        username: 'click-user',
      })
    );
    authServiceSpy.login.and.returnValue(
      of({
        id: 9,
        username: 'click-user',
      })
    );

    component.registerForm.setValue({
      username: 'click-user',
      password: 'StrongPass1',
      confirmPassword: 'StrongPass1',
    });
    fixture.detectChanges();

    const submitButton = fixture.nativeElement.querySelector('button[type="submit"]') as HTMLButtonElement;
    submitButton.click();

    expect(authServiceSpy.register).toHaveBeenCalledWith({
      username: 'click-user',
      password: 'StrongPass1',
      confirmPassword: 'StrongPass1',
    });
  });
});
