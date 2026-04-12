import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { of } from 'rxjs';
import { AuthService, RoomService } from '../../core/services';
import { DashboardComponent } from './dashboard';

describe('DashboardComponent', () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let roomSpy: jasmine.SpyObj<RoomService>;
  let router: Router;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj<AuthService>('AuthService', ['getCurrentUser', 'logout']);
    roomSpy = jasmine.createSpyObj<RoomService>('RoomService', ['createRoom', 'joinRoom', 'getLiveRooms']);

    authSpy.getCurrentUser.and.returnValue({ id: 1, username: 'ace' });
    authSpy.logout.and.returnValue(of(undefined));
    roomSpy.getLiveRooms.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [DashboardComponent],
      providers: [
        provideRouter([]),
        provideNoopAnimations(),
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

  it('should render create-room and join-room components', () => {
    const el = fixture.nativeElement;
    expect(el.querySelector('app-create-room')).toBeTruthy();
    expect(el.querySelector('app-join-room')).toBeTruthy();
  });

  it('should logout and navigate to home', () => {
    component.onLogout();

    expect(authSpy.logout).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/']);
  });
});
