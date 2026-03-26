# Poker Frontend

## Frontend Baseline Decisions

### Node policy
- Required: **Node 20 LTS**
- `.nvmrc` in repo root is the source of truth (`20`)
- Enforced via:
  - `package.json` engines: `"node": "20.x"`
  - `.npmrc`: `engine-strict=true`

Run:

```bash
nvm use 20
```

before any `npm` command.

### Angular version policy
- Angular packages are pinned to `21.1.4`
- No floating (`^`) Angular versions are allowed

### UI library decision
- Selected library: **Angular Material**
- Installed packages:
  - `@angular/material@21.1.4`
  - `@angular/cdk@21.1.4`
  - `@angular/animations@21.1.4`
- Configured:
  - Theme import in `src/styles.scss`
  - Animations provider in `src/app/app.config.ts`
- Verification component:
  - `src/app/pages/login/login.html` uses `mat-card` and a Material button

### Routing architecture
- Public routes:
  - `/login`
  - `/register`
- Protected-shell routes (guard intentionally deferred to auth story):
  - `/dashboard`
  - `/lobby`
  - `/table/:id`
- Redirects:
  - `/` -> `/login`
  - wildcard `**` -> `/login`
- Route config: `src/app/app.routes.ts`

### Scope boundary
- Placeholder-only baseline for setup and architecture
- No authentication logic
- No feature logic

## Authentication Service

The authentication system manages user login, registration, session state, and provides observables for components.

### Architecture

- **Location**: `src/app/core/services/auth.service.ts`
- **State Management**: RxJS BehaviorSubject pattern
- **Observable API**: `currentUser$` for components to subscribe to
- **Session Handling**: HttpOnly cookies (via withCredentials: true)

### Key Methods

```typescript
// Check if user has valid session (calls GET /api/me)
authService.checkSession(): Observable<User | null>

// Login user (POST /api/login, then fetches user profile)
authService.login(credentials): Observable<User | null>

// Register new user (POST /api/register, returns created user)
authService.register(data): Observable<User>

// Logout user (POST /api/logout, clears local state)
authService.logout(): Observable<void>

// Synchronous accessors for route guards
authService.isAuthenticated(): boolean
authService.getCurrentUser(): User | null
```

### Usage in Components

```typescript
// Subscribe to user state
currentUser$ = this.authService.currentUser$;

// Login
login(username: string, password: string) {
  this.authService.login({ username, password }).subscribe({
    next: (user) => this.router.navigate(['/dashboard']),
    error: () => this.showError('Login failed')
  });
}

// Template
<div *ngIf="(currentUser$ | async) as user">
  Welcome {{ user.username }}!
</div>
```

### Session Flow

1. **App Initialization** (via APP_INITIALIZER)
   - `checkSession()` is called before app fully loads
   - Validates existing session cookie via GET /api/me
   - Restores user state if valid

2. **Login**
   - POST credentials to /api/login
   - Backend validates and sets HttpOnly session cookie
   - Automatically fetches user profile via checkSession()
   - currentUser$ observable updated with user data

3. **Register**
   - POST registration data to /api/register
  - Backend creates user and returns created user payload
  - Does not assume an authenticated session is created

4. **Logout**
   - POST to /api/logout (backend clears cookie)
   - Clears currentUser$ (sets to null)
   - Components notified via observable

### Security Features

- ✅ **HttpOnly Cookies**: Token stored in secure cookie, not accessible from JavaScript
- ✅ **CSRF Protection**: X-CSRF-Token automatically injected by AuthInterceptor
- ✅ **Session Validation**: checkSession() validates token on app load
- ✅ **Auto-Logout on 401**: AuthInterceptor redirects to login on unauthorized

### Testing

See `src/app/core/services/auth.service.spec.ts` for comprehensive unit tests covering:
- State management updates
- Login flow (no token extraction from body)
- Register flow
- Logout flow
- Error handling and edge cases
- Loading state tracking

Run tests with:
```bash
npm test -- --watch=false --browsers=ChromeHeadless
```

All 29 tests pass ✅

### Models

User-related interfaces defined in `src/app/core/models/user.model.ts`:
- `User` - Authenticated user data (id, username, createdAt, updatedAt)
  - Mapped from backend fields (ID, CreatedAt, UpdatedAt)
- `LoginCredentials` - Login form data (username, password)
- `RegisterData` - Registration form data (username, password, confirmPassword)

### Further Documentation

Authentication architecture, RxJS patterns, and design decisions are documented in this README and related in-code comments.

### Backend/Frontend Contract Test Matrix

Use this checklist whenever auth service behavior or backend auth endpoints change.

| Contract Area | Source of Truth | Expected Behavior | Frontend Verification |
| --- | --- | --- | --- |
| User JSON shape from backend | `backend/pkg/models/user.go`, `docs/api/openapi.yaml` | Backend returns `ID`, `CreatedAt`, `UpdatedAt`, `username` | `AuthService` maps to frontend `User` (`id`, `createdAt`, `updatedAt`) in `src/app/core/services/auth.service.ts` |
| Session restore (`GET /api/me`) | Backend `/api/me` handler + OpenAPI | `401`: unauthenticated guest; non-`401`: do not force logout on transient failures | `checkSession()` tests in `src/app/core/services/auth.service.spec.ts` |
| Login flow | Backend `/api/login` + `/api/me` | Login sets session, then `/me` resolves current user | `login()` tests in `src/app/core/services/auth.service.spec.ts` |
| Register flow | `backend/pkg/api/auth.go` `RegisterHandler` | Register creates user payload; does not imply authenticated session | `register()` tests in `src/app/core/services/auth.service.spec.ts` (no implicit `/me`) |
| Logout flow | Backend `/api/logout` | Logout clears backend session; frontend clears local user state even if backend call fails | `logout()` tests in `src/app/core/services/auth.service.spec.ts` |

Quick contract regression run:

```bash
npm test -- --watch=false --browsers=ChromeHeadless --include='**/auth.service.spec.ts'
```

## Setup and verification

```bash
nvm use 20
npm install
npm run build
npm start
```

## Running Tests

The frontend uses **Jasmine** as the testing framework with **Karma** as the test runner.

### Run all tests
```bash
npm test
```

### Run tests in watch mode (re-runs on file changes)
```bash
npm test -- --watch
```

### Run tests with code coverage report
```bash
npm test -- --code-coverage
```

### Run tests for a specific file
```bash
npm test -- --include='**/auth.interceptor.spec.ts'
```

### Run tests without watch mode (CI mode)
```bash
npm test -- --watch=false
```

### Run unit tests independently
```bash
npm run test:unit
```

### Run tests in headless Chrome (for CI/CD pipelines)
```bash
npm test -- --watch=false --browsers=ChromeHeadless
```

### Run frontend-backend integration smoke test (register flow)
This test verifies frontend API contract behavior against a live backend using CSRF + cookies.

1. Start backend server (from repo root):
```bash
cd backend
GO_ENV=development go run ./cmd/server
```
2. In another terminal run:
```bash
cd frontend
npm run test:integration:backend
```

#### Testing Framework Details
- **Jasmine**: BDD (Behavior-Driven Development) testing framework for unit tests
- **Karma**: Test runner that launches browsers and runs tests
- **Configuration**: `karma.conf.js` defines the testing setup, browser, reporters, and plugins

> **Note:** These tests are automatically run on every pull request via the GitHub Actions workflow (`.github/workflows/pull_request_test.yaml`). The workflow ensures all frontend unit tests pass before merging.

