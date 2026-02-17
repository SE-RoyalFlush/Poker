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

