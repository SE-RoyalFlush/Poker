import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { CsrfService } from './csrf.service';

describe('CsrfService', () => {
  let service: CsrfService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        CsrfService,
        provideHttpClient(),
        provideHttpClientTesting()
      ]
    });

    service = TestBed.inject(CsrfService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should start with a null token', () => {
    expect(service.getToken()).toBeNull();
  });

  it('should fetch and store the CSRF token', () => {
    const mockToken = 'test-csrf-token-123';

    service.fetchToken().subscribe(response => {
      expect(response.csrfToken).toBe(mockToken);
    });

    const req = httpMock.expectOne('http://localhost:8080/api/csrf');
    expect(req.request.method).toBe('GET');
    expect(req.request.withCredentials).toBeTrue();
    req.flush({ csrfToken: mockToken });

    expect(service.getToken()).toBe(mockToken);
  });

  it('should clear the token', () => {
    service.fetchToken().subscribe();
    httpMock.expectOne('http://localhost:8080/api/csrf')
      .flush({ csrfToken: 'some-token' });

    expect(service.getToken()).toBe('some-token');
    service.clearToken();
    expect(service.getToken()).toBeNull();
  });

  it('ensureToken() returns existing token without extra request', () => {
    service.fetchToken().subscribe();
    httpMock.expectOne('http://localhost:8080/api/csrf')
      .flush({ csrfToken: 'some-token' });

    service.ensureToken().subscribe(token => {
      expect(token).toBe('some-token');
    });

    httpMock.expectNone('http://localhost:8080/api/csrf');
  });

  it('ensureToken() fetches token if none exists', () => {
    service.ensureToken().subscribe(token => {
      expect(token).toBe('fetched-token');
    });

    const req = httpMock.expectOne('http://localhost:8080/api/csrf');
    expect(req.request.method).toBe('GET');
    req.flush({ csrfToken: 'fetched-token' });

    expect(service.getToken()).toBe('fetched-token');
  });
});
