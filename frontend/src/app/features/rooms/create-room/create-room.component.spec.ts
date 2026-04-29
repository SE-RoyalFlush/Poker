import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { of, throwError } from 'rxjs';
import { RoomService } from '../../../core/services';
import { CreateRoomComponent } from './create-room.component';

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

describe('CreateRoomComponent', () => {
  let component: CreateRoomComponent;
  let fixture: ComponentFixture<CreateRoomComponent>;
  let roomSpy: jasmine.SpyObj<RoomService>;
  let router: Router;

  beforeEach(async () => {
    roomSpy = jasmine.createSpyObj<RoomService>('RoomService', ['createRoom']);
    roomSpy.createRoom.and.returnValue(of(mockRoom));

    await TestBed.configureTestingModule({
      imports: [CreateRoomComponent],
      providers: [
        provideRouter([]),
        provideNoopAnimations(),
        { provide: RoomService, useValue: roomSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(CreateRoomComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    spyOn(router, 'navigate').and.returnValue(Promise.resolve(true));
    fixture.detectChanges();
  });

  it('should have submit button disabled when form is invalid', () => {
    component.createForm.patchValue({ roomName: '' });
    fixture.detectChanges();
    const btn = fixture.nativeElement.querySelector('button[type="submit"]');
    expect(btn.disabled).toBeTrue();
  });

  it('should enable submit when all required fields are filled', () => {
    component.createForm.patchValue({
      roomName: 'My Table',
      maxPlayers: 6,
      smallBlind: 1,
      bigBlind: 2,
      isPrivate: false,
    });
    fixture.detectChanges();
    const btn = fixture.nativeElement.querySelector('button[type="submit"]');
    expect(btn.disabled).toBeFalse();
  });

  it('should call createRoom with form values on submit', () => {
    component.createForm.patchValue({
      roomName: 'My Table',
      maxPlayers: 6,
      smallBlind: 1,
      bigBlind: 2,
      isPrivate: false,
    });

    component.onCreateRoom();

    expect(roomSpy.createRoom).toHaveBeenCalled();
  });

  it('should navigate to /lobby/code on success', () => {
    component.createForm.patchValue({
      roomName: 'My Table',
      maxPlayers: 6,
      smallBlind: 1,
      bigBlind: 2,
      isPrivate: false,
    });

    component.onCreateRoom();

    expect(router.navigate).toHaveBeenCalledWith(['/lobby', mockRoom.code]);
  });

  it('should display createError on API failure', () => {
    roomSpy.createRoom.and.returnValue(
      throwError(() => ({ error: { message: 'Server error' } }))
    );
    component.createForm.patchValue({
      roomName: 'My Table',
      maxPlayers: 6,
      smallBlind: 1,
      bigBlind: 2,
      isPrivate: false,
    });

    component.onCreateRoom();

    expect(component.createError).toBe('Server error');
  });

  it('should require roomPassword when isPrivate is true', () => {
    component.createForm.patchValue({ isPrivate: true });
    component.createForm.markAllAsTouched();
    fixture.detectChanges();

    expect(component.createForm.get('roomPassword')?.hasError('required')).toBeTrue();
  });
});
