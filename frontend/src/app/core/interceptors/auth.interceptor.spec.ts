import { TestBed } from '@angular/core/testing';
import {
  provideHttpClient,
  withInterceptorsFromDi,
  HTTP_INTERCEPTORS,
  HttpClient
} from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { Router } from '@angular/router';
import { AuthInterceptor } from './auth.interceptor';
import { CsrfService } from '../services/csrf.service';

describe('AuthInterceptor', () => {
  let httpMock: HttpTestingController;
  let httpClient: HttpClient;
  let csrfService: CsrfService;
  let routerSpy: jasmine.SpyObj<Router>;

  const testUrl = 'http://localhost:8080/api/test';

  beforeEach(() => {
    routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    TestBed.configureTestingModule({
      providers: [
        CsrfService,
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
        {
          provide: HTTP_INTERCEPTORS,
          useClass: AuthInterceptor,
          multi: true
        },
        { provide: Router, useValue: routerSpy }
      ]
    });

    httpMock = TestBed.inject(HttpTestingController);
    httpClient = TestBed.inject(HttpClient);
    csrfService = TestBed.inject(CsrfService);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should attach withCredentials to every request', () => {
    httpClient.get(testUrl).subscribe();
    const req = httpMock.expectOne(testUrl);
    expect(req.request.withCredentials).toBeTrue();
    req.flush({});
  });

  it('should NOT attach X-CSRF-Token on GET requests', () => {
    httpClient.get(testUrl).subscribe();
    const req = httpMock.expectOne(testUrl);
    expect(req.request.headers.has('X-CSRF-Token')).toBeFalse();
    req.flush({});
  });

  it('should attach X-CSRF-Token on POST when token exists', () => {
    spyOn(csrfService, 'getToken').and.returnValue('mock-csrf-token');
    httpClient.post(testUrl, {}).subscribe();
    const req = httpMock.expectOne(testUrl);
    expect(req.request.headers.get('X-CSRF-Token')).toBe('mock-csrf-token');
    req.flush({});
  });

  it('should NOT attach X-CSRF-Token on POST if token is null', () => {
    spyOn(csrfService, 'getToken').and.returnValue(null);
    httpClient.post(testUrl, {}).subscribe();
    const req = httpMock.expectOne(testUrl);
    expect(req.request.headers.has('X-CSRF-Token')).toBeFalse();
    req.flush({});
  });

  it('should attach X-CSRF-Token on PUT, PATCH, DELETE', () => {
    spyOn(csrfService, 'getToken').and.returnValue('mock-csrf-token');

    httpClient.put(testUrl, {}).subscribe();
    const putReq = httpMock.expectOne(testUrl);
    expect(putReq.request.headers.get('X-CSRF-Token')).toBe('mock-csrf-token');
    putReq.flush({});

    httpClient.patch(testUrl, {}).subscribe();
    const patchReq = httpMock.expectOne(testUrl);
    expect(patchReq.request.headers.get('X-CSRF-Token')).toBe('mock-csrf-token');
    patchReq.flush({});

    httpClient.delete(testUrl).subscribe();
    const deleteReq = httpMock.expectOne(testUrl);
    expect(deleteReq.request.headers.get('X-CSRF-Token')).toBe('mock-csrf-token');
    deleteReq.flush({});
  });

  it('should redirect to /login on 401', () => {
    const clearSpy = spyOn(csrfService, 'clearToken');
    httpClient.get(testUrl).subscribe({ error: () => {} });
    const req = httpMock.expectOne(testUrl);
    req.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' });
    expect(clearSpy).toHaveBeenCalled();
    expect(routerSpy.navigate).toHaveBeenCalledWith(['/login']);
  });

  it('should NOT redirect on 401 from session check endpoint /api/me', () => {
    const clearSpy = spyOn(csrfService, 'clearToken');
    httpClient.get('http://localhost:8080/api/me').subscribe({ error: () => {} });
    const req = httpMock.expectOne('http://localhost:8080/api/me');
    req.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' });

    expect(clearSpy).toHaveBeenCalled();
    expect(routerSpy.navigate).not.toHaveBeenCalled();
  });

  it('should NOT redirect on non-401 errors', () => {
    httpClient.get(testUrl).subscribe({ error: () => {} });
    const req = httpMock.expectOne(testUrl);
    req.flush('Error', { status: 500, statusText: 'Server Error' });
    expect(routerSpy.navigate).not.toHaveBeenCalled();
  });
});
