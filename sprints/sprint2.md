# Sprint 2: Room Management & Real-Time Lobby

**Sprint Goal:** Enable users to create private game rooms, join via unique codes, and interact in a real-time lobby. By the end of this sprint, players should be able to see each other join and toggle their "Ready" status instantly.

---

## Issue #1: [Backend] Room Model & Migration

### Linked Stories
* **Blocked by:** None
* **Blocks:** #16, #18

### Story Description
Define the `Room` data structure in GORM. A room represents a single game instance. It needs a unique 6-character alphanumeric code (for invitations), a Host ID (the creator), and a status (Waiting, Playing, Finished).

### Test Outline
* **Verify Schema:** The `rooms` table is created in SQLite with a unique index on the `code` column.
* **Verify Relationships:** Ensure the `HostID` correctly references the `users` table.

### Checklist
- [ ] Create `Room` struct in `pkg/models` (Code, HostID, Status, MaxPlayers).
- [ ] Add `rooms` to GORM AutoMigration in `main.go`.
- [ ] Create a helper function `GenerateRoomCode()` (in `pkg/utils` or the room service/handler layer) that produces a random 6-char string (A-Z, 0-9).
- [ ] Define Room Status constants (`WAITING`, `PLAYING`, `FINISHED`).

---

## Issue #2: [Backend] Room Management APIs (HTTP)

### Linked Stories
* **Blocked by:** #15
* **Blocks:** #20

### Story Description
Implement standard HTTP endpoints to Create and Find rooms. While the game happens over WebSockets, the initial "handshake" to create a room or validate an invite code happens over HTTP.
* `POST /api/rooms`: Creates a room and returns the code.
* `GET /api/rooms/{code}`: Checks if a room exists and is open.

### Test Outline
* **Verify Creation:** Authenticated user calls `POST /api/rooms` -> gets 200 OK + `{"code": "XYZ123"}`. DB shows new row.
* **Verify Lookup:** `GET /api/rooms/XYZ123` returns room details.
* **Verify Invalid Code:** `GET /api/rooms/BADCODE` returns 404.

### Checklist
- [ ] Implement `CreateRoomHandler`: Generate code -> Save to DB -> Return Code.
- [ ] Implement `GetRoomHandler`: specific query by Code (not ID).
- [ ] Ensure endpoints are protected by `AuthMiddleware`.
- [ ] Handle collision retries (if generated code already exists, generate again).

---

## Issue #3: [Backend] WebSocket Infrastructure (Hub & Upgrader)

### Linked Stories
* **Blocked by:** None
* **Blocks:** #18, #19

### Story Description
Set up the foundational WebSocket architecture using `gorilla/websocket`. This involves an `Upgrader` to switch HTTP connections to WS, and a central `Hub` (or Manager) to track all active client connections.
**Critical Security Note:** The Upgrader must validate the **HttpOnly Cookie** to authenticate the WS connection.

### Test Outline
* **Verify Upgrade:** A client connecting to `/ws` with a valid auth cookie gets a 101 Switching Protocols response.
* **Verify Auth Failure:** Connecting without the cookie returns 401 Unauthorized.
* **Verify Ping/Pong:** The connection stays alive with heartbeat messages.

### Checklist
- [ ] Install `github.com/gorilla/websocket`.
- [ ] Create `pkg/socket/hub.go` to manage `map[*Client]bool`.
- [ ] Create `pkg/socket/client.go` with `ReadPump` and `WritePump` goroutines.
- [ ] Implement `ServeWs` handler that upgrades the request.
- [ ] Extract User ID from the HttpOnly cookie during the handshake.

---

## Issue #4: [Backend] Lobby Logic & Broadcasting

### Linked Stories
* **Blocked by:** #15, #17
* **Blocks:** #21, #22

### Story Description
Implement the logic to map specific WebSocket connections to specific Rooms. When a user connects, they should send a `JOIN_ROOM` event. The backend must add them to that room's subscriber list and broadcast `PLAYER_JOINED` to everyone else in that room.

### Test Outline
* **Verify Isolation:** User A in Room 1 should NOT receive messages intended for Room 2.
* **Verify Broadcast:** When User B joins Room 1, User A (already in Room 1) receives a JSON message `{"type": "PLAYER_JOINED", "payload": ...}`.

### Checklist
- [ ] Extend `Hub` to support Room-based broadcasting (`rooms map[string]map[*Client]bool`).
- [ ] Define Message Protocol structs (`Message` with Type and Payload).
- [ ] Handle `JOIN_ROOM` event: Add client to room map.
- [ ] Handle `LEAVE_ROOM` / Disconnect: Remove client and notify others.

---

## Issue #5: [Frontend] WebSocket Service (RxJS)

