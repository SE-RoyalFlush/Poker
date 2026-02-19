# Sprint 1 Retrospective: Backend & Frontend Foundation

**Sprint Duration:** Sprint 1  
**Team Focus:** Backend Development + Frontend Scaffolding  
**Date:** February 2026

---

## Executive Summary

Sprint 1 established the foundational architecture for both backend and frontend development on the Poker platform. We did not complete the full secure authentication user journey, but we delivered a strong backend core and a usable Angular frontend scaffold that is ready for integration work in Sprint 2.

**Key Achievements:**
- Established Go project structure with clean architecture
- Implemented database layer with GORM and SQLite
- Configured production-grade security middleware (CORS & CSRF)
- Delivered first functional API endpoint (User Registration)
- Initialized Angular workspace with Angular Material and routing
- Added frontend route skeleton with public and protected-shell paths
- Created initial page scaffolds and backend API proxy configuration

**Completion Rate:** 
Backend: 4/6 planned issues completed + 1 new issue
Frontend: 1/6 planned issues completed + 2 new issues
Overall: 5/12 planned issues completed, with 8 total work items delivered including unplanned scope


---

## What We Achieved

### Phase 1: Foundation & Learning (Issue #1-14)

Before writing any code, we invested time in understanding Go best practices and industry standards. This proved to be a critical decision that shaped our entire project structure.

