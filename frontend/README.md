# Poker Frontend

Angular frontend for RoyalFlush. The app currently supports registration, login/logout, protected navigation, dashboard room creation/joining, lobby presence over WebSocket, ready toggling, and frontend game-state scaffolding.

## Runtime Policy

### Node

- Required: Node 20 LTS
- `.nvmrc` in the repo root is the source of truth
- `package.json` engines require `"node": "20.x"`
- `.npmrc` has `engine-strict=true`

Run before npm commands:

```bash
nvm use 20
```

### Angular

- Angular packages are pinned to `21.1.4`
- Angular Material, CDK, and animations are pinned to `21.1.4`
- Floating Angular package versions are not used

## Setup

```bash
npm install
npm run start
```

The dev server runs at `http://localhost:4200`. Local API calls target `http://localhost:8080/api`, and the local WebSocket endpoint is `ws://localhost:8080/ws`.

## Routing

Route config lives in `src/app/app.routes.ts`.

| Route | Access | Component |
| --- | --- | --- |
| `/home` | Public | Home page |
| `/login` | Public | Login page |
| `/register` | Public | Registration page |
| `/admin` | Public UI; backend uses Basic Auth | Admin page |
| `/dashboard` | Protected by `authGuard` | Dashboard with create/join room UI |
| `/lobby/:code` | Protected by `authGuard` | WebSocket lobby |
| `/table/:id` | Protected by `authGuard` | Table shell/game-state integration target |
| `/` | Redirect | `/home` |
| `**` | Redirect | `/home` |

`authGuard` checks the in-memory `AuthService` current user and redirects unauthenticated users to `/`.

## API Configuration

`src/app/core/config/endpoints.ts` defines endpoint selection:

- Localhost frontend:
  - `API_URL = http://localhost:8080/api`
  - `WS_URL = ws://localhost:8080/ws`
- Non-localhost frontend:
  - `API_URL = /api`
  - `WS_URL = ws(s)://<current-host>/ws`

## Authentication

`AuthService` lives at `src/app/core/services/auth.service.ts` and is the single source of truth for frontend user state.

Current behavior:

- `checkSession()` calls `GET /api/me` with credentials and restores the current user when the backend returns one.
- `login()` posts username/password to `POST /api/login`, then calls `checkSession()` because login returns `204` and no body.
- `register()` posts to `POST /api/register` and returns the created user; registration does not imply an authenticated session.
- `logout()` posts to `POST /api/logout` and clears local user state even if the backend call fails.
- `currentUser$` exposes user state to components.
- `isAuthenticated()` and `getCurrentUser()` support guards and quick synchronous reads.

The backend uses an HttpOnly signed `session-id` cookie. The frontend does not read or store JWTs.

## CSRF and HTTP Interceptor

`AuthInterceptor` lives at `src/app/core/interceptors/auth.interceptor.ts`.

Current behavior:

- Adds `withCredentials: true` to all HTTP requests.
- For `POST`, `PUT`, `PATCH`, and `DELETE`, ensures a CSRF token by calling `GET /api/csrf` when needed.
- Sends the token in the `X-CSRF-Token` header.
- Clears cached CSRF token on `401`.
- Redirects to `/login` on `401` except for session-check requests.

## Room UI and Service

`RoomService` lives at `src/app/core/services/room.service.ts`.

| Method | Backend route | Notes |
| --- | --- | --- |
| `createRoom(payload)` | `POST /api/rooms` | Backend currently consumes `maxPlayers` and `isPrivate`; UI fields like room name, blinds, and password are frontend-only for now |
| `joinRoom(code, password?)` | `POST /api/rooms/join` | Backend validates room existence/open/full state; password is ignored until private-room enforcement is implemented |
| `getLiveRooms()` | `GET /api/rooms?status=open` | Returns persisted rooms plus current in-memory WebSocket occupancy counts |

Room codes use the shared 6-character uppercase alphanumeric format, for example `AB12CD`.

## WebSocket Service

`WebSocketService` lives at `src/app/core/services/websocket.service.ts`.

Current behavior:

- Uses the browser `WebSocket` API.
- Exposes `messages$` for incoming events.
- Exposes `connected$` for connection state.
- Sends messages as `{ type, payload }`.
- Reconnects unexpected disconnects with exponential backoff, capped at 5 attempts.
- Provides a `WS_FACTORY` injection token for unit tests.

Browser WebSocket handshakes automatically include matching cookies when domain and SameSite rules allow it. There is no `withCredentials` flag for WebSocket.

## Lobby

The lobby page lives at `src/app/pages/lobby/`.

Current behavior:

- Reads the room code from `/lobby/:code`.
- Connects to `/ws`.
- Sends `JOIN_ROOM` after the socket opens.
- Handles `ROOM_STATE`, `PLAYER_JOINED`, `PLAYER_LEFT`, and `PLAYER_UPDATE`.
- Sorts host players first.
- Sends `TOGGLE_READY` when the ready button changes.
- Provides local chat UI and sends `CHAT_MESSAGE`, but backend chat broadcasting is not implemented yet.
- Disconnects the WebSocket when leaving the lobby.

## WebSocket Message Models

Shared frontend message types live in `src/app/core/models/ws-message.model.ts`.

Backend-supported lobby messages:

- Client to server: `JOIN_ROOM`, `LEAVE_ROOM`, `TOGGLE_READY`
- Server to client: `ROOM_STATE`, `PLAYER_JOINED`, `PLAYER_LEFT`, `PLAYER_UPDATE`, `ERROR`

Frontend-only game scaffolding currently includes message models and `GameStateService` handling for:

- `GAME_STARTED`
- `CARDS_DEALT`
- `PLAYER_ACTION`
- `PHASE_CHANGE`
- `GAME_OVER`
- `CHECK`
- `CALL`
- `RAISE`
- `FOLD`

Those poker gameplay events are not handled by the backend yet.

## Testing

Run unit tests:

```bash
npm run test:unit
```

Run Cypress interactively:

```bash
npm run e2e
```

Run Cypress headlessly:

```bash
npm run e2e:headless
```

Cypress expects the Angular dev server at `http://localhost:4200`.

Useful focused runs:

```bash
npm test -- --watch=false --browsers=ChromeHeadless --include='**/auth.service.spec.ts'
npm test -- --watch=false --browsers=ChromeHeadless --include='**/websocket.service.spec.ts'
npm test -- --watch=false --browsers=ChromeHeadless --include='**/lobby.spec.ts'
```

## Backend/Frontend Contract Matrix

| Contract Area | Source of Truth | Expected Behavior | Frontend Verification |
| --- | --- | --- | --- |
| User JSON | `Backend/pkg/models/user.go`, `docs/api/openapi.yaml` | Backend returns `ID`, `CreatedAt`, `UpdatedAt`, `username` | `AuthService` maps to `User` |
| Session restore | `GET /api/me` | `200` returns user; `401` means no session in the routed API | `checkSession()` tests |
| Login | `POST /api/login` | `204`, sets HttpOnly session cookie | `login()` then `checkSession()` tests |
| Register | `POST /api/register` | `201` returns created user; does not set session | `register()` tests |
| Logout | `POST /api/logout` | `204`, clears backend cookie | `logout()` tests |
| Rooms | `/api/rooms*` | Session-protected room metadata APIs | `room.service.spec.ts`, create/join component specs |
| Lobby WebSocket | `/ws` | Session-protected socket with lobby messages | `websocket.service.spec.ts`, `lobby.spec.ts` |
