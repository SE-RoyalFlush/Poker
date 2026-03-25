import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { of, throwError } from 'rxjs';
import { AuthService, RoomService } from '../../core/services';
import { DashboardComponent } from './dashboard';

describe('DashboardComponent', () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let roomSpy: jasmine.SpyObj<RoomService>;
  let router: Router;

  const mockRoom = {
    id: 'r1',
    code: 'RF-7742',
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

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj<AuthService>('AuthService', ['getCurrentUser', 'logout']);
    roomSpy = jasmine.createSpyObj<RoomService>('RoomService', ['createRoom', 'joinRoom', 'getLiveRooms']);

    authSpy.getCurrentUser.and.returnValue({ id: 1, username: 'ace' });
    roomSpy.getLiveRooms.and.returnValue(of([mockRoom]));
    roomSpy.createRoom.and.returnValue(of(mockRoom));
    roomSpy.joinRoom.and.returnValue(of(mockRoom));

    await TestBed.configureTestingModule({
      imports: [DashboardComponent],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: authSpy },
        { provide: RoomService, useValue: roomSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    spyOn(router, 'navigate').and.returnValue(Promise.resolve(true));
    fixture.detectChanges();
  });

  it('should load live rooms on init', () => {
    expect(roomSpy.getLiveRooms).toHaveBeenCalled();
    expect(component.liveRooms.length).toBe(1);
  });

  it('should create room when form is valid', () => {
    component.createForm.patchValue({
      roomName: 'My Table',
      maxPlayers: 6,
      smallBlind: 1,
      bigBlind: 2,
      isPrivate: false,
    });

    component.onCreateRoom();

    expect(roomSpy.createRoom).toHaveBeenCalled();
    expect(component.createdRoomCode).toBe('RF-7742');
  });

  it('should reveal join password on 403 join response', () => {
    roomSpy.joinRoom.and.returnValue(throwError(() => ({ status: 403, error: { message: 'Password required.' } })));
    component.joinForm.patchValue({ roomCode: 'RF-7742', joinPassword: '' });

    component.onJoinRoom();

    expect(component.showJoinPassword).toBeTrue();
  });

  it('should logout and navigate to home', () => {
    component.onLogout();

    expect(authSpy.logout).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/']);
  });
});
