package room_test

import (
	"errors"
	"testing"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
)

func TestLobbyToggleReadyBroadcastsPlayerUpdate(t *testing.T) {
	lobby := room.NewLobby()
	player := lobby.JoinPlayer(room.Player{
		ID:       7,
		Username: "alice",
		IsHost:   true,
	})

	if player.IsReady {
		t.Fatal("expected joined player to start not ready")
	}

	event, err := lobby.HandleMessage(room.MessageTypeToggleReady, player.ID)
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}

	if event.Type != room.MessageTypePlayerUpdate {
		t.Fatalf("expected event type %q, got %q", room.MessageTypePlayerUpdate, event.Type)
	}
	if event.Payload.ID != player.ID {
		t.Fatalf("expected payload player ID %d, got %d", player.ID, event.Payload.ID)
	}
	if !event.Payload.IsReady {
		t.Fatal("expected payload to mark player ready")
	}

	updated, ok := lobby.Player(player.ID)
	if !ok {
		t.Fatalf("expected player %d to remain in lobby", player.ID)
	}
	if !updated.IsReady {
		t.Fatal("expected ready state to persist in lobby")
	}
}

func TestLobbyJoinPlayerResetsReadyStateOnReconnect(t *testing.T) {
	lobby := room.NewLobby()
	player := lobby.JoinPlayer(room.Player{
		ID:       9,
		Username: "bob",
	})

	if _, err := lobby.ToggleReady(player.ID); err != nil {
		t.Fatalf("ToggleReady returned error: %v", err)
	}

	rejoined := lobby.JoinPlayer(room.Player{
		ID:       player.ID,
		Username: player.Username,
	})

	if rejoined.IsReady {
		t.Fatal("expected rejoined player ready state to reset")
	}

	stored, ok := lobby.Player(player.ID)
	if !ok {
		t.Fatalf("expected player %d after rejoin", player.ID)
	}
	if stored.IsReady {
		t.Fatal("expected stored player ready state to be reset on rejoin")
	}
}

func TestLobbyPlayersReturnsDeterministicRosterOrder(t *testing.T) {
	lobby := room.NewLobby()
	lobby.JoinPlayer(room.Player{ID: 3, Username: "zoe"})
	lobby.JoinPlayer(room.Player{ID: 1, Username: "host", IsHost: true})
	lobby.JoinPlayer(room.Player{ID: 2, Username: "alice"})

	players := lobby.Players()
	if len(players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(players))
	}

	expected := []struct {
		id       uint
		username string
		isHost   bool
	}{
		{id: 1, username: "host", isHost: true},
		{id: 2, username: "alice", isHost: false},
		{id: 3, username: "zoe", isHost: false},
	}

	for i, want := range expected {
		if players[i].ID != want.id || players[i].Username != want.username || players[i].IsHost != want.isHost {
			t.Fatalf("unexpected player at index %d: got %+v want id=%d username=%q isHost=%t", i, players[i], want.id, want.username, want.isHost)
		}
	}
}

func TestLobbyToggleReadyReturnsErrorForUnknownPlayer(t *testing.T) {
	lobby := room.NewLobby()

	_, err := lobby.ToggleReady(404)
	if !errors.Is(err, room.ErrPlayerNotFound) {
		t.Fatalf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestLobbyAllReady(t *testing.T) {
	lobby := room.NewLobby()
	lobby.JoinPlayer(room.Player{ID: 1, Username: "alice"})
	lobby.JoinPlayer(room.Player{ID: 2, Username: "bob"})

	if lobby.AllReady() {
		t.Fatal("expected AllReady to be false before toggles")
	}

	if _, err := lobby.ToggleReady(1); err != nil {
		t.Fatalf("ToggleReady returned error: %v", err)
	}
	if lobby.AllReady() {
		t.Fatal("expected AllReady to remain false until all players are ready")
	}

	if _, err := lobby.ToggleReady(2); err != nil {
		t.Fatalf("ToggleReady returned error: %v", err)
	}
	if !lobby.AllReady() {
		t.Fatal("expected AllReady to be true once all players are ready")
	}
}
