Link

Frontend Link - https://poker-frontend-e5dr6xr0w-himanshu-pss-projects.vercel.app/home
Frontend Video - https://youtu.be/VVIyLQ9D1NY 
Backend Video - https://www.youtube.com/watch?v=hy8Zer2STrM

# Sprint 3: Protected Rooms, Live Lobby State, and Shared Auth

**Sprint Goal:** Strengthen the multiplayer foundation by standardizing authentication, protecting room and WebSocket routes, persisting room metadata, and implementing room-scoped real-time lobby behavior with test coverage across backend and frontend.

---

## Sprint 3 Summary

During Sprint 3, the team focused on the unfinished room and real-time lobby work from Sprint 2, plus the architectural cleanup needed to support those features safely. The biggest outcomes were:

- Standardized backend authentication around signed session cookies.
- Added shared auth resolution for both HTTP routes and WebSocket handshake flows.
- Introduced a protected router structure for authenticated endpoints.
- Added a persistent `Room` domain model and migration coverage.
- Standardized room codes to a shared 6-character uppercase alphanumeric format.
- Implemented room-scoped WebSocket join, leave, and ready-state broadcasting.
- Added frontend room-entry and lobby behavior that aligns with the backend protocol.
- Added unit tests for the new backend and frontend functionality.

---

## Work Completed In Sprint 3

### 1. Shared Authentication and Protected Routing

We refactored session-based authentication into reusable helpers so protected HTTP handlers and the WebSocket handshake use the same authentication source of truth.

**Completed**
- Added signed-session auth helpers in `backend/pkg/auth/session.go`.
- Added authenticated user resolution via `AuthenticatedUserFromRequest`.
- Applied shared auth enforcement through `backend/pkg/middleware/auth.go`.
- Built a protected route tree in `backend/pkg/api/router.go`.
- Mounted `/api/me`, `/api/rooms`, `/api/rooms/join`, `/api/rooms/{code}`, and `/ws` behind auth middleware.
- Kept public routes such as `/api/health`, `/api/register`, `/api/login`, `/api/logout`, and `/api/csrf` outside protected routing.

**Result**
- HTTP handlers and WebSocket handshake behavior now follow one consistent auth model: signed session cookies.

### 2. Room Domain Model and Persistence

We added a real persisted room model so invite codes, host ownership, room status, and room configuration survive across requests.

**Completed**
- Added `Room` model in `backend/pkg/models/room.go`.
- Added room fields for code, host user, status, max players, privacy, and timestamps.
- Registered `Room` in model/migration flow.
- Added room creation and lookup helpers in `backend/pkg/room/service.go`.
- Added migration verification in `backend/pkg/migrations/migrations_test.go`.

**Result**
- Room metadata is now database-backed instead of being only implied by live socket state.

### 3. Room Code Standardization

We aligned backend generation, backend validation, frontend validation, tests, and examples around one official room code format.

**Completed**
- Standardized room codes to `^[A-Z0-9]{6}$`.
- Added backend generation and normalization in `backend/pkg/room/code.go`.
- Added collision retry handling in room service internals.
- Updated frontend join validation to reject legacy formats like `RF-7742`.

**Result**
- The backend and frontend now agree on the room code contract, reducing integration mismatch.

### 4. WebSocket Room Mapping and Lobby Events

We implemented room-scoped WebSocket behavior so users only receive lobby events for the room they joined.

**Completed**
- Added in-memory room-to-client mapping in `backend/pkg/api/ws_runtime.go`.
- Implemented `JOIN_ROOM` message handling.
- Broadcast `PLAYER_JOINED` events to connected clients in the same room.
- Broadcast `PLAYER_LEFT` when a player disconnects.
- Sent existing player state to newly joined clients.

**Result**
- Lobby traffic is isolated by room instead of being globally broadcast.

### 5. Ready / Not Ready State Management

We implemented transient ready-state tracking for connected players in a room lobby.

