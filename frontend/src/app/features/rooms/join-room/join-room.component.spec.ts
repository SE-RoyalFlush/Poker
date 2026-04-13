import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { of, throwError } from 'rxjs';
import { RoomService } from '../../../core/services';
import { JoinRoomComponent } from './join-room.component';

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

describe('JoinRoomComponent', () => {
  let component: JoinRoomComponent;
  let fixture: ComponentFixture<JoinRoomComponent>;
  let roomSpy: jasmine.SpyObj<RoomService>;
  let router: Router;

  beforeEach(async () => {
    roomSpy = jasmine.createSpyObj<RoomService>('RoomService', ['joinRoom', 'getLiveRooms']);
    roomSpy.getLiveRooms.and.returnValue(of([mockRoom]));
    roomSpy.joinRoom.and.returnValue(of(mockRoom));

    await TestBed.configureTestingModule({
      imports: [JoinRoomComponent],
      providers: [
        provideRouter([]),
        provideNoopAnimations(),
        { provide: RoomService, useValue: roomSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(JoinRoomComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    spyOn(router, 'navigate').and.returnValue(Promise.resolve(true));
    fixture.detectChanges();
  });

  it('should load live rooms on init', () => {
    expect(roomSpy.getLiveRooms).toHaveBeenCalled();
    expect(component.liveRooms.length).toBe(1);
  });

  it('should reject room codes not matching the shared 6-character format', () => {
    component.joinForm.patchValue({ roomCode: 'RF-7742' });
    component.joinForm.markAllAsTouched();

    expect(component.joinForm.get('roomCode')?.hasError('pattern')).toBeTrue();
  });

  it('should accept valid 6-char alphanumeric codes', () => {
    component.joinForm.patchValue({ roomCode: 'AB12CD' });

    expect(component.joinForm.get('roomCode')?.valid).toBeTrue();
  });

  it('should call joinRoom with uppercased code on submit', () => {
    component.joinForm.patchValue({ roomCode: 'ab12cd' });

    component.onJoinRoom();

    expect(roomSpy.joinRoom).toHaveBeenCalledWith('AB12CD', undefined);
  });

  it('should navigate to /room/code on success', () => {
    component.joinForm.patchValue({ roomCode: 'AB12CD' });

    component.onJoinRoom();

    expect(router.navigate).toHaveBeenCalledWith(['/room', mockRoom.code]);
  });

  it('should reveal password field on 403 response', () => {
    roomSpy.joinRoom.and.returnValue(
      throwError(() => ({ status: 403, error: { message: 'Password required.' } }))
    );
    component.joinForm.patchValue({ roomCode: 'AB12CD' });

    component.onJoinRoom();

    expect(component.showJoinPassword).toBeTrue();
    component.joinForm.markAllAsTouched();
    expect(component.joinForm.get('joinPassword')?.hasError('required')).toBeTrue();
  });

  it('should prefill room code and trigger join via quickJoin on open room', () => {
    component.quickJoin(mockRoom);

    expect(component.joinForm.get('roomCode')?.value).toBe(mockRoom.code);
    expect(roomSpy.joinRoom).toHaveBeenCalled();
  });
});
