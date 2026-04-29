# Changelog

All notable changes to the **RoyalFlush** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- [Frontend: Poker Table Layout Component (Issue #71)](https://github.com/SE-RoyalFlush/Poker/issues/71)
    - Replaced stub `TableComponent` with a full Texas Hold'em layout at `/table/:id`
    - Added `GameStateService` in `frontend/src/app/core/services/` exposing a `BehaviorSubject<GameState>` with `gameState$` observable and `patchState()` method for future WebSocket integration
    - Added `GameState`, `PlayerSeat`, `Card`, and `GamePhase` interfaces in `frontend/src/app/core/models/game-state.model.ts`
    - Community cards area (centre of oval felt): 5 card slot positions — empty dashed placeholders pre-flop, revealed on flop (3), turn (4), and river/showdown (5)
    - Current user hole cards rendered face-up in the bottom player zone; 2 empty placeholders shown when no cards have been dealt
    - Up to 9 opponent seats positioned around the oval felt using pre-calculated ellipse coordinates; each seat shows username, chip count, and 2 face-down `CardComponent`s
    - Oval green felt surface styled with CSS `border-radius: 50%` and `aspect-ratio: 2/1`; seat positions derived from parametric ellipse equations, requiring no JavaScript layout calculations
    - Added unit tests in `table.spec.ts` covering community card counts per phase, opponent seat count for varying player counts, hole card face direction, and lifecycle teardown
    - Added Cypress E2E tests in `cypress/e2e/table.cy.ts` covering page structure, community slots, player zone, phase badge, and pot display
- [Frontend: Generic Card Component (Issue #70)](https://github.com/SE-RoyalFlush/Poker/issues/70)
    - Added reusable `CardComponent` at `frontend/src/app/shared/card/` accepting `rank`, `suit`, and `faceDown` inputs
    - Renders rank label and suit symbol (♠ ♥ ♦ ♣) in top-left and bottom-right corners with a large centre symbol
    - Applies red colouring for hearts/diamonds; black for spades/clubs via `rf-card--red` CSS class
    - Face-down state (`faceDown=true`) shows a styled blue back with a diagonal hatch pattern and hides face content
    - SCSS uses component-scoped CSS custom properties (`--rf-card-*`) aligned with the project `--rf-*` token naming convention
    - Added comprehensive unit test coverage for `CardComponent`, covering all 52 rank/suit combinations, face-down state, colour classification, and aria labels
- [Frontend: Ready Button & Chat UI (Issue #23)](https://github.com/SE-RoyalFlush/Poker/issues/23)
    - Implemented `LobbyWebSocketService` in `frontend/src/app/core/services/lobby-websocket.service.ts` for real-time WebSocket communication; exposes `messages$` and `playerReady$` observables and gracefully degrades when backend is unavailable
    - Implemented full `LobbyComponent` with a ready-toggle button (grey → green on activation) and a scrollable chat panel using Angular Material List
    - Chat supports sending via Enter key or Send button; auto-scrolls to the latest message on receipt
    - Added Cypress 14 E2E tests in `frontend/cypress/e2e/lobby.cy.ts` covering ready button toggle, WebSocket event dispatch, chat send via button and Enter key, empty-message guard, and message list rendering
    - Bootstrapped Cypress E2E infrastructure (`cypress.config.ts`, `cypress/support/e2e.ts`) and added `e2e` / `e2e:headless` npm scripts
    - Added `.claude/` and `CLAUDE.md` to `.gitignore`
- [Backend: Lobby Logic & Broadcasting (Issue #18)](https://github.com/SE-RoyalFlush/Poker/issues/18)
    - Implemented room-scoped WebSocket lobby state in `backend/pkg/api/ws_runtime.go` with per-room client membership and broadcasts
    - Added `JOIN_ROOM`, `LEAVE_ROOM`, `ROOM_STATE`, `PLAYER_JOINED`, and `PLAYER_LEFT` message handling for lobby presence updates
    - `JOIN_ROOM` now returns a room snapshot to the joining client and broadcasts `PLAYER_JOINED` only to the other clients in that room
    - Disconnects and explicit `LEAVE_ROOM` events now remove the client from the room and notify remaining members with `PLAYER_LEFT`
    - Added backend tests covering room isolation, join broadcast semantics, explicit leave, and reconnect-ready-state reset
- [Frontend: Lobby Component & State (Issue #21)](https://github.com/SE-RoyalFlush/Poker/issues/21)
    - Implemented `Lobby` at `frontend/src/app/pages/lobby/lobby.ts` — reads `roomCode` from `?code=` query param, connects via `WebSocketService`, sends `JOIN_ROOM` on open, handles `PLAYER_JOINED` (with `_.uniqBy` deduplication by id) and `PLAYER_LEFT` messages, sorts players host-first via `_.orderBy`
    - Added `Player` interface (`{ id, username, isHost }`) at `frontend/src/app/core/models/player.model.ts` and exported from models barrel
    - Added template at `frontend/src/app/pages/lobby/lobby.html` — room code heading, live player count, player list with avatar initials and "Host" badge
    - Added BEM card-layout styles at `frontend/src/app/pages/lobby/lobby.scss` using `--rf-*` design tokens
    - Added 12 unit tests in `frontend/src/app/pages/lobby/lobby.spec.ts` using `MockWebSocket` + `WS_FACTORY` token pattern (same pattern as WebSocketService spec)
    - Added Cypress E2E tests in `frontend/cypress/e2e/lobby.cy.ts` covering initial render; WS message injection deferred — no native Cypress 14 WS intercept support; covered by unit tests instead
    - First lodash usage in frontend: `uniqBy` and `orderBy` for player list management
- [Frontend: Create & Join Room UI (Issue #20)](https://github.com/SE-RoyalFlush/Poker/issues/20)
    - Extracted `CreateRoomComponent` into `frontend/src/app/features/rooms/create-room/` — room name/players/blinds/privacy form; navigates to `/room/:code` on success
    - Extracted `JoinRoomComponent` into `frontend/src/app/features/rooms/join-room/` — 6-char alphanumeric code input with live rooms browser, 403 password reveal, and `quickJoin`; navigates to `/room/:code` on success
    - Updated frontend room-code validation from legacy `RF-XXXX` to `^[A-Z0-9]{6}$` for the new join flow and `/room/:code` routing convention
    - Fixed post-join navigation to use `room.code` instead of `room.id`
    - Added `room/:code` protected route and placeholder `RoomComponent` at `frontend/src/app/pages/room/`
    - Refactored `DashboardComponent` to embed `<app-create-room>` and `<app-join-room>`, removing all inline form logic
    - Added 13 unit tests covering validation, API calls, navigation, 403 handling, and `quickJoin` behaviour
    - Added Cypress E2E tests in `frontend/cypress/e2e/dashboard.cy.ts` for create flow, join flow, validation, and password reveal
- [Frontend: WebSocket Service (Issue #19)](https://github.com/SE-RoyalFlush/Poker/issues/19)
    - Implemented `WebSocketService` in `src/app/core/services/websocket.service.ts` for managing the client-side WebSocket connection lifecycle
    - Added `WsMessage<T>` interface in `src/app/core/models/ws-message.model.ts` as the typed `{ type, payload }` envelope for all WebSocket traffic
    - Exposes `messages$: Observable<WsMessage>` (Subject-backed) for components to subscribe to incoming server events
    - Exposes `connected$: Observable<boolean>` (BehaviorSubject-backed) for real-time connection state
    - `sendMessage(type, payload)` serializes and sends a typed JSON message; warns without throwing when socket is not open
    - `connect(url?)` wires `onopen`/`onmessage`/`onerror`/`onclose` handlers; guards against double-connect when already OPEN
    - `disconnect()` closes the connection cleanly; safe to call with no active connection
    - Introduced `WS_FACTORY` InjectionToken as a testability seam — overridden in tests to return a `MockWebSocket` instead of a real browser WebSocket
    - Session cookies are sent automatically by the browser on the WS handshake (RFC 6455 §10.5); no explicit credential flag is needed unlike XHR
    - Comprehensive unit tests (17 tests, all passing) using `MockWebSocket` helper covering: connection lifecycle, JSON send/receive, malformed message resilience, double-connect guard, disconnect safety, and cookie URL documentation
    - Exported from `src/app/core/services/index.ts` and `src/app/core/models/index.ts` barrels
    - **Note**: Backend `ws://localhost:8080/ws` endpoint not yet implemented; Cypress E2E test deferred until backend endpoint and consuming UI component exist
- [Frontend: Registration UI Story (Issue #1-11)](https://github.com/SE-RoyalFlush/Poker)
    - Implemented registration UI with Angular Material + Reactive Forms in `frontend/src/app/features/auth/register/`
    - Added validation UX (required/min-length, submit-disabled-until-valid, password match) and `409 -> Username taken` error mapping
    - Added focused register component unit tests for validation, API call trigger, conflict handling, and post-registration navigation to `/dashboard`
    - On successful registration, automatically logs in the user and navigates to `/dashboard`
    - Added `RoomService` unit tests and frontend-backend registration integration smoke test script
    - Added npm script `test:integration:backend` and documented how to run integration checks with backend alive
    - Consolidated frontend auth/room usage to core services + core auth guard for a single maintainable structure
- [Frontend: Authentication Service Implementation](https://github.com/SE-RoyalFlush/Poker/pull/42)
    - Implemented `AuthService` in `src/app/core/services/auth.service.ts` as single source of truth for user authentication state
    - Created User model interfaces (`User`, `LoginCredentials`, `RegisterData`) in `src/app/core/models/user.model.ts`
    - Implemented RxJS BehaviorSubject pattern for state management with `currentUser$` observable for components
    - Implemented `login(credentials)` method - POSTs to `/api/login`, chains to `checkSession()` for user profile fetch
    - Implemented `register(data)` method - POSTs to `/api/register` and returns the created user payload
    - Implemented `logout()` method - POSTs to `/api/logout`, clears user state locally
    - Implemented `checkSession()` method - Calls `GET /api/me` to validate session and restore user state on app initialization
    - Configured `APP_INITIALIZER` to restore user session automatically on app startup
    - Added synchronous helper methods `isAuthenticated()` and `getCurrentUser()` for route guards
    - Comprehensive unit tests (29 tests, 100% passing) using `HttpClientTestingModule`
    - Updated frontend README with authentication architecture, usage examples, and API reference
    - Added detailed authentication and backend auth endpoint documentation to existing project docs
    - **Note**: Backend endpoints `/api/login`, `/api/logout`, and `/api/me` are required but not yet implemented
- [Frontend: Secure HTTP Client with AuthInterceptor](https://github.com/SE-RoyalFlush/Poker/pull/41)
    - Implemented AuthInterceptor in `src/app/core/interceptors/auth.interceptor.ts` for secure HTTP communication
    - Configured interceptor to automatically attach `withCredentials: true` to all HTTP requests for cookie-based authentication
    - Implemented X-CSRF-Token extraction from cookies and automatic injection into request headers
    - Added global error handling in interceptor with automatic redirect to login on 401 (Unauthorized) responses
    - Registered HTTP interceptor in application configuration for all outgoing requests
    - Added comprehensive unit tests for interceptor functionality and error handling
- [Backend: User Model & Registration API](https://github.com/SE-RoyalFlush/Poker/issues/6)
    - Defined User data model using GORM with secure password hashing (bcrypt)
    - Implemented registration endpoint `POST /api/register` with validation
    - Added database auto-migration for User model
    - Integrated unit tests for authentication and models
    - Updated OpenAPI specification and Bruno collection for registration
- [Basic Go Setup: Health Check and APIs](https://github.com/SE-RoyalFlush/Poker/issues/3)
    - Initial Go backend setup with Gorilla Mux router
    - Health check endpoint at `/health`
    - Structured JSON 404 error responses
    - Bruno API test collection for backend endpoints

### Changed
- Moved all backend API routes under the `/api/` prefix for better separation (e.g., `/api/health`).

### Fixed
