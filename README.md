<div align="center">

# 🃏 RoyalFlush

**Online Poker Platform**

[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Angular](https://img.shields.io/badge/Angular-DD0031?style=flat&logo=angular&logoColor=white)](https://angular.io/)
[![SQLite](https://img.shields.io/badge/SQLite-07405E?style=flat&logo=sqlite&logoColor=white)](https://www.sqlite.org/)

</div>

---

## 📖 Project Description

**RoyalFlush** is a multiplayer Texas Hold'em poker application designed to demonstrate modern full-stack development practices. The platform allows users to create private game rooms, invite friends via codes, and play poker in real-time.

The application emphasizes low-latency communication and a responsive user experience. It utilizes a **Go (Golang)** backend to handle game logic and WebSocket connections, paired with an **Angular** frontend for the interactive game interface. The architecture focuses on clean code, secure authentication, and reliable state synchronization using **GORM** for data management.

### 🌟 Key Features
* **Real-Time Gameplay:** Instant state updates using WebSockets.
* **Private Rooms:** Create rooms and invite friends via unique codes.
* **Secure Auth:** User registration and login protected by JWT and bcrypt.
* **Persistent Data:** User profiles and game history stored in SQLite.

---

## 🛠 Tech Stack

| Domain | Technology | Usage |
| :---: | :---: | :---: |
| **Backend** | ![Go](https://img.shields.io/badge/-Go-00ADD8?logo=go&logoColor=white&style=flat) **Gorilla Mux**, **GORM** | Game Engine, API, Routing, ORM |
| **Frontend** | ![Angular](https://img.shields.io/badge/-Angular-DD0031?logo=angular&logoColor=white&style=flat) **RxJS**, **Lodash** | Interactive UI, State Management, Utilities |
| **Database** | ![SQLite](https://img.shields.io/badge/-SQLite-07405E?logo=sqlite&logoColor=white&style=flat) | Persistent Data Storage |

---

## 👥 Team Members

| Role | Name |
| :---: | :---: |
| **Backend** | Sai Puneeth Bonagiri<br>Himanshu Potham Shetty<br>Sai Shravanth Reddy Madem |
| **Frontend** | Sai Shravanth Reddy Madem<br>Devi Sanikommu<br>Himanshu Potham Shetty |

---

## ✅ Current Story Status (Issue #19 WebSocket Service)

- `WebSocketService` implemented in `frontend/src/app/core/services/websocket.service.ts`
- Exposes `messages$` (incoming events) and `connected$` (connection state) observables
- `sendMessage(type, payload)` sends typed `{ type, payload }` JSON envelopes
- `WS_FACTORY` InjectionToken allows mock injection in unit tests (17 tests passing)
- Session cookies sent automatically by browser on WS handshake — no credential flag needed
- **Pending**: backend `ws://localhost:8080/ws` endpoint (future sprint)

---

## ✅ Previous Story Status (Issue #1-11 Registration UI)

- Registration UI is implemented with Angular Reactive Forms and Material components.
- Validation behavior is implemented (required/min length, submit disabled until valid).
- Backend `409` conflict is mapped to a user-facing `Username taken` message.
- Successful registration automatically logs the user in; on successful login, the user is navigated to `/dashboard`.

### Automated verification commands

Run frontend unit tests:

```bash
cd frontend
npm run test:unit
```

Run frontend E2E tests (requires running dev server):

```bash
# terminal 1
cd frontend
npm run start

# terminal 2 (in another terminal)
cd frontend
npm run e2e:headless
```

Run backend unit/API tests:

```bash
cd backend
go test ./pkg/...
```

Run frontend-backend integration smoke test (requires backend running):

```bash
# terminal 1
cd backend
GO_ENV=development go run ./cmd/server

# terminal 2
cd frontend
npm run test:integration:backend
```
