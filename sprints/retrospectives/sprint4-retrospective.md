# Sprint 4: Full Gameplay, Profile, Leaderboard, and Frontend Completion

**Sprint Goal:** Complete the full poker gameplay loop, add player stats and leaderboard, finish all frontend UI with sound and toast feedback, and reach full unit and E2E test coverage across every frontend feature.

---

## Sprint 4 Summary

Sprint 4 delivered the complete frontend gameplay experience and the backend WebSocket and room infrastructure for RoyalFlush. The major outcomes were:

- Backend: Room Management HTTP APIs, Gorilla WebSocket hub/client runtime, shared typed message protocol, live room state ownership model, and lobby broadcasting.
- Frontend: Full poker table UI with real-time game state, game controls, game state visualization, player profile pages, global leaderboard, sound effects, and toast notifications.
- Documentation: Full refresh of root, backend, frontend, deployment, changelog, and OpenAPI docs.
- Testing: 26 Jasmine/Karma unit spec files and 5 Cypress E2E suites passing.

---

## Backend Work — Sprint 4

### Backend Architecture Overview (Whole Project)

The backend is a Go application using Gorilla Mux for routing and GORM with SQLite for persistence. The package layout is:

| Package | Responsibility |
|---|---|
| `cmd/server/` | Entry point — wires routes, middleware, and migrations |
| `pkg/api/` | HTTP handlers: auth, admin, CSRF, health, rooms, WebSocket upgrade |
| `pkg/auth/` | JWT creation/validation, bcrypt password hashing, session management |
| `pkg/db/` | Singleton GORM client backed by SQLite (`data/` directory) |
| `pkg/models/` | Persistent data models: `User`, `Room`; `User.ToResponse()` for safe serialization |
| `pkg/migrations/` | Auto-migrations run at startup |
| `pkg/middleware/` | Auth middleware (JWT/session enforcement) |
| `pkg/room/` | Room service, room code generator, lobby presence and broadcast logic |
| `pkg/socket/` | Gorilla WebSocket hub and client read/write pumps |
| `pkg/protocol/` | Typed WebSocket message envelope definitions shared across backend consumers |

Auth uses session cookies (not Bearer tokens). CSRF tokens are fetched from `GET /api/csrf` and sent via `X-CSRF-Token` on all mutating requests.

---

### 1. WebSocket Message Protocol Contract (Issue #56, PR #82)

Added the shared typed message envelope (`pkg/protocol/`) used by both the WebSocket hub and frontend consumers.

**Completed:**
- `protocol.go` — defines `MessageType` constants and the `Envelope` struct with typed payload unmarshalling.
- Shared across `pkg/socket/`, `pkg/api/ws_handler.go`, and `pkg/room/lobby.go`.
- Provides a single source of truth so backend and frontend never diverge on message shape.

---

### 2. Room Management APIs (Issue #16, PR #83)

Implemented the full REST surface for room lifecycle.

**Completed:**
- `pkg/api/room_handlers.go` — handlers for `POST /api/rooms` (create), `GET /api/rooms/:code` (lookup), and `POST /api/rooms/:code/join`.
- Router wiring in `pkg/api/router.go` with auth middleware on all room routes.
- `pkg/room/service.go` — creates rooms with generated codes, persists to SQLite, validates join attempts.
- `pkg/room/code.go` — random alphanumeric room code generator.
- `pkg/api/room_handlers_test.go` — Ginkgo/Gomega integration tests for create, lookup, join, and conflict cases using temporary SQLite state.

---

### 3. WebSocket Infrastructure (Issue #17, PR #84)

Added the Gorilla WebSocket dependency and the hub/client runtime.

**Completed:**
- `pkg/socket/hub.go` — central hub managing connected clients per room, fan-out broadcasting, and clean client unregistration.
- `pkg/socket/client.go` — per-connection read and write pumps with configurable ping/pong heartbeat.
- `pkg/api/ws_handler.go` — authenticated `/ws?room=<code>` upgrade handler; rejects unauthenticated connections.
- `pkg/api/ws_handler_test.go` — heartbeat and upgrade behavior tests.
- `pkg/api/ws_runtime.go` — wires hub lifecycle to server startup.
- Gorilla WebSocket added to `go.mod`.

---

### 4. Lobby Logic and Broadcasting (Issue #18 — Adjacent, completed 2026-04-14)

