import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap, catchError, of, switchMap } from 'rxjs';
import { User, LoginCredentials, RegisterData } from '../models';

/**
 * AuthService handles all authentication-related operations.
 * It serves as the single source of truth for the current user state.
 *
 * Key responsibilities:
 * - Manage user authentication state via BehaviorSubject
 * - Communicate with backend auth APIs (/login, /register, /me)
 * - Provide observables for components to subscribe to
 * - Handle session management
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly apiUrl = 'http://localhost:8080/api';

  // BehaviorSubject holds the current user state
  // - null: user is not authenticated (guest)
  // - User object: user is authenticated
  private readonly currentUserSubject = new BehaviorSubject<User | null>(null);

  // Exposed as a read-only observable for components to subscribe to
  // Components CANNOT modify this directly; must use login/logout/checkSession methods
  public readonly currentUser$: Observable<User | null> = this.currentUserSubject.asObservable();

  // Track loading state for UI indicators
  private readonly isLoadingSubject = new BehaviorSubject<boolean>(false);
  public readonly isLoading$: Observable<boolean> = this.isLoadingSubject.asObservable();

  constructor(private http: HttpClient) {}

  /**
   * Check if user has an active session on app initialization.
   * Called from app.ts ngOnInit to restore user state.
   *
   * Flow:
   * 1. Calls GET /api/me (backend validates session cookie)
   * 2. If successful: stores user data in currentUserSubject
   * 3. If fails (401): user remains null (guest state)
   * 4. Returns the user or null
   *
   * @returns Observable of the current user (or null if not authenticated)
   */
  checkSession(): Observable<User | null> {
    this.isLoadingSubject.next(true);

    return this.http.get<User>(`${this.apiUrl}/me`, {
      withCredentials: true
    }).pipe(
      tap(user => {
        this.currentUserSubject.next(user);
        this.isLoadingSubject.next(false);
      }),
      catchError(error => {
        // 401 or any error means user is not authenticated
        console.error('Session check failed:', error);
        this.currentUserSubject.next(null);
        this.isLoadingSubject.next(false);
        return of(null); // Return null instead of throwing error
      })
    );
  }

  /**
   * Authenticate user with credentials.
   *
   * Flow:
   * 1. POST credentials to /api/login
   * 2. Backend validates and sets HttpOnly session cookie
   * 3. On success: immediately calls checkSession() to fetch and store user profile
   * 4. On failure: error is thrown and caught by caller
   *
   * IMPORTANT: Token is NOT returned in response body.
   * Authentication is managed via HttpOnly cookies set by the backend.
   *
   * @param credentials - username and password
   * @returns Observable of the logged-in user
   */
  login(credentials: LoginCredentials): Observable<User | null> {
    this.isLoadingSubject.next(true);

    return this.http.post<void>(`${this.apiUrl}/login`, credentials, {
      withCredentials: true
    }).pipe(
      switchMap(() => {
        // After successful login, fetch user profile
        return this.checkSession();
      }),
      catchError(error => {
        console.error('Login failed:', error);
        this.isLoadingSubject.next(false);
        throw error;
      })
    );
  }

  /**
   * Register a new user account.
   *
   * Flow:
   * 1. POST registration data to /api/register
   * 2. Backend validates and creates new user
   * 3. On success: automatically logs in the user and calls checkSession()
   * 4. On failure: error is thrown
   *
   * @param data - username, password, and confirmPassword
   * @returns Observable of newly registered and logged-in user
   */
  register(data: RegisterData): Observable<User | null> {
    this.isLoadingSubject.next(true);

    return this.http.post<void>(`${this.apiUrl}/register`, data, {
      withCredentials: true
    }).pipe(
      switchMap(() => {
        // After successful registration, fetch user profile
        return this.checkSession();
      }),
      catchError(error => {
        console.error('Registration failed:', error);
        this.isLoadingSubject.next(false);
        throw error;
      })
    );
  }

  /**
   * Log out the current user.
   *
   * Flow:
   * 1. POST to /api/logout (backend clears session cookie)
   * 2. Clear currentUser BehaviorSubject (set to null)
   * 3. This triggers all subscribers to be notified of logout
   *
   * @returns Observable that completes when logout is done
   */
  logout(): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/logout`, {}, {
      withCredentials: true
    }).pipe(
      tap(() => {
        // Clear local user state
        this.currentUserSubject.next(null);
        this.isLoadingSubject.next(false);
      }),
      catchError(error => {
        console.error('Logout failed:', error);
        // Even if logout fails on backend, clear local state
        this.currentUserSubject.next(null);
        this.isLoadingSubject.next(false);
        return of(void 0); // Don't throw, just clean up locally
      })
    );
  }

  /**
   * Get the current user synchronously (for guards and quick checks).
   * This is useful for synchronous checks in route guards.
   *
   * @returns Current user object or null if not authenticated
   */
  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  /**
   * Check if user is authenticated (synchronous).
   * Used in route guards to protect routes.
   *
   * @returns true if user is logged in, false otherwise
   */
  isAuthenticated(): boolean {
    return this.currentUserSubject.value !== null;
  }
}

