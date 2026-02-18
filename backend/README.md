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
