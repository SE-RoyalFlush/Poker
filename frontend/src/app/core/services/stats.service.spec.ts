import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { StatsService, UserStats } from './stats.service';
import { API_URL } from '../config/endpoints';

describe('StatsService', () => {
  let service: StatsService;
  let http: HttpTestingController;

  const mockStats: UserStats = {
    handsPlayed: 42,
    wins: 20,
    losses: 22,
    winRate: 47.6,
    totalEarnings: -150.5,
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(StatsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('constructs the correct endpoint URL', () => {
    service.getUserStats('99').subscribe();

    const req = http.expectOne(`${API_URL}/users/99/stats`);
    expect(req.request.method).toBe('GET');
    expect(req.request.withCredentials).toBeTrue();
    req.flush(mockStats);
  });

  it('returns UserStats on success', () => {
    let result: UserStats | undefined;
    service.getUserStats('1').subscribe((s) => (result = s));

    http.expectOne(`${API_URL}/users/1/stats`).flush(mockStats);

    expect(result).toEqual(mockStats);
  });

  it('propagates HTTP errors', () => {
    let err: unknown;
    service.getUserStats('1').subscribe({ error: (e) => (err = e) });

    http.expectOne(`${API_URL}/users/1/stats`).flush('Not Found', {
      status: 404,
      statusText: 'Not Found',
    });

    expect(err).toBeTruthy();
  });
});
