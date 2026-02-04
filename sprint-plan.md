### **Sprint 1: Foundation & Authentication (Weeks 1-2)**

**Goal:** Establish the project skeleton, database connection, and secure user access.
#### **Backend (Go + Gorilla Mux + GORM + SQLite)**
- **Project Setup:** Initialize Go module. Configure **Gorilla Mux** router.
- **Database Init:** Set up **SQLite** connection using **GORM**.
- **User Model:** Define the `User` struct (ID, Username, PasswordHash) and auto-migrate schema using GORM.
- **Authentication APIs:**
    - `POST /register`: Accept JSON, hash password (bcrypt), create user in SQLite.
    - `POST /login`: Verify credentials, generate **JWT**, and set it (or a refresh token) in an **HttpOnly**, **Secure**, `SameSite=strict` cookie rather than exposing it to JavaScript.
- **Middleware:** Create an Auth Middleware function to validate JWTs from cookies on protected routes and implement CSRF protections (e.g., same-site cookies plus CSRF token) and basic XSS mitigations (input validation, safe templating, CSP).
#### **Frontend (Angular + RxJS)**
- **Project Initialization:** Create Angular app with routing enabled.
- **Auth Service:** Implement `AuthService` to handle HTTP calls and rely on JWTs sent via HttpOnly cookies (optionally keeping only a short-lived access token in memory); avoid storing JWTs in `localStorage` or other JavaScript-accessible persistent storage.
- **Auth Views:** Create `LoginComponent` and `RegisterComponent` with form validation.
- **Routing:** Set up `AuthGuard` to redirect unauthenticated users to Login.
- **Dashboard Shell:** Create a basic `HomeComponent` that displays the logged-in username.

> **✅ Viewable Changes:** A user can register, log in, and land on a personalized "Home" screen. Trying to access the home screen without logging in redirects to login.

---

### **Sprint 2: Room Management & Real-Time Lobby (Weeks 3-4)**

**Goal:** Enable users to group together in private rooms and see each other in real-time.
#### **Backend (Go + Gorilla WebSocket)**
- **Room Model:** Define `Room` struct in GORM (RoomCode, HostID, IsActive).
- **Room APIs:**
    - `POST /rooms`: Generate a unique 6-character code, save to SQLite.
    - `GET /rooms/{code}`: Validate room existence.
- **WebSocket Infrastructure:**
    - Create a `Hub` struct to manage active client connections.
    - Implement `Upgrader` to switch HTTP requests to WebSocket.
    - Define Message Protocol (e.g., `{"type": "JOIN_ROOM", "payload": ...}`).
- **Lobby Logic:** Handle players joining/leaving. Broadcast the updated player list to all clients in that specific room.
#### **Frontend (Angular + Lodash)**
- **Room UI:**
    - `CreateRoomComponent`: Button to generate a room.
    - `JoinRoomComponent`: Input field for Room Code.
- **WebSocket Service:** Create `WebSocketService` using `RxJS/webSocket` or native API to maintain the connection.
- **Lobby Component:**
    - Display Room Code.
    - Display list of connected players.
    - **Lodash Usage:** Use `_.uniqBy()` or `_.orderBy()` to sort and deduplicate the player list efficiently before rendering.

> **✅ Viewable Changes:** User A creates a room and shares the code with User B. User B enters the code. Both screens instantly update to show both names in the "Lobby."

---

### **Sprint 3: Core Card System & Poker Implementation (Weeks 5-6)**

**Goal:** Implement a generic card engine and build the Texas Hold'em rules on top of it.
#### **Backend (Go Engine)**
- **Generic Card Package:** Create a standalone package `pkg/cards` defining `Card` (Suit, Rank) and `Deck` structs.
- **Deck Operations:** Implement generic `Shuffle()` and `Draw(n)` methods within the `cards` package (independent of game rules).
- **Poker Game State:** Create a separate `pkg/poker` that imports `pkg/cards`. Define `PokerState` which utilizes the generic `Deck`.
- **Poker Logic:**
    - **Hand Evaluator:** Algorithm specific to Poker to rank hands (Royal Flush, Full House) using the generic `Card` objects.
    - **Betting System:** Handle `Check`, `Call`, `Raise`, `Fold` actions.
    - **Phase Management:** Manage `Pre-Flop` → `Flop` → `Turn` → `River`.
#### **Frontend (Angular + Animations)**
- **Generic Card Component:** Create a reusable `CardComponent` that accepts `rank` and `suit` as inputs and renders the SVG (unaware of the specific game).
- **Poker Table View:** A container component that arranges the generic `CardComponents` into a Texas Hold'em layout (Community cards center, Player cards bottom).
- **Game Controls:** Buttons for Check, Call, Raise, Fold linked to the Poker service.
- **State Visualization:**
    - Display Pot size.
    - Indicate "Active Player".
    - **Lodash Usage:** Use `_.find()` to quickly locate the active player object from the state array.

> **✅ Viewable Changes:** Players can play a complete hand. The system uses a universal deck to deal cards, but applies Poker-specific rules for winning and betting.

---

### **Sprint 4: Persistence, Polish & Stats (Weeks 7-8)**

**Goal:** Save game history, display statistics, and polish the user experience.
#### **Backend (Go + GORM + SQLite)**
- **History Model:** Define `GameResult` struct (WinnerID, PotSize, Date, GameType).
- **Persistence:** At the end of a hand, save the result to SQLite using GORM.
- **Stats API:** `GET /users/{id}/stats` to calculate wins/losses and total earnings.
- **Cleanup:** Implement logic to close empty rooms and clean up memory in the WebSocket Hub.
#### **Frontend (Angular + Polish)**
- **Profile Page:** Display user statistics (Hands Played, Win Rate) fetched from the backend.
- **Leaderboard:** A table showing top players (sorted using **Lodash** `_.orderBy` on the frontend or DB query).
- **UX Improvements:**
    - Sound effects (chips clinking, cards flipping).
    - Toast notifications ("Player X won the pot!").
- **Responsiveness:** Ensure the poker table scales down for mobile screens.

> **✅ Viewable Changes:** A "Leaderboard" page shows top players. Past game results are saved even after server restart. The interface feels smooth and responsive.
