# Sprint 1: Foundation & Secure Authentication

**Sprint Goal:** Establish a production-ready application skeleton with secure, HttpOnly cookie-based authentication. By the end of this sprint, a user must be able to securely register, login, access a protected route, and logout.

---

## Issue #1: [Backend] Initialize Go Module & HTTP Router

### Linked Stories
* **Blocked by:** None
* **Blocks:** #2, #3, #4

### Story Description
We need to initialize the Go project structure to support a clean architecture. This involves setting up the module, organizing folders (handlers, models, middleware), and configuring the `gorilla/mux` router to handle basic HTTP requests.

### Test Outline
* **Verify Server Start:** Ensure the application starts on port 8080 without panic.
* **Verify Health Check:** Call `GET /health` and receive a `200 OK` status with JSON body `{"status": "alive"}`.
* **Verify 404:** Call a random route (e.g., `/random`) and ensure it returns a structured JSON 404 error, not the default text.

### Checklist
- [ ] Initialize `go.mod` with project name `github.com/SE-RoyalFlush/Poker/backend`.
- [ ] Create folder structure: `cmd/server`, `pkg/api`, `pkg/models`, `pkg/middleware`.
- [ ] Install `github.com/gorilla/mux`.
- [ ] Implement `main.go` to initialize the server.
- [ ] Create a simple `/health` handler in `pkg/api/health.go`.
- [ ] Configure `http.Server` with reasonable timeouts (Read/Write) to prevent potential leaks.

---