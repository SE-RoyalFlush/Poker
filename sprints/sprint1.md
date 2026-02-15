# Sprint 1: Foundation & Secure Authentication

**Sprint Goal:** Establish a production-ready application skeleton with secure, HttpOnly cookie-based authentication. By the end of this sprint, a user must be able to securely register, login, access a protected route, and logout.

---

## Issue #1: [Backend] Initialize Go Module & HTTP Router

### Linked Stories
* **Blocked by:** None
* **Blocks:** #3, #4, #5

### Story Description
We need to initialize the Go project structure to support a clean architecture. This involves setting up the module, organizing folders (handlers, models, middleware), and configuring the `gorilla/mux` router to handle basic HTTP requests.

### Test Outline
* **Verify Server Start:** Ensure the application starts on port 8080 without panic.
* **Verify Health Check:** Call `GET /health` and receive a `200 OK` status with JSON body `{"status": "alive"}`.
* **Verify 404:** Call a random route (e.g., `/random`) and ensure it returns a structured JSON 404 error, not the default text.

### Checklist
- [ ] Initialize `go.mod` with project name `github.com/pshimanshu/Poker/backend`.
- [ ] Create folder structure: `cmd/server`, `pkg/api`, `pkg/models`, `pkg/middleware`.
- [ ] Install `github.com/gorilla/mux`.
- [ ] Implement `main.go` to initialize the server.
- [ ] Create a simple `/health` handler in `pkg/api/health.go`.
- [ ] Configure `http.Server` with reasonable timeouts (Read/Write) to prevent potential leaks.

---

## Issue #2: [Backend] Database Connection (SQLite + GORM)

### Linked Stories
* **Blocked by:** #2
* **Blocks:** #5, #6

### Story Description
Establish a robust connection to a SQLite database using GORM. This module must follow the Singleton pattern or Dependency Injection to provide the database instance to other services. It must also handle the creation of the `.db` file if it does not exist.

### Test Outline
* **Verify File Creation:** Start the server and check that `poker.db` is created in the root (or data) directory.
* **Verify Connection:** Check console logs for "Database connection established".
* **Verify Concurrency:** (Optional basic check) Ensure the DB connection pool settings are applied (e.g., max open connections).

### Checklist
- [ ] Install `gorm.io/gorm` and `gorm.io/driver/sqlite`.
- [ ] Create `pkg/db/client.go` to manage the GORM instance.
- [ ] Configure GORM logger to print SQL queries in `Debug` mode (useful for dev).
- [ ] Create a `Connect()` function that returns the `*gorm.DB` instance.
- [ ] Ensure foreign keys are enabled in SQLite (`PRAGMA foreign_keys = ON`).

---

## Issue #3: [Backend] Security Middleware (CORS & CSRF)

### Linked Stories
* **Blocked by:** #2
* **Blocks:** #5, #6

### Story Description
Security is paramount. We need middleware to handle Cross-Origin Resource Sharing (CORS) to allow our Angular frontend (localhost:4200) to communicate with credentials (cookies). Additionally, we must implement CSRF protection to validate state-changing requests, preparing us for cookie-based auth.

### Test Outline
* **Verify CORS Preflight:** Send an `OPTIONS` request to any endpoint and verify `Access-Control-Allow-Origin` matches the frontend URL and `Access-Control-Allow-Credentials` is `true`.
* **Verify CSRF Token Generation:** Ensure `GET` requests return a CSRF token (in header `X-CSRF-Token` or cookie).
* **Verify CSRF Blocking:** Send a `POST` request without the token and verify it returns `403 Forbidden`.

### Checklist
- [ ] Install `github.com/rs/cors` and `github.com/gorilla/csrf` (or similar trusted middleware).
- [ ] Configure CORS to strictly allow `http://localhost:4200` (do not use wildcard `*`).
- [ ] Enable `AllowCredentials: true` in CORS config.
- [ ] Configure CSRF middleware to look for the token in the `X-CSRF-Token` header.
- [ ] Create a `GET /api/csrf` endpoint that effectively hands the token to the frontend (if using double-submit cookie pattern).

---

## Issue #4: [Backend] User Model & Registration API

### Linked Stories
* **Blocked by:** #3, #4
* **Blocks:** #6, #11

### Story Description
Define the `User` data model and implement the registration flow. This involves accepting JSON input, securely hashing the password using `bcrypt`, and persisting the user to SQLite. We must strictly handle duplicate usernames.