**Completed**
- Added in-memory lobby state management in `backend/pkg/room/lobby.go`.
- Implemented `TOGGLE_READY` handling.
- Broadcast `PLAYER_UPDATE` when a player's ready status changes.
- Reset ready state when a player disconnects and rejoins.
- Added `AllReady` support at the lobby layer for future game-start logic.

**Result**
- Players can now change ready state in memory and the room receives synchronized updates immediately.

### 6. Frontend Room and Lobby Integration

We aligned the frontend room flows with the backend APIs and WebSocket protocol.

**Completed**
- Added/updated `CreateRoomComponent` for room creation flow.
- Added/updated `JoinRoomComponent` for room lookup and room code validation.
- Added `Lobby` page behavior that sends `JOIN_ROOM` and reacts to live events.
- Added typed `WsMessage` and `Player` models.
- Used `WebSocketService` to manage connection lifecycle and typed message send/receive.
- Added host-first sorting and duplicate-player protection in the lobby UI.

**Result**
- The frontend now follows the same room code format and live lobby event protocol as the backend.

---

## Frontend Unit Tests

The following frontend unit test files cover Sprint 3 functionality and supporting Sprint 2 functionality that remains part of the active test suite:

- `frontend/src/app/pages/lobby/lobby.spec.ts`
  Covers `JOIN_ROOM`, `PLAYER_JOINED`, `PLAYER_LEFT`, player deduplication, host sorting, room code rendering, player count rendering, and redirect behavior when no room code is present.
- `frontend/src/app/features/rooms/join-room/join-room.component.spec.ts`
  Covers shared 6-character room-code validation, uppercase normalization, join submission, success navigation, password-protected join handling, and quick join.
- `frontend/src/app/features/rooms/create-room/create-room.component.spec.ts`
  Covers room creation form validation, API submission, success navigation, private room password requirement, and error handling.
- `frontend/src/app/core/services/websocket.service.spec.ts`
  Covers WebSocket connection lifecycle, message send/receive, malformed message handling, duplicate-connect protection, disconnect safety, reconnect behavior, and backend socket URL expectations.
- `frontend/src/app/core/services/room.service.spec.ts`
  Covers room create/join service calls and live-room normalization behavior.

Additional frontend unit tests still included in the full suite:

- `frontend/src/app/pages/dashboard/dashboard.component.spec.ts`
- `frontend/src/app/core/services/auth.service.spec.ts`
- `frontend/src/app/core/interceptors/auth.interceptor.spec.ts`
- `frontend/src/app/core/services/csrf.service.spec.ts`
- `frontend/src/app/features/auth/register/register.component.spec.ts`
- `frontend/src/app/pages/home/home.component.spec.ts`
- `frontend/src/app/pages/admin/admin.spec.ts`
- `frontend/src/app/app.spec.ts`

### Frontend Unit Test Report

**Command Run**

```bash
cd frontend && npm run test:unit
```

**Result**

- Browser: Chrome Headless 147.0.0.0
- Total specs executed: 181
- Passed: 181
- Failed: 0
- Status: SUCCESS

**Frontend Unit Test Files**

- `src/app/app.spec.ts`
- `src/app/core/interceptors/auth.interceptor.spec.ts`
- `src/app/core/services/admin.service.spec.ts`
- `src/app/core/services/auth.service.spec.ts`
- `src/app/core/services/csrf.service.spec.ts`
- `src/app/core/services/room.service.spec.ts`
- `src/app/core/services/websocket.service.spec.ts`
- `src/app/features/auth/register/register.component.spec.ts`
- `src/app/features/rooms/create-room/create-room.component.spec.ts`
- `src/app/features/rooms/join-room/join-room.component.spec.ts`
- `src/app/pages/admin/admin.spec.ts`
- `src/app/pages/dashboard/dashboard.component.spec.ts`
- `src/app/pages/home/home.component.spec.ts`
- `src/app/pages/lobby/lobby.spec.ts`
- `src/app/shared/card/card.component.spec.ts`

