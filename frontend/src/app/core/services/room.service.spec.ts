import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { RoomService } from './room.service';

describe('RoomService', () => {
  let service: RoomService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [RoomService],
    });

    service = TestBed.inject(RoomService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create room and compute seats when missing', () => {
    service
      .createRoom({
        roomName: 'My Table',
        maxPlayers: 9,
        smallBlind: 1,
        bigBlind: 2,
        isPrivate: false,
      })
      .subscribe((room) => {
        expect(room.code).toBe('RF-1234');
        expect(room.seats).toBe(6);
      });

    const req = httpMock.expectOne('http://localhost:8080/api/rooms');
    expect(req.request.method).toBe('POST');
    expect(req.request.withCredentials).toBeTrue();
    req.flush({
      id: 'r1',
      code: 'RF-1234',
      name: 'My Table',
      gameType: 'NLH',
      smallBlind: 1,
      bigBlind: 2,
      maxPlayers: 9,
      currentPlayers: 3,
      isPrivate: false,
      isFull: false,
    });
  });

  it('should call join endpoint with password when provided', () => {
    service.joinRoom('RF-1234', 'abcd').subscribe();

    const req = httpMock.expectOne('http://localhost:8080/api/rooms/join');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ code: 'RF-1234', password: 'abcd' });
    expect(req.request.withCredentials).toBeTrue();
    req.flush({
      id: 'r1',
      code: 'RF-1234',
      name: 'My Table',
      gameType: 'NLH',
      smallBlind: 1,
      bigBlind: 2,
      maxPlayers: 9,
      currentPlayers: 4,
      isPrivate: true,
      isFull: false,
      seats: 5,
    });
  });

  it('should fetch live rooms and normalize seats', () => {
    service.getLiveRooms().subscribe((rooms) => {
      expect(rooms.length).toBe(2);
      expect(rooms[0].seats).toBe(1);
      expect(rooms[1].seats).toBe(4);
    });

    const req = httpMock.expectOne('http://localhost:8080/api/rooms?status=open');
    expect(req.request.method).toBe('GET');
    expect(req.request.withCredentials).toBeTrue();
    req.flush([
      {
        id: 'r1',
        code: 'RF-1111',
        name: 'A',
        gameType: 'NLH',
        smallBlind: 1,
        bigBlind: 2,
        maxPlayers: 9,
        currentPlayers: 8,
        isPrivate: false,
        isFull: false,
      },
      {
        id: 'r2',
        code: 'RF-2222',
        name: 'B',
        gameType: 'PLO',
        smallBlind: 2,
        bigBlind: 4,
        maxPlayers: 6,
        currentPlayers: 2,
        isPrivate: true,
        isFull: false,
        seats: 4,
      },
    ]);
  });
});
