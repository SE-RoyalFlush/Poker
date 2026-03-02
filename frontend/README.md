# Poker Frontend

## Frontend Baseline Decisions

### Node policy
- Required: **Node 20 LTS**
- `.nvmrc` in repo root is the source of truth (`20`)
- Enforced via:
  - `package.json` engines: `"node": "20.x"`
  - `.npmrc`: `engine-strict=true`

Run:

```bash
nvm use 20
```

before any `npm` command.

### Angular version policy
- Angular packages are pinned to `21.1.4`
- No floating (`^`) Angular versions are allowed

### UI library decision
- Selected library: **Angular Material**
- Installed packages:
  - `@angular/material@21.1.4`
  - `@angular/cdk@21.1.4`
  - `@angular/animations@21.1.4`
- Configured:
  - Theme import in `src/styles.scss`
  - Animations provider in `src/app/app.config.ts`
- Verification component:
  - `src/app/pages/login/login.html` uses `mat-card` and a Material button

### Routing architecture
- Public routes:
  - `/login`
  - `/register`
- Protected-shell routes (guard intentionally deferred to auth story):
  - `/dashboard`
  - `/lobby`
  - `/table/:id`
- Redirects:
  - `/` -> `/login`
  - wildcard `**` -> `/login`
- Route config: `src/app/app.routes.ts`

### Scope boundary
- Placeholder-only baseline for setup and architecture
- No authentication logic
- No feature logic

## Setup and verification

```bash
nvm use 20
npm install
npm run build
npm start
```

## Running Tests

### Run all tests
```bash
npm test
```

### Run tests in watch mode
```bash
npm test -- --watch
```

### Run tests with code coverage report
```bash
npm test -- --coverage
```

### Run tests with verbose output
```bash
npm test -- --verbose
```

### Run tests for a specific file
```bash
npm test -- auth.interceptor.spec.ts
```

### Run tests without coverage (faster)
```bash
npm test -- --no-coverage
```

> **Note:** These tests are automatically run on every pull request via the GitHub Actions workflow (`.github/workflows/pull_request_test.yaml`). The workflow ensures all frontend unit tests pass before merging.

