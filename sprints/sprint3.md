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
