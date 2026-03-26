# Sprint 2 Retrospective: Authentication & User Journey

**Sprint Duration:** Sprint 2
**Team Focus:** End-to-End Authentication (Login, Register, Session Management)
**Date:** March 2026

---

## Executive Summary

Sprint 2 focused on completing the core authentication journey. We successfully implemented the backend APIs for login, logout, and session verification, and integrated them with the Angular frontend. This sprint marks the completion of the first "vertical slice" of the application, moving from a static scaffold to a functional system with persistent user sessions.

**Key Achievements:**
- Implemented secure, cookie-based authentication (Session ID) in the backend.
- Created `AuthService` as the reactive source of truth for user state in the frontend.
- Built production-ready `LoginComponent` and `RegisterComponent` with Material Design.
- Configured `AuthInterceptor` for automatic CSRF token injection and global 401 handling.
- Established `AuthGuard` to protect private routes like the Dashboard.
- Achieved a complete end-to-end flow: Registration → Login → Protected Dashboard → Logout.

**Completion Rate:**
- Backend: 2/2 planned issues completed.
- Frontend: 4/4 planned issues completed.
- Overall: 6/6 planned issues completed (100% of Sprint 2 scope).

---

## What We Achieved

### Phase 1: Backend Authentication Core (Issue #1-5)

We moved beyond simple registration to a full session-based authentication system.

**Implementation Details:**
- **Session Management:** Switched to `securecookie` for signed and encrypted session cookies.
- **Login API (`POST /api/login`):** Validates credentials against Bcrypt hashes and issues a `HttpOnly`, `Lax`, `Secure` (in prod) cookie.
- **Logout API (`POST /api/logout`):** Clears the session cookie by setting its `MaxAge` to -1.
- **Session Check API (`GET /api/me`):** Decodes the session cookie and returns the user object if valid. Returns `204 No Content` for unauthenticated requests to support silent frontend checks.

**Key Files:**
- `backend/pkg/api/auth.go` — Handlers for login, logout, and me.

---

### Phase 2: Frontend Core HTTP & Interceptor (Issue #1-7)

To ensure secure communication, we centralized our request logic.

**Implementation Details:**
- **AuthInterceptor:** Automatically attaches `withCredentials: true` to every request.
- **CSRF Protection:** Interceptor extracts the CSRF token from `CsrfService` and injects it into the `X-CSRF-Token` header for mutating requests (POST, PUT, etc.).
- **Global Error Handling:** Automatically redirects to `/login` if a `401 Unauthorized` is received (excluding the session check itself).

**Key Files:**
- `frontend/src/app/core/interceptors/auth.interceptor.ts`

---

### Phase 3: Authentication Service (Issue #1-8)

We implemented the reactive state management for the user's identity.

**Implementation Details:**
- Created `AuthService` using `BehaviorSubject<User | null>` to track the current user.
- Implemented `checkSession()` to restore user state on app load.
- Wrapped login/register/logout calls to update the local state automatically.
- Provided an `isAuthenticated()` synchronous check for route guards.

**Key Files:**
- `frontend/src/app/core/services/auth.service.ts`

---

### Phase 4: User Registration & Login UI (Issue #1-9, #1-10)

The UI was transformed from placeholders to functional Material Design components.

**Implementation Details:**
- **Reactive Forms:** Used `FormBuilder` with robust validators (Required, MinLength).
- **Material UI:** Integrated `MatFormField`, `MatInput`, and `MatProgressSpinner`.
- **Error Handling:** Specific handling for 409 (Username Taken) and 401 (Invalid Credentials) with UI feedback.
- **Success Flow:** Redirects to `/home` or `/login` upon success.

**Key Files:**
- `frontend/src/app/features/auth/register/`
- `frontend/src/app/pages/login/`

---

### Phase 5: Protected Routes & Guard (Issue #1-12)

We secured the application's internal pages.

**Implementation Details:**
- Implemented `authGuard` using the `AuthService.isAuthenticated()` state.
- Applied the guard to `/dashboard` and other protected routes in `app.routes.ts`.
- Integrated `checkSession()` in `app.config.ts` via `APP_INITIALIZER` to prevent "flicker" on page refresh.