**Coverage Summary by Area**

- App bootstrap and CSRF initialization
- Authentication service, interceptor, and CSRF service
- Room service API calls and seat normalization
- WebSocket service connection and message flow
- Register, create-room, and join-room component behavior
- Dashboard, admin, home, and lobby page behavior
- Card component rendering and state handling

**Notes**

- The suite completed successfully with no failures.
- The run was executed in headless Chrome through Angular Karma.

---

## Backend Unit Tests

The following backend unit and API test files cover Sprint 3 functionality and the supporting authentication foundation used by Sprint 3 features:

- `backend/pkg/auth/session_test.go`
  Verifies valid session resolution, missing cookie behavior, invalid cookie behavior, and the case where a session is valid but the user no longer exists.
- `backend/pkg/middleware/auth_test.go`
  Verifies protected middleware behavior for valid, missing, and invalid session cookies.
- `backend/pkg/api/router_test.go`
  Verifies protected route access, unauthorized access, middleware coverage for room routes and `/ws`, correct method restrictions, and preservation of public routes.
- `backend/pkg/api/ws_handler_test.go`
  Verifies room join flow, `PLAYER_JOINED` broadcasting, `TOGGLE_READY` updates, `PLAYER_LEFT` on disconnect, and ready-state reset on reconnect.
- `backend/pkg/room/lobby_test.go`
  Verifies ready toggling, reconnect reset behavior, unknown-player error handling, and `AllReady` logic.
- `backend/pkg/room/service_test.go`
  Verifies room persistence, host association, default values, generated code validity, lookup by code, invalid code rejection, and invalid max-player rejection.
- `backend/pkg/room/service_internal_test.go`
  Verifies room-code collision retry behavior and failure after exhausting retry attempts.
- `backend/pkg/migrations/migrations_test.go`
  Verifies that migrations create the rooms table.

Additional backend tests still included in the full suite:

- `backend/pkg/api/auth_test.go`
- `backend/pkg/api/auth_handlers_test.go`
- `backend/pkg/api/handlers_test.go`
- `backend/pkg/db/client_test.go`
- `backend/pkg/models/models_test.go`

### Backend Test Run Results

Latest backend package test results:

```text
ok      github.com/SE-RoyalFlush/Poker/backend/pkg                     0.296s  coverage: [no statements]
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/api                 5.713s  coverage: 75.9% of statements
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/auth                1.801s  coverage: 44.2% of statements
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/db                  0.872s  coverage: 67.8% of statements
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/middleware          0.284s  coverage: 69.2% of statements
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/migrations          1.553s  coverage: 66.7% of statements
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/models              1.104s  coverage: 50.0% of statements
ok      github.com/SE-RoyalFlush/Poker/backend/pkg/room                1.337s  coverage: 81.8% of statements
```

---

## Updated Backend API Documentation

Backend API documentation is maintained in:

- `docs/api/openapi.yaml`

This file should be shown during the Sprint 3 presentation as the current backend API reference in the repository.

**Documented today**
- `/api/health`
- `/api/register`
- `/api/me`
- standardized JSON error responses

**Sprint 3 backend route structure now also includes**
- `/api/rooms`
- `/api/rooms/join`
- `/api/rooms/{code}`
- `/ws`

**Note**
- The router and backend implementation now include the protected room endpoints and WebSocket handshake path.
- The OpenAPI file exists and should be presented as the backend API documentation artifact for this sprint.
- Expanding the spec to fully describe all new room and WebSocket-related routes would be a logical next documentation pass.

---

## Demo Checklist For Presentation

During the narrated video, we should demonstrate the following:

1. Login or session-authenticated access to protected room features.
2. Create a room or join a room using a valid 6-character room code.
3. Enter the lobby and show the room code on screen.
4. Show one player joining and another player receiving the live `PLAYER_JOINED` update.
5. Toggle ready state and show the lobby update.
6. Disconnect/rejoin and explain that ready state resets because it is transient in-memory state.
7. Show frontend unit test results.
8. Show backend unit test results.
9. Open `docs/api/openapi.yaml` and point out the backend API documentation artifact.

