# Poker Backend Setup

This guide will help you set up and run the Poker backend server locally.

## Prerequisites

- **Go 1.21 or higher** installed on your machine
- **Git** for version control
- A terminal/command prompt

## Installation Steps

### 1. Clone the Repository

```bash
git clone https://github.com/pshimanshu/Poker.git
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

### 4. API Quick Check
Health Check: Expected `200 OK`
```bash
curl -i http://localhost:8080/health
```
