package api

import (
	"encoding/json"
	"net/http"
)

type demoCard struct {
	Rank string `json:"rank"`
	Suit string `json:"suit"`
}

type demoSeat struct {
	SeatIndex     int        `json:"seatIndex"`
	PlayerID      int        `json:"playerId"`
	Username      string     `json:"username"`
	ChipCount     int        `json:"chipCount"`
	HoleCards     []demoCard `json:"holeCards"`
	IsCurrentUser bool       `json:"isCurrentUser"`
	IsActive      bool       `json:"isActive"`
	IsDealer      bool       `json:"isDealer"`
}

type demoGameState struct {
	Phase          string     `json:"phase"`
	CommunityCards []demoCard `json:"communityCards"`
	Seats          []demoSeat `json:"seats"`
	Pot            int        `json:"pot"`
	CurrentBet     int        `json:"currentBet"`
	CurrentUserID  int        `json:"currentUserId"`
}

// DemoGameStateHandler returns a static demo game state for UI prototyping.
func DemoGameStateHandler(w http.ResponseWriter, r *http.Request) {
	state := demoGameState{
		Phase: "pre-flop",
		CommunityCards: []demoCard{
			{Rank: "A", Suit: "spades"},
			{Rank: "K", Suit: "hearts"},
			{Rank: "7", Suit: "diamonds"},
			{Rank: "2", Suit: "clubs"},
			{Rank: "J", Suit: "spades"},
		},
		Seats: []demoSeat{
			{
				SeatIndex: 0, PlayerID: 1, Username: "You", ChipCount: 1500,
				HoleCards:     []demoCard{{Rank: "A", Suit: "hearts"}, {Rank: "A", Suit: "clubs"}},
				IsCurrentUser: true, IsActive: true, IsDealer: false,
			},
			{
				SeatIndex: 1, PlayerID: 2, Username: "Alice", ChipCount: 2000,
				HoleCards: []demoCard{{Rank: "K", Suit: "spades"}, {Rank: "Q", Suit: "hearts"}},
				IsDealer:  true,
			},
			{
				SeatIndex: 2, PlayerID: 3, Username: "Bob", ChipCount: 800,
				HoleCards: []demoCard{{Rank: "9", Suit: "clubs"}, {Rank: "8", Suit: "diamonds"}},
			},
			{
				SeatIndex: 3, PlayerID: 4, Username: "Carol", ChipCount: 3200,
				HoleCards: []demoCard{{Rank: "5", Suit: "hearts"}, {Rank: "5", Suit: "spades"}},
			},
		},
		Pot:           450,
		CurrentBet:    50,
		CurrentUserID: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}