Implemented room-scoped presence tracking and player event broadcasting.

**Completed:**
- `pkg/room/lobby.go` — handles `JOIN_ROOM` and `LEAVE_ROOM` protocol messages, maintains live player list per room, emits `PLAYER_JOINED` and `PLAYER_LEFT` broadcasts to all room members.
- Room isolation: each room's lobby state is independent.
- `pkg/room/lobby_test.go` — covers join, leave, duplicate prevention, and broadcast correctness.

---

### 5. Live Room State Ownership Design (Issue #57, PR #85)

Clarified the separation between persistent room metadata (SQLite) and transient live state (in-memory hub).

**Completed:**
- Persistent state: room code, name, host user ID, password hash, status (`waiting` / `in_game`) stored in `models.Room`.
- Transient state: live player list, ready flags, chat messages — owned by the hub and reset on server restart.
- Documented reconnect behavior: clients that reconnect receive a `ROOM_SNAPSHOT` with current player list.
- Expanded room API test coverage with reset and reconnect scenarios.

---

### 6. Frontend Room Route and Join Flow Alignment (Issue #59, PR #87)

Aligned the backend room API response shape and navigation contract with the frontend route definitions.

**Completed:**
- Create-room response returns `{ code }` matching the Angular router's `lobby/:code` path parameter.
- Join-room returns `403` with a password prompt signal when a password-protected room is joined without credentials.
- Backend integration tests updated to assert against the exact JSON shape the frontend consumes.

---

## Frontend Work — Sprint 4

### 1. Poker Table Page and Game State

Implemented the live poker table at `/table/:id` backed by WebSocket game events.

**Completed:**
- `pages/table/table.ts` — full table arena with 9 seat slots, community card display (5 slots), hole cards (2 per player), phase badge, and pot display.
- `core/services/game-state.service.ts` — processes incoming WebSocket game state messages, tracks phase transitions (waiting → pre-flop → flop → turn → river → showdown), manages seat assignments, and detects winners.
- Winner announcement modal with sound effect on showdown.
- Active player indication during betting rounds.
- Chip count and current bet display per seat.

### 2. Game Controls Component

Implemented `features/table/game-controls/game-controls.component.ts`.

**Completed:**
- Action buttons: fold, call, check, raise, bet.
- Raise/bet input validated against minimum raise and player chip count.
- Controls adapt to game phase — check shown when no active bet, call/raise shown when bet is outstanding.
- Emits typed action events to the table component, which sends them via WebSocket.

### 3. Sound Effects Service

Implemented `core/services/sound-effects.service.ts`.

**Completed:**
- Real audio asset files added for win fanfare and game events.
- Mute toggle with persistent state exposed as `isMuted$` observable.
- `isMuted$` cached to avoid repeated subscriptions.
- Called from the table component on winner announcement.

### 4. Toast Notification Service

Implemented `core/services/toast.service.ts`.

**Completed:**
- Global toast notification display via Angular CDK overlay or positioned container.
- Auto-dismissal after configurable timeout.
- Used throughout login, register, room creation, and join flows for user feedback.

### 5. Profile Page

Implemented `pages/profile/profile.ts` and `core/services/stats.service.ts`.

**Completed:**
- `/profile/:id` page fetching and displaying individual player statistics.
- Stats: games played, wins, total chip earnings.
- Protected by `authGuard`.
- Linked from the leaderboard table (player name is a navigable link).

### 6. Leaderboard Page

Implemented `pages/leaderboard/leaderboard.ts` and `core/services/leaderboard.service.ts`.

**Completed:**
- `/leaderboard` page with global player rankings.
- Sortable by total earnings, win rate, and number of wins.
- Public route — visible without authentication.
- Player names link to their individual profile pages.

### 7. Login Component

Finalized `features/auth/login/login.component.ts`.

**Completed:**
- Reactive form with username and password validation.
- Calls `AuthService.login()` and handles success/error states.
- On success, redirects to `/dashboard`.
- Error shown inline.

### 8. App Routes and Route Guards

Finalized `app.routes.spec.ts` and route configuration.

**Completed:**
- Full route table: home, login, register, dashboard, profile/:id, leaderboard, lobby/:code, table/:id, admin.
- `authGuard` enforced on all protected routes.
- Catch-all redirect to `/home`.

