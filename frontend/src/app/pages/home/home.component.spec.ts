import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { AuthService, RoomService, ToastNotificationService } from '../../core/services';
import { HomeComponent } from './home.component';

const mockRoom = {
  id: 'r1',
  code: 'AB12CD',
  name: 'Test Table',
  gameType: 'NLH',
  smallBlind: 1,
  bigBlind: 2,
  maxPlayers: 9,
  currentPlayers: 4,
  isPrivate: false,
  isFull: false,
  seats: 5,
};

describe('HomeComponent', () => {
  let component: HomeComponent;
  let fixture: ComponentFixture<HomeComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let roomSpy: jasmine.SpyObj<RoomService>;
  let toastSpy: jasmine.SpyObj<ToastNotificationService>;
  let router: Router;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj<AuthService>('AuthService', ['login', 'register', 'isAuthenticated']);
    roomSpy = jasmine.createSpyObj<RoomService>('RoomService', ['joinRoom']);
    toastSpy = jasmine.createSpyObj<ToastNotificationService>('ToastNotificationService', ['show', 'dismiss']);

    authSpy.login.and.returnValue(of({ id: 1, username: 'ace' }));
    authSpy.register.and.returnValue(of({ id: 2, username: 'shark' }));
    authSpy.isAuthenticated.and.returnValue(false);

    await TestBed.configureTestingModule({
      imports: [HomeComponent],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authSpy },
        { provide: RoomService, useValue: roomSpy },
        { provide: ToastNotificationService, useValue: toastSpy },
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, 'navigate').and.returnValue(Promise.resolve(true));

    fixture = TestBed.createComponent(HomeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should open login panel', () => {
    component.openLogin();
    expect(component.activePanel).toBe('login');
  });

  it('should call login with username/password payload', () => {
    component.loginForm.setValue({ username: 'ace', password: 'secret123' });
    component.onLogin();

    expect(authSpy.login).toHaveBeenCalledWith({ username: 'ace', password: 'secret123' });
  });

  it('should call register with username/password payload', () => {
    component.registerForm.setValue({ username: 'ace', password: 'secret123' });
    component.onRegister();

    expect(authSpy.register).toHaveBeenCalledWith({
      username: 'ace',
      password: 'secret123',
    });
    expect(authSpy.login).toHaveBeenCalledWith({ username: 'ace', password: 'secret123' });
    expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
  });

  it('should open login panel when unauthenticated user tries to join room', () => {
    component.joinForm.setValue({ roomCode: 'AB12CD' });
    component.onJoinRoom();

    expect(component.activePanel).toBe('login');
    expect(roomSpy.joinRoom).not.toHaveBeenCalled();
  });

  it('should navigate to /lobby/code when an authenticated user joins a room', () => {
    authSpy.isAuthenticated.and.returnValue(true);
    roomSpy.joinRoom.and.returnValue(of(mockRoom));
    component.joinForm.setValue({ roomCode: 'ab12cd' });

    component.onJoinRoom();

    expect(roomSpy.joinRoom).toHaveBeenCalledWith('AB12CD');
    expect(router.navigate).toHaveBeenCalledWith(['/lobby', mockRoom.code]);
  });
});
