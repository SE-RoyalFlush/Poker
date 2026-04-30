# Poker Backend

The backend is a Go HTTP API using Gorilla Mux, Gorilla CSRF, signed HttpOnly session cookies, GORM, and SQLite. It currently owns authentication, persistent room metadata, admin user operations, and the authenticated WebSocket lobby runtime.

## Prerequisites

- Go 1.21 or newer
- Git
- SQLite-compatible local filesystem storage

## Setup

```bash
cd Backend
go mod download
go mod tidy
```

## Run

```bash
GO_ENV=development go run ./cmd/server
```

The server listens on `http://localhost:8080`.

## Test

```bash
go clean -testcache
go test ./pkg/...
go test -cover ./pkg/...
```

## Authentication

The backend does not expose JWTs to the frontend. Login sets a signed HttpOnly cookie named `session-id`:

- `POST /api/login` validates username/password and returns `204 No Content`.
- `GET /api/me` resolves the signed cookie and returns the current user.
- `POST /api/logout` clears the cookie and returns `204 No Content`.
- Protected API routes use `middleware.AuthMiddleware`.
- `/ws` also resolves the same signed cookie before upgrading.

`SESSION_KEY` must be 32 or 64 bytes outside development/test. In development and test, the server can use an insecure fallback key.

## CSRF and CORS

The server wraps the router with Gorilla CSRF:

- `GET /api/csrf` returns `{ "csrfToken": "..." }` and exposes the token in the `X-CSRF-Token` response header.
- Mutating requests must send `X-CSRF-Token`.
- CSRF cookies are HttpOnly, `SameSite=Lax`, and Secure unless `APP_ENV=development` or `GO_ENV=development`.
- CORS allows credentials from `http://localhost:4200` and from `FRONTEND_URL` when configured.

## Public HTTP Routes

| Method | Path | Behavior |
| --- | --- | --- |
| `GET` | `/api/health` | Server/database health |
| `GET` | `/api/csrf` | CSRF token |
| `POST` | `/api/register` | Create user |
| `POST` | `/api/login` | Set session cookie |
| `POST` | `/api/logout` | Clear session cookie |
| `GET` | `/api/admin/users` | List users with Basic Auth |
| `DELETE` | `/api/admin/users/{id}` | Delete user with Basic Auth |

Admin routes use HTTP Basic Auth. `ADMIN_USERNAME` and `ADMIN_PASSWORD` are required outside development-like environments; development falls back to `admin` / `admin`.

## Session-Protected HTTP Routes

| Method | Path | Behavior |
| --- | --- | --- |
| `GET` | `/api/me` | Current authenticated user, or `401` when no valid session exists |
| `GET` | `/api/rooms` | List rooms, optional `?status=open|closed|in_game` |
| `POST` | `/api/rooms` | Create room for current user |
| `POST` | `/api/rooms/join` | Validate an open, non-full room before WebSocket join |
| `GET` | `/api/rooms/{code}` | Look up persisted open room metadata |

Room codes are exactly 6 uppercase alphanumeric characters. Input codes are trimmed and normalized to uppercase before lookup.

## Room State Ownership

Room metadata is persistent and lives in the `rooms` table: room code, host user, status, maximum players, and privacy flag. HTTP room endpoints create and look up that metadata from the database, so `GET /api/rooms/{code}` does not depend on an active WebSocket connection.

Live room state is transient and owned by the WebSocket hub in memory: active client membership and player ready flags. When a socket disconnects, that player's transient lobby state is removed and their ready flag is reset. If the process restarts, all active membership and ready state starts empty while persisted room metadata remains in the database.

Private-room password enforcement is not implemented yet. `POST /api/rooms` stores `isPrivate`, and `POST /api/rooms/join` ignores any supplied password.

Future Redis support should move only transient hub state to Redis or a similar shared runtime store. The database should remain the source of truth for room metadata.

## WebSocket Contract

`GET /ws` upgrades an authenticated request. Unauthenticated requests receive a JSON `401` instead of upgrading.

All messages use:

```json
{ "type": "EVENT_TYPE", "payload": {} }
```

Client-to-server events currently handled:

- `JOIN_ROOM` with `{ "roomCode": "AB12CD" }`
- `LEAVE_ROOM` with `{}`
- `TOGGLE_READY` with `{}`

Server-to-client events currently emitted:

- `ROOM_STATE` with `{ "roomCode": "AB12CD", "players": [...] }`
- `PLAYER_JOINED` with a player payload
- `PLAYER_LEFT` with `{ "id": 1 }`
- `PLAYER_UPDATE` with a player payload
- `ERROR` with `{ "code": "...", "message": "..." }`

Poker gameplay WebSocket events are not handled by the backend yet.