**Activities:**
- Studied the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) to understand conventional Go project organization
- Reviewed [Uber's Go Style Guide](https://github.com/uber-go/guide) for idiomatic coding patterns
- Analyzed production-grade Go repositories (simplebank, go-api-boilerplate) to learn from real-world implementations
- Researched error handling patterns and testing best practices
- Studied GitHub issue writing conventions and repository management

**Impact:**
This foundational learning phase, while not initially planned in detail, was essential. It prevented us from building on incorrect patterns and ensured our codebase would be maintainable and scalable. The time invested here saved us from potential refactoring later.

**Deliverable:** Knowledge base for team alignment on Go conventions and project standards

---

### Phase 2: Project Initialization (Issue #1-1)

With our understanding of Go best practices established, we initialized the project structure.

**Implementation Details:**
- Initialized Go module: `github.com/SE-RoyalFlush/Poker/backend`
- Created clean architecture folder structure:
  - `cmd/server/` - Application entry point
  - `pkg/api/` - HTTP handlers and routing logic
  - `pkg/models/` - Data models and business entities
  - `pkg/middleware/` - Cross-cutting concerns
  - `pkg/db/` - Database connection management
- Integrated `gorilla/mux` router for robust HTTP request handling
- Implemented health check endpoint (`GET /health`) returning JSON `{"status": "alive", "database": "connected"}`
- Created custom 404 handler returning structured JSON errors (not plain text)
- Configured `http.Server` with production-ready timeouts:
  - `ReadTimeout: 15s` - Prevents slow-read attacks
  - `WriteTimeout: 15s` - Prevents slow-write issues
  - `IdleTimeout: 60s` - Manages keep-alive connections efficiently

**Key Files Created:**
- [cmd/server/main.go](cmd/server/main.go) - Server initialization and graceful shutdown
- [pkg/api/health_handler.go](pkg/api/health_handler.go) - Health check endpoint
- [pkg/api/not_found_handler.go](pkg/api/not_found_handler.go) - Structured 404 responses

**Testing:**
- ✅ Server starts on port 8080 without panic
- ✅ `GET /health` returns 200 OK with JSON body
- ✅ Random routes return structured JSON 404 errors

**Architectural Decisions:**
We chose `gorilla/mux` over the standard library router for its superior route matching, middleware support, and ease of testing. The clean architecture separation ensures our handlers remain testable and our business logic stays decoupled from HTTP concerns.

---

### Phase 3: Database Layer (Issue #1-2)

With the HTTP layer functional, we established our data persistence foundation.

**Implementation Details:**
- Integrated GORM ORM with SQLite driver (`gorm.io/gorm`, `gorm.io/driver/sqlite`)
- Created `pkg/db/client.go` implementing singleton pattern for database management
- Configured GORM logger in debug mode for development visibility
- Enabled SQLite foreign keys enforcement (`PRAGMA foreign_keys = ON`)
- Implemented connection pooling with reasonable defaults
- Added graceful database shutdown hooks
- Database file: `poker.db` created in project backend directory

**Key Technical Decisions:**
- **SQLite Choice:** Ideal for our MVP phase - zero configuration, serverless, perfect for development and testing
- **Singleton Pattern:** Ensures single database connection instance shared across the application, preventing connection leaks
- **GORM Benefits:** Type-safe queries, automatic migrations, relationship handling, and excellent Go integration

**Key Files Created:**
- [pkg/db/client.go](pkg/db/client.go) - Database connection management with Connect(), GetDB(), Ping(), and Close() functions

**Testing:**
- ✅ `poker.db` file automatically created on first startup
- ✅ Console logs confirm "Database connection established"
- ✅ Connection pool configured with proper limits
- ✅ Foreign key constraints enforced

**Database Architecture:**
Our database layer provides a clean abstraction that will allow us to swap to PostgreSQL or MySQL in production without touching business logic - we simply change the driver. GORM's auto-migration feature ensures schema changes deploy automatically.

---

### Phase 4: Security Middleware (Issue #1-3)

Security was non-negotiable. We implemented two critical middleware layers before building any user-facing features.

**Implementation Details:**

#### CORS Middleware (`github.com/rs/cors`)
- Strict origin allowlist: `http://localhost:4200` (no wildcards)
- `AllowCredentials: true` - Enables cookie-based authentication
- Allowed methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
- Allowed headers: Accept, Authorization, Content-Type, X-CSRF-Token
- Exposed headers: X-CSRF-Token (for frontend token extraction)

#### CSRF Middleware (`github.com/gorilla/csrf`)
- 32-byte authentication key loaded from `CSRF_AUTH_KEY` environment variable
- Token validation via `X-CSRF-Token` header
- Cookie configuration:
  - `HttpOnly: true` - Prevents JavaScript access
  - `Secure: true` (production) / `false` (development)
  - `SameSite: Lax` - Balanced security and usability
- Created `GET /api/csrf` endpoint for token distribution

**Key Files Created:**
- [pkg/api/csrf_handler.go](pkg/api/csrf_handler.go) - CSRF token handler
- [cmd/server/main.go](cmd/server/main.go) - Middleware chain configuration

**Testing:**
- ✅ OPTIONS requests return proper CORS headers
- ✅ `Access-Control-Allow-Origin: http://localhost:4200`
- ✅ `Access-Control-Allow-Credentials: true`
- ✅ POST requests without CSRF token return 403 Forbidden
- ✅ `GET /api/csrf` returns valid CSRF token

**Security Posture:**
This middleware configuration prevents common web attacks:
- **CSRF Protection:** Validates state-changing requests originate from our frontend
- **CORS Policy:** Prevents unauthorized domains from making authenticated requests
- **Cookie Security:** HttpOnly + Secure + SameSite prevents XSS and CSRF attacks

**Production Considerations:**
The environment-aware CSRF configuration (`APP_ENV`, `GO_ENV`) allows us to develop locally without HTTPS while maintaining strict security in production. This pattern is essential for team collaboration.

---

### Phase 5: User Registration API (Issue #1-4)

With infrastructure in place, we delivered our first business feature: secure user registration.

**Implementation Details:**

#### Data Model (`pkg/models/user.go`)
```go
type User struct {
    gorm.Model                    // Includes: ID, CreatedAt, UpdatedAt, DeletedAt
    Username     string           // Unique index, not null
    PasswordHash string           // Bcrypt hash, excluded from JSON (json:"-")
}
```

#### Registration Endpoint (`POST /api/register`)
- **Input Validation:**
  - Username: Minimum 3 characters, trimmed whitespace
  - Password: Minimum 6 characters (enforced before hashing)
- **Duplicate Detection:**
  - Explicit database check before insertion
  - Returns `409 Conflict` for existing usernames
  - Handles concurrent duplicate insertions gracefully
- **Password Security:**
  - Bcrypt hashing with default cost factor (10)
  - Original password never stored or logged
  - Hash excluded from JSON responses via `json:"-"` tag
- **Response:**
  - Success: `201 Created` with user object (hash excluded)
  - Duplicate: `409 Conflict` with error message
  - Bad Input: `400 Bad Request` with validation error

**Key Files Created:**
- [pkg/models/user.go](pkg/models/user.go) - User data model
- [pkg/api/auth.go](pkg/api/auth.go) - RegisterHandler with full validation

**Testing:**
- ✅ Valid registration returns 201 Created
- ✅ Duplicate username returns 409 Conflict
- ✅ Empty/short username returns 400 Bad Request
- ✅ Short password returns 400 Bad Request
- ✅ Database inspection confirms bcrypt hash (not plaintext)
- ✅ JSON response excludes PasswordHash field

**Security Highlights:**
- **Bcrypt:** Industry-standard adaptive hashing that becomes more expensive as hardware improves
- **Validation:** Server-side enforcement prevents malicious or malformed input
- **JSON Exclusion:** The `json:"-"` tag on PasswordHash ensures it never leaks in API responses
- **Race Condition Handling:** Duplicate constraint violations handled even under concurrent load

**Code Quality:**
- Comprehensive error handling with structured JSON responses
- Proper logging without exposing sensitive data
- Clean separation: validation → business logic → persistence
- Testable handlers with dependency injection

---

### Phase 6: Frontend Initialization & Global Styles (Issue #1-6)

In parallel with backend hardening, we established the Angular frontend baseline and development tooling.

**Implementation Details:**
- Initialized Angular workspace in `frontend/` with standalone component architecture
- Installed and configured Angular Material (`@angular/material`, `@angular/cdk`, `@angular/animations`)
- Added Material prebuilt theme (`indigo-pink`) in `frontend/angular.json`
- Configured global styling and typography baseline in `frontend/src/styles.scss`
- Enforced Node version policy with `engines.node: 20.x` in `frontend/package.json`
- Added `lodash` and `@types/lodash` as utility dependencies

**Key Files Created:**
- [frontend/angular.json](frontend/angular.json) - Angular build/serve configuration and global theme setup
- [frontend/package.json](frontend/package.json) - Frontend dependencies and scripts
- [frontend/src/styles.scss](frontend/src/styles.scss) - Global styles and Material surface defaults
- [frontend/src/app/app.config.ts](frontend/src/app/app.config.ts) - Router + animations providers

**Testing:**
- ✅ Frontend unit test suite runs successfully (`npm test -- --watch=false`)
- ✅ Angular Material components render in UI scaffold (`login` page uses `mat-card` + `mat-button`)
- ✅ Router outlet bootstraps correctly via unit tests (`frontend/src/app/app.spec.ts`)

**Architectural Decisions:**
We chose standalone Angular components and provider-based app configuration to reduce module overhead and keep feature delivery fast. Angular Material was selected to ensure accessible, consistent UI primitives from the start.

---

### Phase 7: Frontend Routing & Layout Scaffolding (New Scope)

To avoid blocking future feature work, we added route architecture and page-level navigation scaffolding ahead of auth service implementation.

**Implementation Details:**
- Configured route groups for:
  - Public pages: `/login`, `/register`
  - Protected-shell pages: `/dashboard`, `/lobby`, `/table/:id`
- Added default redirect from `/` to `/login` and wildcard fallback to `/login`
- Implemented `ProtectedShell` layout with nested `<router-outlet />` for protected pages
- Added root application toolbar shell for shared app chrome

**Key Files Created:**
- [frontend/src/app/app.routes.ts](frontend/src/app/app.routes.ts) - Route map and lazy component loading
- [frontend/src/app/layouts/protected-shell/protected-shell.ts](frontend/src/app/layouts/protected-shell/protected-shell.ts) - Protected layout shell
- [frontend/src/app/layouts/protected-shell/protected-shell.html](frontend/src/app/layouts/protected-shell/protected-shell.html) - Nested outlet template
- [frontend/src/app/app.html](frontend/src/app/app.html) - App-level toolbar and router outlet

**Testing:**
- ✅ Route configuration compiles with lazy-loaded standalone components
- ✅ App shell unit tests verify router outlet rendering
- 🟡 AuthGuard integration intentionally deferred to authentication stories

**Architecture Rationale:**
This route-first scaffold allows frontend and backend work to proceed in parallel. It provides a stable navigation contract while auth/session logic is still being implemented.

---

### Phase 8: Frontend Page Scaffolds & API Proxy Setup (New Scope)

We delivered placeholder pages and local API proxying to enable rapid integration in Sprint 2.

**Implementation Details:**
- Created standalone page components for `login`, `register`, `dashboard`, `lobby`, and `table`
- Added placeholder templates/styles to define route targets and UI structure
- Configured Angular dev server proxy for `/api` → `http://localhost:8080`
- Created core project directories (`core`, `features`, `shared`) for future feature organization

**Key Files Created:**
- [frontend/src/app/pages/login/login.ts](frontend/src/app/pages/login/login.ts) - Login page scaffold
- [frontend/src/app/pages/register/register.ts](frontend/src/app/pages/register/register.ts) - Registration page scaffold
- [frontend/src/app/pages/dashboard/dashboard.ts](frontend/src/app/pages/dashboard/dashboard.ts) - Dashboard scaffold
- [frontend/src/app/pages/lobby/lobby.ts](frontend/src/app/pages/lobby/lobby.ts) - Lobby scaffold
- [frontend/src/app/pages/table/table.ts](frontend/src/app/pages/table/table.ts) - Table scaffold
- [frontend/proxy.conf.json](frontend/proxy.conf.json) - Backend API proxy configuration

**Testing:**
- ✅ Placeholder route components compile and load through Angular router
- ✅ Proxy configuration is wired into Angular `serve` development target
- 🟡 No auth flow integration yet (depends on backend login/me/logout APIs)

**Delivery Impact:**
Sprint 2 can start directly with real feature implementation (forms, services, guards) without revisiting setup work. Integration with backend endpoints can begin immediately via `/api` proxy.

---

## What We Didn't Achieve

### Planned But Not Completed (Backend)

#### Issue #1-5: Login, Logout & Session Check APIs
**Scope:** JWT-based authentication with HttpOnly cookies, session validation, logout functionality

**Why Not Completed:**
- Time underestimation: The learning phase and foundation work took longer than anticipated
- Security complexity: Implementing JWT correctly with HttpOnly cookies, refresh tokens, and proper expiration requires careful design
- Dependency chain: Required Issues #1-1 through #1-4 to be fully tested first

**Impact:**
Without authentication endpoints, users cannot log in. This blocks the entire user experience flow.

#### Issue #1-11: Dashboard "Me" API
**Scope:** Protected endpoint returning current user profile

**Why Not Completed:**
- Direct dependency on Issue #1-5 (authentication must exist first)
- Blocked: Cannot implement auth-protected endpoints without auth middleware

**Impact:**
Frontend cannot display user-specific data or verify session state.

---

### Frontend Work (Partially Completed)

Sprint 1 frontend delivery was foundational rather than feature-complete.

**Completed in Sprint 1:**
- ✅ Issue #1-6: Angular Init & Global Styles
- ✅ New Scope: Routing and protected-shell scaffolding
- ✅ New Scope: Placeholder page components and dev proxy setup

**Planned But Not Completed (Frontend):**

#### Issue #1-7: Core HTTP & Interceptor Setup
**Scope:** `HttpClient` setup with `withCredentials` and CSRF header injection

**Why Not Completed:**
- Prioritized frontend scaffolding before API integration logic
- Backend auth/session endpoints were still pending

**Impact:**
Frontend cannot yet send authenticated cookie-based requests with a centralized interceptor.

#### Issue #1-8: Authentication Service
**Scope:** Central auth state service (`login`, `register`, `logout`, `checkSession`)

**Why Not Completed:**
- Blocked by incomplete backend login/me/logout APIs
- Scope deferred to Sprint 2 vertical-slice implementation

**Impact:**
No shared authentication state or reactive session model in frontend.

#### Issue #1-9: Registration Component
**Scope:** Production-ready reactive registration form integrated with backend

**Why Not Completed:**
- Current registration page is a placeholder scaffold
- Form validation, API wiring, and error UX were not implemented

**Impact:**
Users cannot register through the frontend yet, despite backend endpoint availability.

#### Issue #1-10: Login Component
**Scope:** Login form with API integration, UX feedback, and route navigation

**Why Not Completed:**
- Depends on backend login endpoint and auth service completion
- Placeholder screen exists but lacks business logic

**Impact:**
No frontend login user journey is currently available.

#### Issue #1-12: Protected Home Dashboard & Guard
**Scope:** AuthGuard-protected dashboard with user-aware session checks

**Why Not Completed:**
- Guard implementation depends on Issue #1-8 and backend `/api/me`
- Dashboard page exists only as a placeholder

**Impact:**
Protected routes are scaffolded but not yet enforced by authentication.

---

## Analysis: Why We Fell Short

### 1. **Optimistic Planning Without Learning Buffer**

**Issue:** Our original sprint plan assumed the team had Go expertise and could immediately begin feature development.

**Reality:** We needed significant upfront investment in:
- Go language fundamentals and idioms
- GORM ORM patterns and best practices
- Middleware patterns in Go HTTP servers
- Testing strategies in Go (table-driven tests, suite patterns)

**Time Invested:** Issue #1-14 (learning phase) took approximately 20-25% of sprint capacity but wasn't initially accounted for as a formal work item.

### 2. **Underestimated Infrastructure Complexity**

**Issue:** We treated "set up database" and "add CSRF" as small tasks.

**Reality:** Production-ready infrastructure requires:
- **Database Layer:** Connection pooling, migration strategy, error handling, testing patterns
- **Security Middleware:** Environment-aware configuration, proper CORS policies, CSRF token distribution strategy, cookie security flags
- **Each component:** Required research, implementation, testing, and team review

**Time Impact:** Infrastructure issues (#1-1, #1-2, #1-3) consumed ~50% of sprint capacity vs. the estimated ~30%.

### 3. **No Frontend-Backend Parallelization**

**Issue:** We attempted to staff backend-first without parallel frontend work.

**Reality:** This creates a waterfall pattern that's inefficient:
- Frontend blocked until APIs exist
- Cannot validate API design without frontend consumption
- No end-to-end testing possible

**Better Approach:** Frontend could have mocked backend responses and developed in parallel.

### 4. **Testing Debt**

**Issue:** While we have test files created (`*_test.go`, `*_suite_test.go`), comprehensive test coverage was not completed.

**Reality:** Writing production-quality tests takes time:
- Table-driven tests for handlers
- Mock database for unit tests
- Integration tests for middleware chains
- CSRF token flow testing

**Impact:** We have test infrastructure but limited coverage. This creates technical debt.

### 5. **Scope Ambiguity in "MVP"**

**Issue:** Sprint 1 goal stated "production-ready application skeleton with secure authentication."

**Reality:** "Production-ready" and "authentication" are both large scopes:
- Production-ready implies: logging, monitoring, error tracking, deployment configs
- Authentication implies: register, login, logout, session management, token refresh

**Learning:** We should have defined "walking skeleton" vs "production-ready" more clearly.

---

## What Went Well

### ✅ Code Quality & Architecture

- **Clean Architecture:** Our folder structure follows Go conventions and will scale well
- **Security-First:** CORS and CSRF implemented from day one (not bolted on later)
- **No Technical Debt in Code:** What we built is production-quality, not throw-away prototype code

### ✅ Team Learning Investment

- Issue #1-14 ensured the entire team aligned on Go best practices
- We didn't rush into building "the wrong thing quickly"
- Codebase reads cleanly; new developers can onboard easily

### ✅ Dependency Management

- We respected the blocking relationships between issues
- Didn't skip security to "move faster"
- Each layer was tested before building on top

### ✅ Documentation

- Comprehensive issue descriptions with test outlines and checklists
- Self-documenting code with clear function names and comments
- OpenAPI spec foundation in `docs/api/openapi.yaml`

---

## Lessons Learned

### 1. **Buffer for Learning is Not Optional**

**Lesson:** When working with a new technology stack, allocate 20-30% of sprint capacity to learning and environment setup.

**Action for Next Sprint:**
- Frontend will require Angular learning time
- Account for this explicitly in Sprint 2 planning

### 2. **Infrastructure is a Feature**

**Lesson:** Database, security, and middleware are not "quick setup tasks" - they are foundational features that deserve proper estimation.

**Action for Next Sprint:**
- Treat infrastructure work with same respect as user-facing features
- Add buffer for edge cases and testing

### 3. **Test Coverage Matters From Day One**

**Lesson:** We have test files but incomplete coverage. This will slow us down later.

**Action for Next Sprint:**
- Include "write tests" as explicit checklist items in every issue
- Define coverage targets (e.g., 80% for business logic)
- Block PR merges without tests

### 4. **Vertical Slices > Horizontal Layers**

**Lesson:** Building "all backend" then "all frontend" creates long feedback loops and blocks integration testing.

**Action for Next Sprint:**
- Deliver vertical slices (e.g., "user can register end-to-end") instead of horizontal layers
- Enable frontend and backend to work in parallel using API contracts

### 5. **Definition of Done Must Include Integration**

**Lesson:** We completed backend stories but cannot demonstrate end-user value without frontend.

**Action for Next Sprint:**
- Define "Done" as "feature works end-to-end" or explicitly call out if it's a backend-only deliverable
- Demo real user workflows in sprint review

---

## Sprint 2 Planning Recommendations

### Priority 1: Complete Authentication Flow (Backend)

**Must Complete:**
- Issue #1-5: Login, Logout & Session Check APIs
- Issue #1-11: Dashboard "Me" API (protected endpoint)

**Why:** Without authentication, users cannot interact with the application. This is the highest-priority blocker.

**Estimation:**
- **Issue #1-5:** 3-4 days (JWT implementation, cookie management, middleware, testing)
- **Issue #1-11:** 1 day (simple endpoint but requires auth middleware)

### Priority 2: Frontend Foundation & Integration

**Must Complete:**
- Issue #1-6: Angular Init & Global Styles
- Issue #1-7: Core HTTP & Interceptor Setup
- Issue #1-8: Authentication Service
- Issue #1-9: Registration Component (integrate with existing backend API)

**Why:** Enables end-to-end user registration flow. Demonstrates tangible user value.

**Estimation:**
- **Issue #1-6:** 1 day (Angular setup, Material Design)
- **Issue #1-7:** 1 day (HttpClient, interceptors for CSRF)
- **Issue #1-8:** 2 days (AuthService, state management)
- **Issue #1-9:** 2 days (component, form validation, integration)

### Priority 3: Complete Auth User Journey

**Must Complete:**
- Issue #1-10: Login Component
- Issue #1-12: Protected Home Dashboard & Guard

**Why:** Delivers complete "register → login → dashboard → logout" user journey.

**Estimation:**
- **Issue #1-10:** 2 days (login form, error handling, navigation)
- **Issue #1-12:** 2 days (AuthGuard, dashboard, logout flow)

### Capacity Adjustment

**Sprint 1 Reality Check:**
- Completed Planned Scope: 5/12 issues (~42%)
- Total Delivered: 8 issues (5 backend, 3 frontend) including unplanned scaffolding work
- Delivery Pattern: Planned auth features slipped; foundational platform work expanded

**Sprint 2 Adjusted Plan:**
- Target: 8-9 issues (finish remaining auth backend + frontend foundation)
- Focus: Deliver one complete vertical slice (user registration + login)
- Buffer: 20% for testing, integration issues, and inevitable surprises

### Risk Mitigation

**Risk:** Frontend team lacks Angular experience  
**Mitigation:** 
- Schedule team learning session (similar to Go learning in Sprint 1)
- Assign Issue #1-6 and #1-7 to most experienced frontend developer
- Pair programming for Issue #1-8 and #1-9

**Risk:** CSRF token handling between Angular and Go backend  
**Mitigation:**
- Backend CSRF already implemented
- Frontend interceptor should fetch token from `GET /api/csrf` and attach `X-CSRF-Token` on state-changing requests
- Allocate extra testing time for this integration point

**Risk:** Testing debt from Sprint 1  
**Mitigation:**
- Do NOT carry testing debt into Sprint 2
- First task of Sprint 2: Complete test coverage for Sprint 1 deliverables
- Block new feature work until tests pass

---

## Metrics & Velocity

### Sprint 1 Statistics

| Metric | Value                                                       |
|--------|-------------------------------------------------------------|
| **Planned Issues** | 12 (6 backend, 6 frontend) + 3 New (1 backend + 2 frontend) |
| **Completed Issues** | 8 (5 backend, 3 frontend)                                   |
| **Planned Completion Rate** | 42% (5/12 planned issues)                               |
| **API Endpoints** | 3 (`/api/health`, `/api/register`, `/api/csrf`)             |
| **Database Models** | 1 (`User`)                                                  |
| **Frontend Routes Scaffolded** | 5 (`/login`, `/register`, `/dashboard`, `/lobby`, `/table/:id`) |

### Velocity Calculation for Sprint 2

**Sprint 1 Completed:** 8 issues across ~14 days  
**Adjusted Velocity:** ~0.57 issues/day or ~4 issues/week  
**Sprint 2 Capacity:** 14 days × 0.57 = ~8 issues at current pace  
**With Learning:** Account for Angular learning curve → realistic target: 7-8 issues (including smaller frontend tasks)

---

## Technical Debt Register

### Immediate (Must Address in Sprint 2)

1. **Test Coverage Gaps**
   - **Debt:** Test files exist but coverage is not complete
   - **Impact:** Cannot refactor safely; bugs may slip through
   - **Remediation:** Write tests for all handlers and database operations

2. **Environment Configuration**
   - **Debt:** `.env` file not in repository (correct), but no `.env.example` template
   - **Impact:** New developers don't know what variables to set
   - **Remediation:** Create `.env.example` with dummy values

3. **Error Logging**
   - **Debt:** Using `log.Printf` directly instead of structured logging
   - **Impact:** Hard to parse logs in production, no log levels
   - **Remediation:** Integrate structured logger (e.g., `zap` or `logrus`)

### Medium Priority (Sprint 3-4)

4. **Database Migrations**
   - **Debt:** Using GORM AutoMigrate (development convenience)
   - **Impact:** No migration history, hard to rollback changes
   - **Remediation:** Switch to explicit migrations (e.g., `golang-migrate`)

5. **OpenAPI Spec Incomplete**
   - **Debt:** `docs/api/openapi.yaml` exists but not up-to-date with actual endpoints
   - **Impact:** Frontend lacks clear API contract
   - **Remediation:** Auto-generate OpenAPI spec from code or manually update

### Low Priority (Future)

6. **SQLite for Production**
   - **Debt:** SQLite is suitable for development but not production-scale
   - **Impact:** No horizontal scaling, limited concurrent writers
   - **Remediation:** Migration path to PostgreSQL (Sprint 5+)

---

## Conclusion

Sprint 1 was a **foundation-building sprint** that established production-grade backend infrastructure and a frontend application scaffold for the Poker platform. While planned authentication features remain incomplete, the delivered architecture is stable and integration-ready.

**Key Accomplishments:**
- ✅ Team aligned on Go best practices (Issue #1-14)
- ✅ Clean architecture project structure (Issue #1-1)
- ✅ Robust database layer with GORM (Issue #1-2)
- ✅ Production-ready security middleware (Issue #1-3)
- ✅ Functional user registration API (Issue #1-4)
- ✅ Angular workspace + Material + routing baseline (Issue #1-6)
- ✅ Frontend route shell and page scaffolds ready for feature integration

**Key Learning:**
We learned that **"going slow to go fast"** is real. The time invested in learning Go conventions and building proper infrastructure will pay dividends in Sprint 2 and beyond. We did not accumulate technical debt by rushing.

**Sprint 2 Focus:**
Complete the authentication story (login, logout, session management) and connect the existing frontend scaffolds to backend APIs to deliver the first end-to-end user journey.

**Team Morale:**
Despite not hitting our original targets, the team should be proud. We built something **right**, not just **fast**. The codebase is clean, secure, and ready to scale. Sprint 2 will move faster because the foundation is solid.

---

## Appendix: Completed Work by File

### Backend Code Delivered

```
backend/
├── cmd/
│   └── server/
│       └── main.go                    ✅ Server initialization, middleware chain, graceful shutdown
├── pkg/
│   ├── api/
│   │   ├── auth.go                    ✅ RegisterHandler with validation and bcrypt
│   │   ├── csrf_handler.go            ✅ CSRF token distribution
│   │   ├── health_handler.go          ✅ Health check endpoint
│   │   └── not_found_handler.go       ✅ Structured 404 responses
│   ├── db/
│   │   └── client.go                  ✅ Database connection management
│   └── models/
│       └── user.go                    ✅ User model with GORM tags
├── go.mod                              ✅ Dependency management
└── data/
    └── poker.db                        ✅ SQLite database (generated)
```

### Test Infrastructure (Partial)

```
backend/
├── pkg/
│   ├── api/
│   │   ├── api_suite_test.go          🟡 Suite setup (incomplete coverage)
│   │   ├── auth_test.go               🟡 Registration tests (partial)
│   │   └── handlers_test.go           🟡 Handler tests (partial)
│   ├── db/
│   │   ├── db_suite_test.go           🟡 DB suite setup
│   │   └── client_test.go             🟡 Connection tests (partial)
│   └── models/
│       └── models_test.go             🟡 Model tests (partial)
└── pkg/
    └── suite_test.go                   🟡 Root test suite
```

### Frontend Code Delivered

```
frontend/
├── angular.json                         ✅ Angular build/serve config + Material theme wiring
├── package.json                         ✅ Angular/Material dependencies and scripts
├── proxy.conf.json                      ✅ `/api` proxy to backend (`localhost:8080`)
├── src/
│   ├── styles.scss                      ✅ Global styles + Material surface defaults
│   └── app/
│       ├── app.config.ts                ✅ Router + animations providers
│       ├── app.routes.ts                ✅ Public/protected route scaffolding
│       ├── app.html                     ✅ Root toolbar + router outlet shell
│       ├── layouts/
│       │   └── protected-shell/         ✅ Nested layout for protected pages
│       └── pages/
│           ├── login/                   ✅ Material-based login page scaffold
│           ├── register/                ✅ Registration page scaffold
│           ├── dashboard/               ✅ Dashboard placeholder scaffold
│           ├── lobby/                   ✅ Lobby placeholder scaffold
│           └── table/                   ✅ Table placeholder scaffold (`/table/:id`)
```

### Frontend Test Infrastructure

```
frontend/
└── src/
    └── app/
        └── app.spec.ts                 🟡 Base app tests (app creation + router outlet)
```

**Legend:**
- ✅ Complete and production-ready
- 🟡 Infrastructure exists but incomplete coverage
- ❌ Not started

---

**Document Version:** 1.1  
**Last Updated:** February 19, 2026  
**Next Review:** Sprint 2 Planning Meeting