---

## What Went Well

- **Vertical Slice Success:** Completing the full auth journey gave the team a clear sense of progress and a "real" application feel.
- **Security by Default:** The combination of `HttpOnly` cookies and CSRF headers provides a high security baseline without complicating the component code.
- **Reactive State:** `AuthService` with RxJS made it easy to sync the UI across the toolbar, login page, and guards.
- **Consistency:** Following the structure established in Sprint 1 allowed us to move much faster in Sprint 2.

---

## Lessons Learned

1. **Cookie-Based Auth Complexity:** Dealing with `withCredentials` and CORS requires careful configuration. We spent significant time debugging why cookies weren't sent before realizing the interceptor needed to clone the request.
2. **Silent Session Checks:** Using `204 No Content` for `/api/me` was a great choice — it prevents error logs in the console during the initial "Who am I?" check on the landing page.
3. **Form Validation UX:** Reactive forms are powerful but require boilerplate. We should consider creating a shared `ValidationService` or custom components to reduce repetitive HTML for error messages.

---

## Technical Debt Register

1. **Form Error Repetition:** Same validation logic repeated in Login and Register.
2. **Missing E2E Tests:** We have unit tests, but a full Playwright/Cypress suite for the login flow would be beneficial.
3. **Loading States:** Many buttons still need the `[disabled]="isLoading$ | async"` binding to prevent double submissions.

---

## Sprint 3 Backlog (Carried Forward from Sprint Report)

### 1. Authentication & Session Endpoint Refinement

Refine the `/api/me` endpoint to ensure it is fully protected and returns the exact data structure required by the frontend dashboard.

**Requirements:**
- Hide sensitive fields (e.g., `PasswordHash`) using struct tags (`json:"-"`)
- Ensure `CreatedAt` is formatted in ISO 8601
- Align backend response contract with frontend expectations (Username, User ID, Join Date)

**Test Outline:**
- Data Integrity: JSON response matches TypeScript interface exactly
- Performance: Endpoint avoids unnecessary database latency

---

### 2. Frontend Dashboard Implementation

Replace the placeholder dashboard UI with a fully functional, protected dashboard that displays authenticated user details using the existing authentication state.

**Scope:**
- Replace placeholder UI
- Bind dashboard to `AuthService.currentUser$`
- Render: Username, User ID, Join Date (formatted from `createdAt`)
- Handle loading state and unauthenticated fallback
- Ensure responsive design (desktop + mobile)
- No additional backend calls

**Acceptance Criteria:**
- `/dashboard` displays Username, User ID, and Join Date
- Uses existing `AuthService` state — no direct `/api/me` calls from component
- Shows loading state during session restoration
- Shows fallback UI if user is unauthenticated

**Relevant Files:**
- `frontend/src/app/pages/dashboard/dashboard.ts`
- `frontend/src/app/pages/dashboard/dashboard.html`
- `frontend/src/app/pages/dashboard/dashboard.scss`
- `frontend/src/app/core/models/user.model.ts`
- `frontend/src/app/core/services/auth.service.ts`

---

### 3. Room Management Data Model

Define the Room data structure using GORM. A room represents a game instance with a unique invitation code and associated metadata.

**Requirements:**
- Unique 6-character alphanumeric room code
- Host ID (creator reference)
- Room status: `WAITING`, `PLAYING`, `FINISHED`
- Maximum player capacity

**Checklist:**
- Create `Room` struct in `pkg/models` with fields: `Code`, `HostID`, `Status`, `MaxPlayers`
- Add `rooms` table to GORM AutoMigration in `main.go`
- Implement `GenerateRoomCode()` utility
- Define status constants

**Dependencies:**
- Blocks: Room Management APIs (HTTP), Lobby Logic & Broadcasting
- Blocked by: None

---

### 4. Core Authentication System — JWT Migration

Migrate session-based auth to JWT stored in HttpOnly cookies.

**Features:**
- Login: Issue JWT in HttpOnly cookie
- Logout: Clear cookie
- Session Validation: `/api/me`

