import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { LeaderboardService, LeaderboardEntry } from './leaderboard.service';
import { API_URL } from '../config/endpoints';

describe('LeaderboardService', () => {
  let service: LeaderboardService;
  let http: HttpTestingController;

  const mockEntries: LeaderboardEntry[] = [
    { rank: 1, userId: 1, username: 'Alice', handsPlayed: 100, wins: 60, winRate: 60.0, totalEarnings: 1500 },
    { rank: 2, userId: 2, username: 'Bob',   handsPlayed: 80,  wins: 40, winRate: 50.0, totalEarnings: 800  },
    { rank: 3, userId: 3, username: 'Carol', handsPlayed: 50,  wins: 20, winRate: 40.0, totalEarnings: 200  },
  ];

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(LeaderboardService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('calls the correct endpoint URL', () => {
    service.getLeaderboard().subscribe();

    const req = http.expectOne(`${API_URL}/leaderboard`);
    expect(req.request.method).toBe('GET');
    expect(req.request.withCredentials).toBeTrue();
    req.flush(mockEntries);
  });

  it('returns LeaderboardEntry array on success', () => {
    let result: LeaderboardEntry[] | undefined;
    service.getLeaderboard().subscribe((entries) => (result = entries));

    http.expectOne(`${API_URL}/leaderboard`).flush(mockEntries);

    expect(result).toEqual(mockEntries);
    expect(result?.length).toBe(3);
  });

  it('propagates HTTP errors', () => {
    let err: unknown;
    service.getLeaderboard().subscribe({ error: (e) => (err = e) });

    http.expectOne(`${API_URL}/leaderboard`).flush('Internal Server Error', {
      status: 500,
      statusText: 'Internal Server Error',
    });

    expect(err).toBeTruthy();
  });
});
