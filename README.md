<div align="center">

# RoyalFlush

**Online Poker Platform**

[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Angular](https://img.shields.io/badge/Angular-DD0031?style=flat&logo=angular&logoColor=white)](https://angular.io/)
[![SQLite](https://img.shields.io/badge/SQLite-07405E?style=flat&logo=sqlite&logoColor=white)](https://www.sqlite.org/)

</div>

---

## Project Description

RoyalFlush is a multiplayer Texas Hold'em poker application built with a Go backend and an Angular frontend. The current implementation supports account registration, cookie-based login sessions, protected frontend routes, room creation/join flows, and a WebSocket-backed lobby presence system.

The codebase is still in active development. Full poker gameplay is not complete on the backend yet; frontend game-state models and table services exist as integration groundwork for future game events.

## Current Implemented Features

- User registration with bcrypt password hashing.
- Login/logout using a signed HttpOnly `session-id` cookie.
- CSRF protection for mutating HTTP requests through `X-CSRF-Token`.
- Session restore through `GET /api/me`.
- Admin user list/delete endpoints protected by HTTP Basic Auth.
- Persistent room metadata in SQLite with 6-character uppercase alphanumeric room codes.
- Room list, create, lookup, and join-validation HTTP APIs.
- Authenticated WebSocket endpoint at `/ws`.
- Lobby WebSocket events for joining/leaving rooms and toggling ready state.
- Angular Material frontend with public routes, protected dashboard/lobby/table shell, auth guard, room create/join UI, lobby player list, ready toggle, and local chat UI.

## Not Yet Implemented

- Backend poker hand/game engine.
- Server-side chat broadcasting.
- Backend handling for frontend game action messages such as `CHECK`, `CALL`, `RAISE`, and `FOLD`.
- Password enforcement for private rooms. The backend stores the `isPrivate` flag but currently ignores room passwords.
- Persisted lobby ready state. Live membership and ready flags are in memory and reset on disconnect or server restart.

## Tech Stack

| Domain | Technology | Usage |
| --- | --- | --- |
| Backend | Go, Gorilla Mux, Gorilla CSRF, Gorilla SecureCookie, GORM | API, routing, auth/session cookies, database access |
| Frontend | Angular 21, Angular Material, RxJS, Lodash | UI, route protection, state streams, lobby rendering |
| Database | SQLite | Persistent users and room metadata |
| Realtime | Browser WebSocket API, Go WebSocket hub | Authenticated lobby presence and ready updates |

## Repository Layout

```text
Backend/                 Go API server and tests
frontend/                Angular application and tests
docs/api/openapi.yaml    Current HTTP API and WebSocket handshake contract
tools/bruno/             Bruno API collection
sprints/                 Sprint planning and retrospective notes
```

## API Summary

HTTP API routes are served under `/api` on `http://localhost:8080`.

| Route | Auth | Current behavior |
| --- | --- | --- |
| `GET /api/health` | Public | Server/database health |
| `GET /api/csrf` | Public | Returns CSRF token for mutating requests |
| `POST /api/register` | Public + CSRF | Creates a user; does not log the user in |
| `POST /api/login` | Public + CSRF | Sets signed HttpOnly session cookie; returns `204` |
| `POST /api/logout` | Public + CSRF | Clears session cookie; returns `204` |
| `GET /api/me` | Session cookie | Returns current user or `401` when no valid session exists |
| `GET /api/rooms` | Session cookie | Lists persisted rooms, optionally filtered by `status` |
| `POST /api/rooms` | Session cookie + CSRF | Creates a room for the current user |
| `POST /api/rooms/join` | Session cookie + CSRF | Validates an open, non-full room |
| `GET /api/rooms/{code}` | Session cookie | Looks up open room metadata |
| `GET /api/admin/users` | Basic Auth | Lists users |
| `DELETE /api/admin/users/{id}` | Basic Auth + CSRF | Deletes a user |

The OpenAPI contract is maintained at `docs/api/openapi.yaml`.

## WebSocket Summary

The WebSocket endpoint is `ws://localhost:8080/ws` locally. It requires the same signed session cookie as protected HTTP routes. Browser WebSocket connections send matching cookies automatically when the cookie domain and SameSite policy allow it.

Current client-to-server lobby events:

- `JOIN_ROOM` with `{ "roomCode": "AB12CD" }`
- `LEAVE_ROOM` with `{}`
- `TOGGLE_READY` with `{}`

Current server-to-client lobby events:

- `ROOM_STATE`
- `PLAYER_JOINED`
- `PLAYER_LEFT`
- `PLAYER_UPDATE`
- `ERROR`

## Local Setup

Run backend tests:

```bash
cd Backend
go test ./pkg/...
```

Run the backend locally:

```bash
cd Backend
GO_ENV=development go run ./cmd/server
```

Run frontend setup and tests:

```bash
cd frontend
nvm use 20
npm install
npm run test:unit
```

Run the frontend locally:

```bash
cd frontend
npm run start
```

Local frontend defaults to `http://localhost:4200`; local backend defaults to `http://localhost:8080`.

## Environment Notes

- `SESSION_KEY` must be 32 or 64 bytes outside development/test.
- `CSRF_AUTH_KEY` should be 32 bytes. A development fallback is used when missing.
- `APP_ENV=development` or `GO_ENV=development` disables the Secure flag for local cookies.
- `FRONTEND_URL` can add another allowed CORS origin in addition to `http://localhost:4200`.
- `ADMIN_USERNAME` and `ADMIN_PASSWORD` configure admin Basic Auth. In development, missing values fall back to `admin` / `admin`.