### Linked Stories
* **Blocked by:** #17
* **Blocks:** #21

### Story Description
Create a robust Angular Service to manage the WebSocket connection. It should use `RxJS/webSocket` (or a custom wrapper around native `WebSocket`) to expose a stream of events. It needs to handle the connection lifecycle (connect, disconnect, auto-reconnect).

### Test Outline
* **Verify Connection:** Service successfully connects to `ws://localhost:8080/ws`.
* **Verify Send/Receive:** Can send a JSON object and log a received JSON object.
* **Verify Credentials:** Ensure `withCredentials: true` is respected so cookies are sent during handshake.

### Checklist
- [ ] Create `WebSocketService` in `src/app/core/services`.
- [ ] Implement `connect()` method.
- [ ] Expose a `messages$` Subject for components to subscribe to.
- [ ] Implement `sendMessage(type: string, payload: any)` helper.
- [ ] Ensure error handling (if connection drops, notify user).

---

## Issue #6: [Frontend] Create & Join Room UI

### Linked Stories
* **Blocked by:** #16
* **Blocks:** #21

### Story Description
Create the UI components for the "Lobby Entrance".
* **Create Room:** A simple button that calls the HTTP API and redirects to the lobby.
* **Join Room:** A form asking for the 6-digit code, validating it via HTTP, and then redirecting.

### Test Outline
* **Verify Create:** Click "Create" -> API success -> Navigate to `/room/{code}`.
* **Verify Join:** Enter Code -> API validates -> Navigate to `/room/{code}`.
* **Verify Validation:** Input should limit to 6 alphanumeric characters.

### Checklist
- [ ] Create `CreateRoomComponent` (Button + API call).
- [ ] Create `JoinRoomComponent` (Input Form + API call).
- [ ] Update `AppRoutingModule` to include `room/:code`.
- [ ] Integrate with `RoomService` (HTTP) created in this sprint.

---

## Issue #7: [Frontend] Lobby Component & State

### Linked Stories
* **Blocked by:** #18, #19, #20
* **Blocks:** #23

### Story Description
The main Lobby View where players gather. This component subscribes to `WebSocketService`. When it initializes, it sends `JOIN_ROOM`. It maintains a list of players.
**Lodash Usage:** Use `_.uniqBy` to prevent duplicate players in the list if the backend sends redundant join events, and `_.orderBy` to sort the host to the top.

### Test Outline
* **Verify Player List:** Open two browser windows. Join the same room. Both windows should list 2 players.
* **Verify Deduplication:** If the backend accidentally sends "User A Joined" twice, the UI should still only show User A once.

### Checklist
- [ ] Create `LobbyComponent`.
- [ ] On `ngOnInit`, connect WS and send `JOIN_ROOM`.
- [ ] Render list of players (Avatar + Name).
- [ ] Handle `PLAYER_JOINED` and `PLAYER_LEFT` events to update the local array.
- [ ] Use `lodash` for array manipulation.

---

## Issue #8: [Backend] Player Ready Toggle Logic

### Linked Stories
* **Blocked by:** #18
* **Blocks:** #23

### Story Description
Add state management for "Ready/Not Ready". When a user clicks Ready, the backend must update their status in memory (or Redis/DB) and broadcast `PLAYER_UPDATE` to the room so others see the green checkmark. Game start logic (checking if all are ready) can be a stub for now.

### Test Outline
* **Verify State Change:** Send `TOGGLE_READY`. Backend broadcasts updated user object with `isReady: true`.
* **Verify Persistence:** If I disconnect and reconnect, my ready state should arguably reset (or persist, depending on design. Let's reset for simplicity).

### Checklist
- [ ] Add `IsReady` boolean to the `Client` or `Player` struct in Backend.
- [ ] Handle `TOGGLE_READY` message type.
- [ ] Broadcast `PLAYER_UPDATE` with the specific user's new status.
- [ ] (Optional) Check if `AllReady == true` (Stub for Sprint 3).

---

## Issue #9: [Frontend] Ready Button & Chat UI

### Linked Stories
* **Blocked by:** #21, #22
* **Blocks:** None

### Story Description
Final Polish for the lobby.
1.  **Ready Button:** Toggles the user's state. Changes color (Green/Grey) based on status.
2.  **Chat:** Simple text input to send messages to the room.

### Test Outline
* **Verify Ready UI:** Click Ready -> Icon turns green. Other players see my icon turn green.
* **Verify Chat:** Type "Hello" -> Enter. Text appears in chat box for all users in the room.

### Checklist
- [ ] Add "Ready" button to `LobbyComponent`.
- [ ] Add Chat Box (Input + Scrollable Message List).
- [ ] Listen for `CHAT_MESSAGE` events from WebSocket.
- [ ] Use `Angular Material` List for the chat display.