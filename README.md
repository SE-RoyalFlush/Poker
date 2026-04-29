<div align="center">

# RoyalFlush

Online poker platform built with Go, Angular, and SQLite.

[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Angular](https://img.shields.io/badge/Angular-DD0031?style=flat&logo=angular&logoColor=white)](https://angular.io/)
[![SQLite](https://img.shields.io/badge/SQLite-07405E?style=flat&logo=sqlite&logoColor=white)](https://www.sqlite.org/)

</div>

## Current State

RoyalFlush currently ships:

- Cookie-based authentication with signed `session-id` cookies
- User registration, login, logout, and session restore via `/api/me`
- Protected room APIs for creating, listing, joining, and fetching rooms
- Authenticated WebSocket handshake at `/ws`
- Lobby-style real-time room presence and ready-state updates over WebSockets
- SQLite-backed persistence for users and room metadata

The codebase already includes room and WebSocket features. Documentation in this
repo should describe the implemented session-cookie auth flow and the current
room/WebSocket contracts rather than earlier JWT- or placeholder-based plans.

## Architecture

| Area | Stack | Notes |
| --- | --- | --- |
| Backend | Go, Gorilla Mux, GORM | REST API, session auth, room lifecycle, WebSocket upgrade |
| Frontend | Angular, RxJS, Angular Material | Auth flows, dashboard/lobby/room UI, WebSocket client |
| Database | SQLite | Users and persisted room metadata |

## Auth Model

- Login sets an HttpOnly signed session cookie
- Protected HTTP routes are enforced by auth middleware and return `401` when the
  cookie is missing or invalid
- Frontend requests use `withCredentials: true`
- The WebSocket handshake reuses the same authenticated browser cookie

Implementation references:

- Backend session auth: `backend/pkg/auth/session.go`
- Auth handlers: `backend/pkg/api/auth.go`
- Route protection: `backend/pkg/middleware/auth.go`
- Frontend auth service: `frontend/src/app/core/services/auth.service.ts`
- Frontend WebSocket client: `frontend/src/app/core/services/websocket.service.ts`

## Room And WebSocket Scope

- Room metadata is persisted through the database
- Active room membership and ready flags are tracked in the in-memory WebSocket hub
- `POST /api/rooms/join` validates room availability; active membership is
  established after the client sends `JOIN_ROOM` over `/ws`
- `GET /api/rooms/{code}` returns persisted room metadata without requiring an
  active socket for that room

See `backend/README.md` for the persisted-vs-transient room state split.

## Verification

Frontend unit tests:

```bash
cd frontend
npm run test:unit
```

Backend unit/API tests:

```bash
cd backend
go test ./pkg/...
```

Frontend-backend integration smoke test:

```bash
# terminal 1
cd backend
GO_ENV=development go run ./cmd/server

# terminal 2
cd frontend
npm run test:integration:backend
```
