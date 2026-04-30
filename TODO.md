# Frontend Todo List

## Critical — Broken Features

- [x] **Replace empty MP3 placeholder files** *(see note below — manual step required)*
  - `public/assets/sounds/chips-clink.mp3` — 0 bytes
  - `public/assets/sounds/card-flip.mp3` — 0 bytes
  - `public/assets/sounds/win-fanfare.mp3` — 0 bytes
  - All three fail silently when `SoundEffectsService` tries to play them. Replace with real royalty-free audio files.

- [x] **Fix the `/login` route — points to a dead placeholder**
  - Route now loads `LoginComponent` from `features/auth/login/login.component.ts`
  - Dead placeholder at `pages/login/` deleted

- [x] **Add auth guard to admin route**
  - `app.routes.ts` — `/admin` now has `canActivate: [authGuard]`

---

## High — Missing Tests

- [x] **Add spec file for `LoginComponent`**
  - `features/auth/login/login.component.spec.ts` created — 19 tests covering form validation,
    timeout handling, 401 vs server error distinction, successful login navigation, and isSubmitting guard

---

## Medium — Incomplete Implementations

- [x] **Build out the protected shell layout**
  - `layouts/protected-shell/protected-shell.html` now wraps `<router-outlet>` in a proper `.rf-shell` container with full-height flex layout

- [x] **Delete dead `RoomComponent`**
  - `pages/room/` fully removed (room.ts, room.html, room.scss)

---

## Low — Code Hygiene

- [x] **Fix unsubscribed form observable in `home.component.ts`**
  - No `valueChanges.subscribe()` exists in the current code — this item was already resolved

- [x] **Remove `console.error`/`console.warn` from production services**
  - `websocket.service.ts` — all console calls removed; malformed frames dropped silently; existing specs updated
  - `auth.service.ts` — all console calls removed from login, register, logout, and checkSession

---

## Audio files — [x] Done

The `SoundEffectsService` expects three MP3 files at:

```
frontend/public/assets/sounds/chips-clink.mp3  (~0.38s)
frontend/public/assets/sounds/card-flip.mp3    (~0.40s)
frontend/public/assets/sounds/win-fanfare.mp3  (~3.00s)
```

Real MP3 assets have been added and trimmed to the appropriate durations. No further action required.