---

## Commands To Show Test Results

### Backend

```bash
cd backend
go clean -testcache
go test ./pkg/...
```

### Frontend

```bash
cd frontend
npm test -- --watch=false --browsers=ChromeHeadless
```

If a narrower demo is needed, targeted frontend test files that directly support Sprint 3 include:

```bash
cd frontend
npm test -- --watch=false --browsers=ChromeHeadless --include="src/app/pages/lobby/lobby.spec.ts" --include="src/app/features/rooms/join-room/join-room.component.spec.ts" --include="src/app/features/rooms/create-room/create-room.component.spec.ts" --include="src/app/core/services/websocket.service.spec.ts" --include="src/app/core/services/room.service.spec.ts"
```

---

## Final Sprint 3 Outcome

Sprint 3 moved the project from room scaffolding toward a real protected multiplayer foundation. Authentication is now shared and consistent, room metadata is persisted, room codes are standardized, room-scoped WebSocket events are implemented, ready-state behavior is tested, and the frontend lobby flow is aligned with backend behavior. The result is a stronger base for future work on full room APIs, richer lobby controls, and eventually live poker gameplay.

# Sprint 3 Retrospective: Real-Time Lobby, Room Management, and UI Foundations

**Sprint Duration:** Sprint 3  
**Team Focus:** Real-time lobby features, room lifecycle, and frontend gameplay primitives  
**Analysis Window:** Changes after March 25, 2026 (from git history)  
**Date:** April 2026

---

## Executive Summary

Sprint 3 moved the project from "authenticated app" to "real-time multiplayer foundation." Based on git history after March 25, the team delivered major backend and frontend infrastructure for room creation/joining, lobby WebSocket communication, ready/chat interactions, and the first reusable gameplay UI component (generic card rendering).

**Key Outcomes:**
- Authentication flow was completed and hardened (login integration, `/api/me`, session package extraction).
- Backend routing was centralized and prepared for room/lobby endpoints.
- WebSocket runtime and lobby broadcasting logic were implemented and tested.
- Room domain expanded with model, service, validation, and generated room codes.
- Frontend gained Create/Join Room flows, Lobby page, and chat/ready UX.
- A reusable SVG-based Card component was added for upcoming table/gameplay work.

**Delivery Snapshot (from git since 2026-03-25):**
- 14 commits merged
- 9,212 lines added / 2,443 deleted (net +6,769)
- Churn by area: `frontend` (+3,992 net), `backend` (+2,382 net), `sprints` (+318 net)

---

## Evidence-Based Timeline

### Phase 1: Auth Completion and Test Foundations (Mar 25 - Mar 26)