### Test Outline
* **Verify Hashing:** Inspect the SQLite database manually to ensure the password column contains a bcrypt hash, not plain text.
* **Verify Success:** `POST /register` with valid data returns `201 Created`.
* **Verify Duplicate:** `POST /register` with an existing username returns `409 Conflict`.
* **Verify Bad Input:** `POST /register` with empty fields returns `400 Bad Request`.

### Checklist
- [ ] Create `User` struct in `pkg/models` (Fields: ID, Username, PasswordHash, CreatedAt).
- [ ] Add GORM auto-migration in the server startup for `User`.
- [ ] Install `golang.org/x/crypto/bcrypt`.
- [ ] Implement `RegisterHandler` in `pkg/api/auth.go`.
- [ ] Add validation: Username min length 3, Password min length 6.
- [ ] Register the route `POST /api/register`.

---

## Issue #5: [Backend] Login, Logout & Session Check APIs

### Linked Stories
* **Blocked by:** #3, #4, #5
* **Blocks:** #12, #13

### Story Description
Implement the core authentication logic using JWTs stored in **HttpOnly Cookies**. This includes Login (issue cookie), Logout (clear cookie), and Me (validate cookie). This is the critical security barrier for the app.

### Test Outline
* **Verify Cookie Security:** Login and inspect browser/Postman headers to see `Set-Cookie` with `HttpOnly`, `Secure` (if HTTPS), and `SameSite=Strict`.
* **Verify Session Validation:** Call `GET /api/me` without a cookie -> `401 Unauthorized`. Call with cookie -> `200 OK` + User Profile.
* **Verify Logout:** Call `POST /api/logout`, then call `GET /api/me` immediately after. The second call must fail.

### Checklist
- [ ] Define JWT claims struct (UserID, Username, Exp).
- [ ] Implement `LoginHandler`: Validate creds -> Generate JWT -> Set Cookie.
- [ ] Implement `LogoutHandler`: Set cookie expiration to the past.
- [ ] Implement `MeHandler`: Parse cookie -> Validate JWT -> Return User info.
- [ ] Implement `AuthMiddleware`: Protects routes by checking the cookie before allowing access.
- [ ] Ensure NO tokens are ever returned in the JSON response body.

---

## Issue #6: [Frontend] Angular Init & Global Styles

### Linked Stories
* **Blocked by:** None
* **Blocks:** #8, #9

### Story Description
Initialize the Angular workspace. Clean up the boilerplate code and set up the global styling framework (Angular Material) and project folder structure. This sets the stage for all UI development.

### Test Outline
* **Verify Build:** Run `ng serve` and ensure the app loads without console errors.
* **Verify Material:** Add a simple Material Button to the main page and ensure the ripple effect and typography work.
* **Verify Routing:** Ensure `<router-outlet>` is present and working.

### Checklist
- [ ] Create new Angular project `royalflush-web`.
- [ ] Add `@angular/material` and configure a theme (Indigo/Pink).
- [ ] Install `lodash` and `@types/lodash`.
- [ ] Create core folders: `src/app/core`, `src/app/features`, `src/app/shared`.
- [ ] Configure `proxy.conf.json` (optional) if we need to proxy requests to backend during dev.

---

## Issue #7: [Frontend] Core HTTP & Interceptor Setup

### Linked Stories
* **Blocked by:** #7
* **Blocks:** #9

### Story Description
We need to configure the Angular `HttpClient` to communicate securely. A specific HttpInterceptor is required to ensure `withCredentials: true` is attached to *every* request so cookies are sent. It must also handle the `X-CSRF-Token` extraction and injection if the backend requires it.

### Test Outline
* **Verify Credentials:** Inspect Network tab for any XHR request. `Access-Control-Allow-Credentials` headers should be visible.
* **Verify CSRF Injection:** (If implemented) Check that `POST` requests automatically have the `X-CSRF-Token` header attached.

### Checklist
- [ ] Create `AuthInterceptor` in `src/app/core/interceptors`.
- [ ] Configure interceptor to clone requests and set `withCredentials: true`.
- [ ] Register interceptor in `app.module.ts`.
- [ ] Implement error handling in interceptor (e.g., auto-redirect to login on global 401).

---

## Issue #8: [Frontend] Authentication Service

### Linked Stories
* **Blocked by:** #8
* **Blocks:** #10, #11