### 9. Responsive Design (Issue #95, PR #102)

Added full responsive layout support across all pages.

**Completed:**
- Global SCSS breakpoint variables and reset.
- Responsive table, card, and game controls SCSS.
- Responsive profile, leaderboard, and hamburger nav.
- Cypress responsive suite covering mobile (375px), tablet (768px), and desktop (1280px) breakpoints.
- Note: PR #102 was merged but GitHub issue #95 was still open at report time — should be closed.

### 10. PR Review Fixes and Cleanup (PR #103)

**Completed:**
- Addressed PR review comments across multiple components.
- Removed stale TODO.md file.
- Removed `console.error` calls from production code.
- Cached `isMuted$` in sound effects service to avoid repeated subscriptions.
- Fixed subscription leaks with `takeUntilDestroyed`.
- Touched routes, auth, WebSocket service/specs, room/login/dashboard/table/profile/leaderboard files, and sound assets.

---

## Adjacent Tickets Completed Just Before Sprint 4

These closed immediately before the sprint window but directly feed sprint 4 work.

| Ticket | Area | Completed | Summary |
|---|---|---|---|
| #18 Lobby Logic & Broadcasting | Backend | 2026-04-14 | Room-scoped lobby presence, JOIN_ROOM/LEAVE_ROOM handling, room snapshots, PLAYER_JOINED/PLAYER_LEFT broadcasts, room isolation tests. |
| #19 WebSocket Service | Frontend | 2026-04-12 | Angular `WebSocketService`, typed message envelope model, connection state stream, send/receive helpers, lifecycle handling, and unit tests. |
| #20 Create & Join Room UI | Frontend | 2026-04-12 | Create-room and join-room components, room-code validation, protected room route, dashboard integration, unit tests, and Cypress dashboard coverage. |
| #21 Lobby Component & State | Frontend | 2026-04-12 | Lobby room-code view, live player list, host-first ordering, duplicate-player prevention with Lodash, WebSocket event handling, unit tests, and initial Cypress coverage. |
| #23 Ready Button & Chat UI | Frontend | 2026-04-14 | Lobby ready toggle, chat UI, WebSocket lobby service behavior, send-on-enter support, empty-message guard, and Cypress coverage. |
| #70 Generic Card Component | Frontend | 2026-04-14 | Reusable card rendering component with rank/suit inputs, red/black suit styling, face-down state, scoped SCSS tokens, and card unit tests. |

---

## Documentation (Issue #60, PR #101)

Full refresh of all project documentation to match implemented behavior.

**Completed:**
- Root README updated with project overview and quick-start.
- Backend README updated with package layout and test commands.
- Frontend README updated with Angular architecture and service descriptions.
- `docs/api/openapi.yaml` updated to reflect session-cookie auth, CSRF endpoint, rooms API, admin routes, WebSocket upgrade, and `in_game` room status.
- Deployment docs updated with environment variables and startup sequence.
- CHANGELOG updated through sprint 4.

---

## Frontend Unit Tests

**Command to run all unit tests:**

```bash
cd frontend
npm run test:unit
```

**All 26 unit test spec files:**