**Relevant commits:**
- `98cd217` Added login component and backend integration (#48)
- `66c71de` Implemented dashboard "Me" API (#46)
- `0580fba` Added Cypress registration test (#50)

**What changed:**
- Login UI and auth interceptor behavior were tightened.
- Dashboard `/api/me` support was improved in backend auth/JWT flow.
- Cypress infrastructure and first E2E auth coverage were introduced.

---

### Phase 2: Session/Router Architecture Refactor (Apr 11)

**Relevant commits:**
- `8833289` Extracted session handling to auth package (#61)
- `fa1fa9d` Introduced API router and placeholder handlers (#63)
- `425033b` Added frontend WebSocket service with tests (#62)

**What changed:**
- Session logic moved into dedicated auth package with tests.
- `api.NewRouter()` centralized route setup and improved testability.
- Frontend real-time plumbing began via `WebSocketService` and typed models.

---

### Phase 3: Room Domain and Create/Join UX (Apr 12)

**Relevant commits:**
- `f394873` Added room model, service, and tests (#65)
- `2013798` Added create/join room UI (#64)
- `0722cc4` Added lobby component baseline (#66)

**What changed:**
- Backend room model/service were introduced with validation and tests.
- Frontend delivered dedicated create-room and join-room feature components.
- Dashboard was refactored to route into room flows.
- Initial lobby page and lobby-oriented E2E/unit tests were added.

---

### Phase 4: Room Codes, Lobby Runtime, and Chat/Ready (Apr 13)

**Relevant commits:**
- `9814f16` Room code generation and validation (#67)
- `9eeb57a` Lobby logic completion (#68 / #22)
- `05811c5` Lobby broadcasting runtime refinements (#75)
- `fd4bd15` Ready button and chat UI (#76)

**What changed:**
- Unique room code generation and normalization rules were added.
- Lobby runtime and broadcast behavior were implemented end-to-end.
- Frontend lobby added chat panel, ready toggles, and richer state updates.
- Backend WebSocket tests significantly expanded during these changes.

---

### Phase 5: Gameplay UI Primitive (Apr 13)

**Relevant commits:**
- `af0941e` Generic card SVG component (#77)
- `631c111` PR review fixes for card component

**What changed:**
- Reusable card rendering component with specs and styling was added.
- Follow-up fix commit improved quality after code review.

---

## What Went Well

- Vertical slice momentum: auth -> rooms -> lobby -> reusable gameplay UI happened in one sprint window.
- Strong testing discipline: large test additions accompanied service/runtime changes.
- Clear feature decomposition: Create Room, Join Room, Lobby, Card were implemented as separated units.
- Refactor timing was effective: session extraction and centralized router reduced duplication before real-time complexity increased.

---

## Lessons Learned

1. Real-time concurrency needs explicit guardrails early. Multiple commits in one day were required to stabilize lobby/broadcast behavior.
2. UI scope can inflate quickly in lobby features. Chat, ready state, and visual state management expanded complexity beyond initial page scaffolding.
3. Large style files become a maintenance signal. Lobby and home styles grew significantly and now trigger build-size warnings.
4. CommonJS dependencies should be monitored. `lodash` in lobby path currently causes Angular optimization bailout warnings.

---

## Technical Debt Register

1. Room API handlers still began as placeholders and need complete endpoint parity with frontend expectations.
2. WebSocket resilience (reconnect/backoff and reconnect state sync) remains limited.
3. Lobby state and chat persistence are still primarily in-memory concerns.
4. Style budget warning on home page indicates need to modularize/reduce SCSS footprint.
5. CommonJS `lodash` usage should be replaced by ESM-native or framework-native alternatives.

---

## Team and Velocity Metrics

### Commits by Contributor (since Mar 25)

- Himanshu PS: 6
- SaiShravanthReddy: 4
- B Sai Puneeth: 2
- Sanikommu Devi: 2

### Change Volume

| Area | Added | Deleted | Net |
|---|---:|---:|---:|
| Backend | 2619 | 237 | +2382 |
| Frontend | 6191 | 2199 | +3992 |
| Sprint Docs | 321 | 3 | +318 |
| Total | 9212 | 2443 | +6769 |

---

## Sprint 4 Recommendations

1. Complete production-ready Room REST handlers and align contracts with frontend service calls.
2. Add reconnect + resync strategy for WebSocket clients (player list, ready status, chat feed).
3. Start integrating the card component into table/game flows, not as isolated UI only.
4. Reduce style and bundle warnings by splitting oversized SCSS and removing CommonJS dependencies.
5. Add focused end-to-end tests for host actions and lobby-to-table transition.

---

## Conclusion

Git history after March 25 shows Sprint 3 as a high-throughput, architecture-heavy sprint that established the multiplayer core of RoyalFlush. The project now has the essential building blocks for live room-based gameplay: room lifecycle, lobby synchronization, and a reusable card UI foundation. The next sprint should focus on production hardening, gameplay progression, and optimization cleanup.
