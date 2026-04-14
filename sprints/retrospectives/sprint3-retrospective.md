# Sprint 3 Retrospective: Real-Time Lobby, Room Management, and UI Foundations

**Sprint Duration:** Sprint 3  
**Team Focus:** Real-time lobby features, room lifecycle, and frontend gameplay primitives  
**Analysis Window:** Changes after March 25, 2026 (from git history)  
**Date:** April 2026

---

## Executive Summary

Sprint 3 moved the project from "authenticated app" to "real-time multiplayer foundation." Based on git history after March 25, the team delivered major backend and frontend infrastructure for room creation/joining, lobby WebSocket communication, ready/chat interactions, and the first reusable gameplay UI component (generic card rendering).

**Key Outcomes:**
- Authentication flow was completed and hardened (login integration, `/api/me`, session package extraction).
- Backend routing was centralized and prepared for room/lobby endpoints.
- WebSocket runtime and lobby broadcasting logic were implemented and tested.
- Room domain expanded with model, service, validation, and generated room codes.
- Frontend gained Create/Join Room flows, Lobby page, and chat/ready UX.
- A reusable SVG-based Card component was added for upcoming table/gameplay work.

**Delivery Snapshot (from git since 2026-03-25):**
- 14 commits merged
- 9,212 lines added / 2,443 deleted (net +6,769)
- Churn by area: `frontend` (+3,992 net), `backend` (+2,382 net), `sprints` (+318 net)

---

## Evidence-Based Timeline

### Phase 1: Auth Completion and Test Foundations (Mar 25 - Mar 26)

**Relevant commits:**
- `98cd217` Added login component and backend integration (#48)
- `66c71de` Implemented dashboard "Me" API (#46)
- `0580fba` Added Cypress registration test (#50)

**What changed:**
- Login UI and auth interceptor behavior were tightened.
- Dashboard `/api/me` support was improved in backend auth/JWT flow.
- Cypress infrastructure and first E2E auth coverage were introduced.

---

### Phase 2: Session/Router Architecture Refactor (Apr 11)

**Relevant commits:**
- `8833289` Extracted session handling to auth package (#61)
- `fa1fa9d` Introduced API router and placeholder handlers (#63)
- `425033b` Added frontend WebSocket service with tests (#62)

**What changed:**
- Session logic moved into dedicated auth package with tests.
- `api.NewRouter()` centralized route setup and improved testability.
- Frontend real-time plumbing began via `WebSocketService` and typed models.

---

### Phase 3: Room Domain and Create/Join UX (Apr 12)

**Relevant commits:**
- `f394873` Added room model, service, and tests (#65)
- `2013798` Added create/join room UI (#64)
- `0722cc4` Added lobby component baseline (#66)

**What changed:**
- Backend room model/service were introduced with validation and tests.
- Frontend delivered dedicated create-room and join-room feature components.
- Dashboard was refactored to route into room flows.
- Initial lobby page and lobby-oriented E2E/unit tests were added.

---

### Phase 4: Room Codes, Lobby Runtime, and Chat/Ready (Apr 13)

**Relevant commits:**
- `9814f16` Room code generation and validation (#67)
- `9eeb57a` Lobby logic completion (#68 / #22)
- `05811c5` Lobby broadcasting runtime refinements (#75)
- `fd4bd15` Ready button and chat UI (#76)

**What changed:**
- Unique room code generation and normalization rules were added.
- Lobby runtime and broadcast behavior were implemented end-to-end.
- Frontend lobby added chat panel, ready toggles, and richer state updates.
- Backend WebSocket tests significantly expanded during these changes.

---

### Phase 5: Gameplay UI Primitive (Apr 13)

**Relevant commits:**
- `af0941e` Generic card SVG component (#77)
- `631c111` PR review fixes for card component

**What changed:**
- Reusable card rendering component with specs and styling was added.
- Follow-up fix commit improved quality after code review.

---

## What Went Well

- Vertical slice momentum: auth -> rooms -> lobby -> reusable gameplay UI happened in one sprint window.
- Strong testing discipline: large test additions accompanied service/runtime changes.
- Clear feature decomposition: Create Room, Join Room, Lobby, Card were implemented as separated units.
- Refactor timing was effective: session extraction and centralized router reduced duplication before real-time complexity increased.

---

## Lessons Learned

1. Real-time concurrency needs explicit guardrails early. Multiple commits in one day were required to stabilize lobby/broadcast behavior.
2. UI scope can inflate quickly in lobby features. Chat, ready state, and visual state management expanded complexity beyond initial page scaffolding.
3. Large style files become a maintenance signal. Lobby and home styles grew significantly and now trigger build-size warnings.
4. CommonJS dependencies should be monitored. `lodash` in lobby path currently causes Angular optimization bailout warnings.

---

## Technical Debt Register

1. Room API handlers still began as placeholders and need complete endpoint parity with frontend expectations.
2. WebSocket resilience (reconnect/backoff and reconnect state sync) remains limited.
3. Lobby state and chat persistence are still primarily in-memory concerns.
4. Style budget warning on home page indicates need to modularize/reduce SCSS footprint.
5. CommonJS `lodash` usage should be replaced by ESM-native or framework-native alternatives.

---

## Team and Velocity Metrics

### Commits by Contributor (since Mar 25)

- Himanshu PS: 6
- SaiShravanthReddy: 4
- B Sai Puneeth: 2
- Sanikommu Devi: 2

### Change Volume

| Area | Added | Deleted | Net |
|---|---:|---:|---:|
| Backend | 2619 | 237 | +2382 |
| Frontend | 6191 | 2199 | +3992 |
| Sprint Docs | 321 | 3 | +318 |
| Total | 9212 | 2443 | +6769 |

---

## Sprint 4 Recommendations

1. Complete production-ready Room REST handlers and align contracts with frontend service calls.
2. Add reconnect + resync strategy for WebSocket clients (player list, ready status, chat feed).
3. Start integrating the card component into table/game flows, not as isolated UI only.
4. Reduce style and bundle warnings by splitting oversized SCSS and removing CommonJS dependencies.
5. Add focused end-to-end tests for host actions and lobby-to-table transition.

---

## Conclusion

Git history after March 25 shows Sprint 3 as a high-throughput, architecture-heavy sprint that established the multiplayer core of RoyalFlush. The project now has the essential building blocks for live room-based gameplay: room lifecycle, lobby synchronization, and a reusable card UI foundation. The next sprint should focus on production hardening, gameplay progression, and optimization cleanup.
