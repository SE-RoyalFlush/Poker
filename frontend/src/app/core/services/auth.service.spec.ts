import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { AuthService } from './auth.service';
import { User } from '../models';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  const apiUrl = 'http://localhost:8080/api';

  const mockUser: User = {
    id: 1,
    username: 'testuser',
    createdAt: '2026-03-02T00:00:00Z',
    updatedAt: '2026-03-02T00:00:00Z'
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [AuthService]
    });

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    // Verify no outstanding HTTP requests
    httpMock.verify();
  });

  describe('checkSession()', () => {
    it('should fetch and update currentUser$ when session is valid', (done) => {
      service.checkSession().subscribe(user => {
        expect(user).toEqual(mockUser);
      });

      // Verify the HTTP request was made correctly
      const req = httpMock.expectOne(`${apiUrl}/me`);
      expect(req.request.method).toBe('GET');
      expect(req.request.withCredentials).toBe(true);
      req.flush(mockUser);

      // Verify currentUser$ observable was updated
      service.currentUser$.subscribe(user => {
        expect(user).toEqual(mockUser);
        done();
      });
    });

    it('should set currentUser$ to null when session is invalid (401)', (done) => {
      service.checkSession().subscribe(user => {
        expect(user).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/me`);
      req.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' });

      service.currentUser$.subscribe(user => {
        expect(user).toBeNull();
        done();
      });
    });

    it('should set isLoading$ to true during request and false after', (done) => {
      const loadingStates: boolean[] = [];

      service.isLoading$.subscribe(isLoading => {
        loadingStates.push(isLoading);
      });

      service.checkSession().subscribe(() => {
        // After response
        setTimeout(() => {
          expect(loadingStates).toContain(true);
          expect(loadingStates[loadingStates.length - 1]).toBe(false);
          done();
        }, 0);
      });

      const req = httpMock.expectOne(`${apiUrl}/me`);
      req.flush(mockUser);
    });
  });

  describe('login()', () => {
    it('should post credentials and fetch user profile on success', (done) => {
      const credentials = { username: 'testuser', password: 'password123' };

      service.login(credentials).subscribe(() => {
        // Verify currentUser$ was updated
        service.currentUser$.subscribe(user => {
          expect(user).toEqual(mockUser);
          done();
        });
      });

      // First request: POST to /login
      const loginReq = httpMock.expectOne(`${apiUrl}/login`);
      expect(loginReq.request.method).toBe('POST');
      expect(loginReq.request.body).toEqual(credentials);
      expect(loginReq.request.withCredentials).toBe(true);
      loginReq.flush(null); // No body returned from login

      // Second request: GET to /me (from checkSession)
      const meReq = httpMock.expectOne(`${apiUrl}/me`);
      expect(meReq.request.method).toBe('GET');
      meReq.flush(mockUser);
    });

    it('should throw error on failed login', (done) => {
      const credentials = { username: 'testuser', password: 'wrongpassword' };

      service.login(credentials).subscribe(
        () => {
          fail('should have thrown error');
        },
        (error) => {
          expect(error).toBeTruthy();
          done();
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/login`);
      req.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' });
    });

    it('should NOT look for token in response body', (done) => {
      const credentials = { username: 'testuser', password: 'password123' };

      service.login(credentials).subscribe();

      const loginReq = httpMock.expectOne(`${apiUrl}/login`);
      // Response body is intentionally empty/null
      loginReq.flush(null);

      const meReq = httpMock.expectOne(`${apiUrl}/me`);
      meReq.flush(mockUser);

      // Verify no token extraction attempt
      service.currentUser$.subscribe(user => {
        expect(user).toEqual(mockUser);
        expect((user as any).token).toBeUndefined();
        done();
      });
    });
  });

  describe('register()', () => {
    it('should post registration data and fetch user profile on success', (done) => {
      const registerData = {
        username: 'newuser',
        password: 'password123',
        confirmPassword: 'password123'
      };

      service.register(registerData).subscribe(user => {
        expect(user).toEqual(mockUser);
        done();
      });

      // First request: POST to /register
      const registerReq = httpMock.expectOne(`${apiUrl}/register`);
      expect(registerReq.request.method).toBe('POST');
      expect(registerReq.request.body).toEqual(registerData);
      registerReq.flush(null);

      // Second request: GET to /me (from checkSession)
      const meReq = httpMock.expectOne(`${apiUrl}/me`);
      meReq.flush(mockUser);
    });

    it('should throw error on failed registration', (done) => {
      const registerData = {
        username: 'testuser',
        password: 'password123',
        confirmPassword: 'password123'
      };

      service.register(registerData).subscribe(
        () => {
          fail('should have thrown error');
        },
        (error) => {
          expect(error).toBeTruthy();
          done();
        }
      );

      const req = httpMock.expectOne(`${apiUrl}/register`);
      req.flush('Username already exists', { status: 409, statusText: 'Conflict' });
    });
  });

  describe('logout()', () => {
    it('should post to logout endpoint and clear currentUser$', (done) => {
      // First set a user
      service['currentUserSubject'].next(mockUser);

      service.logout().subscribe(() => {
        service.currentUser$.subscribe(user => {
          expect(user).toBeNull();
          done();
        });
      });

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      expect(req.request.method).toBe('POST');
      req.flush(null);
    });

    it('should clear currentUser$ even if logout fails', (done) => {
      // First set a user
      service['currentUserSubject'].next(mockUser);

      service.logout().subscribe(() => {
        service.currentUser$.subscribe(user => {
          expect(user).toBeNull();
          done();
        });
      });

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      // Simulate backend error
      req.flush('Server error', { status: 500, statusText: 'Internal Server Error' });
    });
  });

  describe('getCurrentUser()', () => {
    it('should return current user synchronously', () => {
      service['currentUserSubject'].next(mockUser);
      const user = service.getCurrentUser();
      expect(user).toEqual(mockUser);
    });

    it('should return null when not authenticated', () => {
      const user = service.getCurrentUser();
      expect(user).toBeNull();
    });
  });

  describe('isAuthenticated()', () => {
    it('should return true when user is logged in', () => {
      service['currentUserSubject'].next(mockUser);
      expect(service.isAuthenticated()).toBe(true);
    });

    it('should return false when user is not authenticated', () => {
      expect(service.isAuthenticated()).toBe(false);
    });
  });
});