### Story Description
Implement the client-side logic for authentication. This service will communicate with the backend APIs (`/login`, `/register`, `/me`). It acts as the source of truth for the current user's state (Logged In vs. Guest) using RxJS BehaviorSubjects.

### Test Outline
* **Verify State Management:** calling `checkSession()` should update the `currentUser$` observable.
* **Verify Login Call:** `login()` method should trigger the API and, on success, fetch the user profile. It should NOT look for a token in the body.
* **Verify Logout:** `logout()` should hit the backend API and then clear the local BehaviorSubject.

### Checklist
- [ ] Create `AuthService` in `src/app/core/services`.
- [ ] Define `User` interface matching backend model.
- [ ] Implement `login(credentials)`, `register(data)`, `logout()`.
- [ ] Implement `checkSession()` which calls `GET /api/me`.
- [ ] Expose `currentUser$` observable for components to subscribe to.

---

## Issue #9: [Frontend] Registration Component

### Linked Stories
* **Blocked by:** #9
* **Blocks:** None

### Story Description
Build the UI for user registration using Reactive Forms. It needs to look professional (using Material Design) and handle validation errors (e.g., "Username already taken") gracefully.

### Test Outline
* **Verify Validation:** Submit button should be disabled until form is valid.
* **Verify Error Display:** If backend returns 409, a generic error "Username taken" should appear.
* **Verify Success:** On success, redirect to `/login` (or auto-login if we choose that path).

### Checklist
- [ ] Create `RegisterComponent` in `src/app/features/auth`.
- [ ] Build Reactive Form with validators (Required, MinLength).
- [ ] Use `MatFormField`, `MatInput`, and `MatButton`.
- [ ] Integrate `AuthService.register()`.
- [ ] Add loading spinner (`MatProgressSpinner`) during submission.

---

## Issue #10: [Frontend] Login Component

### Linked Stories
* **Blocked by:** #9
* **Blocks:** #13

### Story Description
Build the Login UI. This is the entry point for existing users. It must communicate with the Auth Service and handle the redirect logic upon successful authentication.

### Test Outline
* **Verify Submission:** Pressing Enter in the password field should submit the form.
* **Verify Auth Failure:** Invalid credentials should show a "Invalid username or password" snackbar or alert.
* **Verify Navigation:** Success should immediately route to `/home`.

### Checklist
- [ ] Create `LoginComponent` in `src/app/features/auth`.
- [ ] Build Reactive Form (Username, Password).
- [ ] Integrate `AuthService.login()`.
- [ ] Handle 401 errors specifically to show "Wrong password" vs "Server error".

---

## Issue #11: [Backend] Dashboard "Me" API

### Linked Stories
* **Blocked by:** #6
* **Blocks:** #13

### Story Description
This is a specific refinement of the session check. We need to ensure the `/api/me` endpoint is fully protected and returns the exact data structure the frontend expects for the dashboard (e.g., Join Date, Username, ID).

### Test Outline
* **Verify Data Integrity:** Ensure the JSON response matches the TypeScript interface exactly.
* **Verify Speed:** This endpoint will be called often; ensure it has no unnecessary DB latency.

### Checklist
- [ ] Refine `User` JSON response (hide PasswordHash, etc. using `json:"-"`).
- [ ] Ensure `CreatedAt` is formatted ISO 8601.

---

## Issue #12: [Frontend] Protected Home Dashboard & Guard

### Linked Stories
* **Blocked by:** #11, #12
* **Blocks:** None

### Story Description
Create the landing page for logged-in users and secure it. This involves an **AuthGuard** that checks the session before allowing the route to activate. If the session check fails, it must redirect to Login.

### Test Outline
* **Verify Guard Protection:** Open Incognito window, paste `http://localhost:4200/home`. Should redirect to `/login`.
* **Verify Data Display:** Login, go to Home. Should see "Welcome, [Username]".
* **Verify Logout Flow:** Click Logout button on Home. Should redirect to Login. Back button should NOT work (or at least guard should catch it).

### Checklist
- [ ] Create `AuthGuard` in `src/app/core/guards`.
- [ ] Create `HomeComponent` in `src/app/features/home`.
- [ ] Implement `CanActivate` to call `AuthService.checkSession()`.
- [ ] Display User data in `HomeComponent` template.
- [ ] Add a prominent "Logout" button.