**JWT Claims:** `UserID`, `Username`, `Exp`

**Checklist:**
- Implement `LoginHandler`, `LogoutHandler`, `MeHandler`, `AuthMiddleware`
- Ensure no tokens appear in JSON response body

**Test Outline:**
- Cookie Security: Verify `HttpOnly`, `Secure`, `SameSite=Strict`
- Session Validation: No cookie → `401 Unauthorized`; Valid cookie → `200 OK`
- Logout Flow: Subsequent `/api/me` call fails after logout

---

## Detailed Backend API Documentation

> Full OpenAPI spec is maintained in `docs/api/openapi.yaml`

### Authentication & Session Endpoints

#### `POST /api/register`
- **Description:** Creates a new user account with a secure bcrypt-hashed password.
- **Request Body:**
  ```json
  {
    "username": "string (min 3 chars)",
    "password": "string (min 6 chars)"
  }
  ```
- **Responses:**
  - `201 Created`: User successfully registered (returns User object without password hash).
  - `400 Bad Request`: Validation failure or invalid JSON.
  - `409 Conflict`: Username already exists.
  - `500 Internal Server Error`: Database failure or hashing error.

#### `POST /api/login`
- **Description:** Authenticates user and issues a secure session cookie.
- **Request Body:**
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Responses:**
  - `204 No Content`: Login successful. Sets `session-id` HttpOnly cookie.
  - `400 Bad Request`: Missing username or password.
  - `401 Unauthorized`: Invalid credentials.
  - `500 Internal Server Error`: Database or session creation error.

#### `POST /api/logout`
- **Description:** Terminates the current user session.
- **Responses:**
  - `204 No Content`: Session cleared. Cookie `MaxAge` set to -1.

#### `GET /api/me`
- **Description:** Returns the profile of the currently authenticated user based on the session cookie.
- **Responses:**
  - `200 OK`: Returns authenticated User object (Username, User ID, Join Date).
  - `204 No Content`: No active session found (used for silent client-side checks).
  - `401 Unauthorized`: User record not found for the provided session.
  - `500 Internal Server Error`: Database not initialized.

---

## Testing Strategy & Results

### Backend Unit Tests

We maintain a 1:1 test-to-function ratio for all API handlers and models.

| Test File | Coverage |
|---|---|
| `backend/pkg/api/auth_test.go` | `RegisterHandler` logic and validation |
| `backend/pkg/api/auth_handlers_test.go` | Login/Logout/Me flows |
| `backend/pkg/api/handlers_test.go` | System handlers (Health, 404) |

**Results:** 100% Pass. All edge cases (duplicate users, invalid sessions) are covered.

### Frontend Unit & Integration Tests

| Test File | Coverage |
|---|---|
| `frontend/src/app/core/services/auth.service.spec.ts` | Session management and API calls |
| `frontend/src/app/core/interceptors/auth.interceptor.spec.ts` | CSRF and 401 handling |
| `frontend/src/app/features/auth/register/register.component.spec.ts` | Form validation and submission |
| `frontend/src/app/pages/dashboard/dashboard.component.spec.ts` | Protected route access |

**Results:** 100% Pass.

### Cypress E2E Testing

| Test Spec | Coverage |
|---|---|
| `frontend/cypress/e2e/auth/register.cy.ts` | Registration button functionality and form submission |

**Results:** 100% Pass. Run via `npm run e2e:headless`.

---

## Metrics & Velocity

| Metric | Value |
|---|---|
| **Planned Issues** | 6 |
| **Completed Issues** | 6 |
| **Completion Rate** | 100% |
| **New Endpoints** | 3 (`/api/login`, `/api/logout`, `/api/me`) |
| **Total Lines of Code** | ~1200 (TS + Go) |

---

## Conclusion

Sprint 2 was a resounding success. We delivered exactly what was promised: a secure, functional authentication system. The foundation is now ready for Sprint 3, where we will focus on:

- Refining the `/api/me` endpoint and frontend dashboard
- Migrating to JWT-based authentication
- Building out the Room data model for Poker lobby management and table creation