| File | What It Covers |
|---|---|
| `src/app/app.spec.ts` | App bootstrap, CSRF initialization |
| `src/app/app.routes.spec.ts` | Route config, auth guard behavior |
| `src/app/core/interceptors/auth.interceptor.spec.ts` | HTTP interception, CSRF injection, 401 redirect |
| `src/app/core/services/admin.service.spec.ts` | Admin user listing and deletion operations |
| `src/app/core/services/auth.service.spec.ts` | Login, register, checkSession, logout, error paths |
| `src/app/core/services/csrf.service.spec.ts` | CSRF token fetching and caching |
| `src/app/core/services/game-state.service.spec.ts` | Game state updates, phase transitions, winner detection, player actions |
| `src/app/core/services/leaderboard.service.spec.ts` | Leaderboard fetching and sorting |
| `src/app/core/services/room.service.spec.ts` | Room create/join service calls, seat normalization |
| `src/app/core/services/sound-effects.service.spec.ts` | Sound playback, mute toggle, isMuted$ caching |
| `src/app/core/services/stats.service.spec.ts` | User stats retrieval by player ID |
| `src/app/core/services/toast.service.spec.ts` | Toast display, auto-dismissal timing |
| `src/app/core/services/websocket.service.spec.ts` | Connection lifecycle, send/receive, reconnect, cleanup |
| `src/app/features/auth/login/login.component.spec.ts` | Login form validation and submission |
| `src/app/features/auth/register/register.component.spec.ts` | Registration form validation and submission |
| `src/app/features/rooms/create-room/create-room.component.spec.ts` | Room creation form, validation, API calls, error handling |
| `src/app/features/rooms/join-room/join-room.component.spec.ts` | Join room form, code validation, password-protected join |
| `src/app/features/table/game-controls/game-controls.component.spec.ts` | Game action controls, bet validation, adaptive button state |
| `src/app/pages/admin/admin.spec.ts` | Admin page rendering and user management actions |
| `src/app/pages/dashboard/dashboard.component.spec.ts` | Dashboard rendering, logout, session state |
| `src/app/pages/home/home.component.spec.ts` | Home page login/register/join room modal flows |
| `src/app/pages/leaderboard/leaderboard.spec.ts` | Leaderboard loading, column sorting, row rendering |
| `src/app/pages/lobby/lobby.spec.ts` | JOIN_ROOM, PLAYER_JOINED, PLAYER_LEFT, deduplication, ready state |
| `src/app/pages/profile/profile.spec.ts` | Profile page stats loading and rendering |
| `src/app/pages/table/table.spec.ts` | Table arena rendering, card slot display, phase badge, pot display |
| `src/app/shared/card/card.component.spec.ts` | Card rendering for all suits and values, face-down state |

**Expected result:**
```
Browser: Chrome Headless
SUMMARY:
  Executed X specs, 0 failures.
  Status: SUCCESS
```

---

## Cypress E2E Tests

**Command to run headless:**

```bash
cd frontend
npm run e2e:headless
```

**Command to run interactively:**

```bash
cd frontend
npm run e2e
```

**All 5 Cypress E2E test suites:**

| File | What It Covers |
|---|---|
| `cypress/e2e/register.cy.ts` | Registration form submission, validation, success/error handling |
| `cypress/e2e/dashboard.cy.ts` | Create Room form and navigation; Join Room form, code validation, password reveal on 403 |
| `cypress/e2e/lobby.cy.ts` | Lobby page render, room code display, player list, ready button, chat UI, WebSocket join/leave events |
| `cypress/e2e/table.cy.ts` | Table arena render, 5 community card slots, 2 hole card slots, phase badge, pot display, player zones |
| `cypress/e2e/responsive.cy.ts` | Responsive layout at mobile (375px), tablet (768px), and desktop (1280px) breakpoints; touch interaction |

---

## Backend Test Validation Evidence

- PR #84: `env GOCACHE=/tmp/poker-go-build-cache go test ./...` passes.
- PR #85: Ginkgo/Gomega room API tests using temporary SQLite state.
- PR #83: Room handler tests cover create, lookup, join, and conflict (duplicate code) cases.
- PR #87: Integration tests assert exact JSON shape consumed by the frontend.

---

## Frontend Architecture Summary

The Sprint 4 frontend is a fully standalone-component Angular 18 application with the following structure:

**Public routes:** `/home`, `/login`, `/register`, `/leaderboard`

**Protected routes (authGuard):** `/dashboard`, `/profile/:id`, `/admin`, `/lobby/:code`, `/table/:id`

**Core services:** `AuthService`, `WebSocketService`, `GameStateService`, `RoomService`, `LeaderboardService`, `StatsService`, `CsrfService`, `ToastNotificationService`, `SoundEffectsService`, `AdminService`

**HTTP layer:** `AuthInterceptor` adds `withCredentials: true` and `X-CSRF-Token` to all requests; handles 401 by clearing state and redirecting to login.

**Real-time layer:** `WebSocketService` manages connection lifecycle with exponential backoff reconnection. `GameStateService` consumes typed WebSocket messages and updates observable game state.

**Testing:** 26 Jasmine/Karma unit spec files, 5 Cypress E2E suites.

---

## Open Items / Follow-Ups

- Close GitHub issue #95 — PR #102 (responsive design) was fully merged but the issue remained open.
- Link PR #103 to an issue or add a changelog entry so the cleanup is traceable.
- Pull latest `develop` branch before generating metrics from local git; local branch does not include all PRs merged on 2026-04-29 and 2026-04-30.
