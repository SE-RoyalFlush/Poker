# Poker Backend Setup

This guide will help you set up and run the Poker backend server locally.

## Prerequisites

- **Go 1.21 or higher** installed on your machine
- **Git** for version control
- A terminal/command prompt

## Installation Steps

### 1. Clone the Repository

```bash
git clone https://github.com/SE-RoyalFlush/Poker.git
cd Poker/backend
```

### 2. Verify Go Modules

```bash
go mod download
go mod tidy
```

### 3. Run

```bash
go run ./cmd/server
```

### 4. Run Tests
```bash
go clean -testcache

# Run all tests
go test ./pkg/...

# Run tests with coverage
go test -cover ./pkg/...
```

## Live Room State Ownership

Room metadata is persistent and lives in the database through the `rooms` table:
room code, host user, status, maximum players, and privacy flag. HTTP room
endpoints create and look up that metadata from the database, so `GET
/api/rooms/{code}` does not depend on any active WebSocket connection.

Live room state is transient and owned by the WebSocket hub in memory for Sprint
2: active client membership and player ready flags. When a socket disconnects,
that player's transient lobby state is removed and their ready flag is reset.
If the process restarts, all active membership and ready state starts empty
while persisted room metadata remains in the database.

Future Redis support should move only this transient hub state to Redis or a
similar shared runtime store. The database should remain the source of truth for
room metadata.
