package game

import (
	"errors"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
)

// Phase is the current stage of a poker hand.
type Phase string

const (
	PhasePreFlop  Phase = "pre-flop"
	PhaseFlop     Phase = "flop"
	PhaseTurn     Phase = "turn"
	PhaseRiver    Phase = "river"
	PhaseShowdown Phase = "showdown"
)

const (
	startingChips = 1000
	smallBlind    = 1
	bigBlind      = 2
)

var (
	ErrNotYourTurn  = errors.New("it is not your turn")
	ErrGameNotActive = errors.New("no active game")
	ErrInvalidAmount = errors.New("invalid raise amount")
	ErrInsufficientChips = errors.New("insufficient chips")
)

// Action types sent by the client.
const (
	ActionCheck = "CHECK"
	ActionCall  = "CALL"
	ActionRaise = "RAISE"
	ActionFold  = "FOLD"
)

// PlayerState tracks one player's in-hand state.
type PlayerState struct {
	UserID     uint
	Username   string
	Chips      int
	HoleCards  []Card
	SeatIndex  int
	IsDealer   bool
	Folded     bool
	AllIn      bool
	RoundBet   int  // amount bet in the current betting round
	HasActed   bool // true once the player has made a voluntary action this round
}

// GameEvent is an event produced by a game action, to be translated and broadcast by the hub.
type GameEvent struct {
	Type         string
	Personalized bool           // if true each player gets their own payload from Payloads
	Payload      any            // used when Personalized == false
	Payloads     map[uint]any   // keyed by UserID when Personalized == true
}

// Game is an in-progress Texas Hold'em hand.
type Game struct {
	Phase          Phase
	Players        []PlayerState // seat order; never reordered during hand
	CommunityCards []Card
	Deck           *Deck
	Pot            int
	CurrentBet     int // highest total bet any player has made this round
	ActiveIndex    int // index into Players of the player whose turn it is
	DealerIndex    int
	RoomCode       string
}

// NewGame creates a new Game from lobby players. Players are seated in the
// order provided by the lobby.
func NewGame(players []room.Player, roomCode string) *Game {
	states := make([]PlayerState, len(players))
	for i, p := range players {
		states[i] = PlayerState{
			UserID:    p.ID,
			Username:  p.Username,
			Chips:     startingChips,
			SeatIndex: i,
		}
	}
	return &Game{
		Phase:    PhasePreFlop,
		Players:  states,
		Deck:     NewDeck(),
		RoomCode: roomCode,
	}
}

// Start shuffles, assigns the dealer/blinds, deals hole cards, and returns the
// initial events to send to clients.
func (g *Game) Start() []GameEvent {
	g.Deck.Shuffle()

	n := len(g.Players)
	g.DealerIndex = 0
	g.Players[g.DealerIndex].IsDealer = true

	// Post blinds
	sbIndex := g.nextActive(g.DealerIndex)
	bbIndex := g.nextActive(sbIndex)

	g.postBlind(sbIndex, smallBlind)
	g.postBlind(bbIndex, bigBlind)
	g.CurrentBet = bigBlind

	// Deal 2 hole cards to each player
	for i := range g.Players {
		g.Players[i].HoleCards = g.Deck.Deal(2)
	}

	// First to act pre-flop: player after BB
	if n == 2 {
		// Heads-up: dealer/SB acts first pre-flop
		g.ActiveIndex = sbIndex
	} else {
		g.ActiveIndex = g.nextActive(bbIndex)
	}

	// Build personalized GAME_STARTED payloads
	startedPayloads := make(map[uint]any, n)
	for _, p := range g.Players {
		startedPayloads[p.UserID] = GameStartedPayload{
			TableID:        g.RoomCode,
			Seats:          g.buildSeatsFor(p.UserID),
			Phase:          string(g.Phase),
			Pot:            g.Pot,
			CurrentBet:     g.CurrentBet,
			ActivePlayerID: g.Players[g.ActiveIndex].UserID,
			CurrentUserID:  p.UserID,
		}
	}

	dealtPayloads := make(map[uint]any, n)
	for _, p := range g.Players {
		dealtPayloads[p.UserID] = CardsDealtPayload{
			Seats:     g.buildSeatsFor(p.UserID),
			HoleCards: p.HoleCards,
		}
	}

	return []GameEvent{
		{Type: "GAME_STARTED", Personalized: true, Payloads: startedPayloads},
		{Type: "CARDS_DEALT", Personalized: true, Payloads: dealtPayloads},
	}
}

// Act applies a player action and returns the resulting events.
func (g *Game) Act(userID uint, action string, amount int) ([]GameEvent, error) {
	if g.Players[g.ActiveIndex].UserID != userID {
		return nil, ErrNotYourTurn
	}

	idx := g.ActiveIndex
	player := &g.Players[idx]

	switch action {
	case ActionFold:
		player.Folded = true
		player.HasActed = true

	case ActionCheck:
		if g.CurrentBet > player.RoundBet {
			return nil, errors.New("cannot check — there is a bet to call")
		}
		player.HasActed = true

	case ActionCall:
		toCall := g.CurrentBet - player.RoundBet
		if toCall <= 0 {
			player.HasActed = true
			break
		}
		if toCall >= player.Chips {
			// All-in
			g.Pot += player.Chips
			player.RoundBet += player.Chips
			player.Chips = 0
			player.AllIn = true
		} else {
			player.Chips -= toCall
			player.RoundBet += toCall
			g.Pot += toCall
		}
		player.HasActed = true

	case ActionRaise:
		minRaise := g.CurrentBet + bigBlind
		if amount < minRaise {
			return nil, ErrInvalidAmount
		}
		if amount > player.Chips+player.RoundBet {
			return nil, ErrInsufficientChips
		}
		additional := amount - player.RoundBet
		player.Chips -= additional
		g.Pot += additional
		player.RoundBet = amount
		g.CurrentBet = amount
		player.HasActed = true
		// Reopen action: all other active players must act again
		for i := range g.Players {
			if i != idx && !g.Players[i].Folded && !g.Players[i].AllIn {
				g.Players[i].HasActed = false
			}
		}

	default:
		return nil, errors.New("unknown action: " + action)
	}

	actionEvent := GameEvent{
		Type: "PLAYER_ACTION",
		Payload: PlayerActionPayload{
			PlayerID:       userID,
			Action:         action,
			Amount:         amount,
			Pot:            g.Pot,
			CurrentBet:     g.CurrentBet,
			Seats:          g.buildSeatsFor(0), // public view: no hole cards
			ActivePlayerID: g.Players[g.ActiveIndex].UserID,
		},
	}

	events := []GameEvent{actionEvent}

	// Check if only one player remains (others folded)
	if g.activeCount() == 1 {
		winnerEvents := g.awardPotToLastStanding()
		events = append(events, winnerEvents...)
		return events, nil
	}

	// Check if the betting round is complete
	if g.roundComplete() {
		phaseEvents := g.advancePhase()
		events = append(events, phaseEvents...)
		return events, nil
	}

	// Advance to the next active player
	g.ActiveIndex = g.nextActive(g.ActiveIndex)
	// Update PLAYER_ACTION with correct next active player
	events[0].Payload = PlayerActionPayload{
		PlayerID:       userID,
		Action:         action,
		Amount:         amount,
		Pot:            g.Pot,
		CurrentBet:     g.CurrentBet,
		Seats:          g.buildSeatsFor(0),
		ActivePlayerID: g.Players[g.ActiveIndex].UserID,
	}

	return events, nil
}

// advancePhase deals community cards and moves to the next phase.
func (g *Game) advancePhase() []GameEvent {
	// Reset round bets
	for i := range g.Players {
		g.Players[i].RoundBet = 0
		g.Players[i].HasActed = false
	}
	g.CurrentBet = 0

	switch g.Phase {
	case PhasePreFlop:
		g.Phase = PhaseFlop
		g.CommunityCards = append(g.CommunityCards, g.Deck.Deal(3)...)
	case PhaseFlop:
		g.Phase = PhaseTurn
		g.CommunityCards = append(g.CommunityCards, g.Deck.Deal(1)...)
	case PhaseTurn:
		g.Phase = PhaseRiver
		g.CommunityCards = append(g.CommunityCards, g.Deck.Deal(1)...)
	case PhaseRiver:
		g.Phase = PhaseShowdown
		return g.evaluateShowdown()
	}

	// Post-flop action starts left of dealer
	g.ActiveIndex = g.nextActive(g.DealerIndex)

	return []GameEvent{{
		Type: "PHASE_CHANGE",
		Payload: PhaseChangePayload{
			Phase:          string(g.Phase),
			CommunityCards: g.CommunityCards,
			Pot:            g.Pot,
			CurrentBet:     g.CurrentBet,
			ActivePlayerID: g.Players[g.ActiveIndex].UserID,
		},
	}}
}

// evaluateShowdown determines the winner(s) at showdown.
// Tied hands result in a split pot; any odd chip goes to the earliest seat.
func (g *Game) evaluateShowdown() []GameEvent {
	var bestResult EvalResult
	var tiedIndices []int

	for i, p := range g.Players {
		if p.Folded || len(p.HoleCards) == 0 {
			continue
		}
		all := append(p.HoleCards, g.CommunityCards...)
		result := Evaluate(all)
		cmp := compareEval(result, bestResult)
		if len(tiedIndices) == 0 || cmp > 0 {
			bestResult = result
			tiedIndices = []int{i}
		} else if cmp == 0 {
			tiedIndices = append(tiedIndices, i)
		}
	}

	share := g.Pot / len(tiedIndices)
	remainder := g.Pot % len(tiedIndices)
	winnerIDs := make([]uint, len(tiedIndices))
	for j, idx := range tiedIndices {
		extra := 0
		if j == 0 {
			extra = remainder
		}
		g.Players[idx].Chips += share + extra
		winnerIDs[j] = g.Players[idx].UserID
	}

	return []GameEvent{{
		Type: "GAME_OVER",
		Payload: GameOverPayload{
			WinnerID:  winnerIDs[0],
			WinnerIDs: winnerIDs,
			Pot:       g.Pot,
			Seats:     g.buildAllSeats(),
		},
	}}
}

// awardPotToLastStanding gives the pot to the sole non-folded player.
func (g *Game) awardPotToLastStanding() []GameEvent {
	for i, p := range g.Players {
		if !p.Folded {
			g.Players[i].Chips += g.Pot
			g.Phase = PhaseShowdown

			return []GameEvent{{
				Type: "GAME_OVER",
				Payload: GameOverPayload{
					WinnerID:  p.UserID,
					WinnerIDs: []uint{p.UserID},
					Pot:       g.Pot,
					Seats:     g.buildAllSeats(),
				},
			}}
		}
	}
	return nil
}

// roundComplete returns true when all non-folded, non-all-in players have
// acted and all bets are equal.
func (g *Game) roundComplete() bool {
	for _, p := range g.Players {
		if p.Folded || p.AllIn {
			continue
		}
		if !p.HasActed || p.RoundBet != g.CurrentBet {
			return false
		}
	}
	return true
}

// activeCount returns the number of players still in the hand.
func (g *Game) activeCount() int {
	n := 0
	for _, p := range g.Players {
		if !p.Folded {
			n++
		}
	}
	return n
}

// nextActive returns the index of the next non-folded, non-all-in player after from.
func (g *Game) nextActive(from int) int {
	n := len(g.Players)
	for i := 1; i <= n; i++ {
		idx := (from + i) % n
		if !g.Players[idx].Folded && !g.Players[idx].AllIn {
			return idx
		}
	}
	return from
}

// postBlind posts a forced blind bet without marking HasActed.
func (g *Game) postBlind(idx int, amount int) {
	p := &g.Players[idx]
	if amount >= p.Chips {
		g.Pot += p.Chips
		p.RoundBet += p.Chips
		p.Chips = 0
		p.AllIn = true
	} else {
		p.Chips -= amount
		p.RoundBet += amount
		g.Pot += amount
	}
}

// buildSeatsFor builds the seat list as seen by viewerID (only their hole
// cards are populated; pass 0 to return all seats with no hole cards).
func (g *Game) buildSeatsFor(viewerID uint) []SeatPayload {
	seats := make([]SeatPayload, len(g.Players))
	for i, p := range g.Players {
		var holeCards []Card
		if viewerID != 0 && p.UserID == viewerID {
			holeCards = p.HoleCards
		}
		seats[i] = SeatPayload{
			SeatIndex:     p.SeatIndex,
			PlayerID:      p.UserID,
			Username:      p.Username,
			ChipCount:     p.Chips,
			HoleCards:     holeCards,
			IsCurrentUser: p.UserID == viewerID,
			IsActive:      g.Players[g.ActiveIndex].UserID == p.UserID,
			IsDealer:      p.IsDealer,
		}
	}
	return seats
}

// buildAllSeats returns all seats with hole cards revealed (for showdown).
func (g *Game) buildAllSeats() []SeatPayload {
	seats := make([]SeatPayload, len(g.Players))
	for i, p := range g.Players {
		seats[i] = SeatPayload{
			SeatIndex: p.SeatIndex,
			PlayerID:  p.UserID,
			Username:  p.Username,
			ChipCount: p.Chips,
			HoleCards: p.HoleCards,
			IsDealer:  p.IsDealer,
		}
	}
	return seats
}

// ActivePlayerID returns the user ID of the player whose turn it is.
func (g *Game) ActivePlayerID() uint {
	if len(g.Players) == 0 {
		return 0
	}
	return g.Players[g.ActiveIndex].UserID
}

// ---- Payload types used by Game (protocol.go mirrors these for the wire) ----

// SeatPayload is the per-seat data sent over the wire to clients.
type SeatPayload struct {
	SeatIndex     int    `json:"seatIndex"`
	PlayerID      uint   `json:"playerId"`
	Username      string `json:"username"`
	ChipCount     int    `json:"chipCount"`
	HoleCards     []Card `json:"holeCards"`
	IsCurrentUser bool   `json:"isCurrentUser"`
	IsActive      bool   `json:"isActive"`
	IsDealer      bool   `json:"isDealer"`
}

// GameStartedPayload is sent individually to each player when a game begins.
type GameStartedPayload struct {
	TableID        string        `json:"tableId"`
	Seats          []SeatPayload `json:"seats"`
	Phase          string        `json:"phase"`
	Pot            int           `json:"pot"`
	CurrentBet     int           `json:"currentBet"`
	ActivePlayerID uint          `json:"activePlayerId"`
	CurrentUserID  uint          `json:"currentUserId"`
}

// CardsDealtPayload is sent individually to each player with their own hole cards.
type CardsDealtPayload struct {
	Seats     []SeatPayload `json:"seats"`
	HoleCards []Card        `json:"holeCards"`
}

// PlayerActionPayload is broadcast after any player action.
type PlayerActionPayload struct {
	PlayerID       uint          `json:"playerId"`
	Action         string        `json:"action"`
	Amount         int           `json:"amount,omitempty"`
	Pot            int           `json:"pot"`
	CurrentBet     int           `json:"currentBet"`
	Seats          []SeatPayload `json:"seats"`
	ActivePlayerID uint          `json:"activePlayerId"`
}

// PhaseChangePayload is broadcast when the hand advances to a new phase.
type PhaseChangePayload struct {
	Phase          string `json:"phase"`
	CommunityCards []Card `json:"communityCards"`
	Pot            int    `json:"pot"`
	CurrentBet     int    `json:"currentBet"`
	ActivePlayerID uint   `json:"activePlayerId"`
}

// GameOverPayload is broadcast when the hand ends.
// WinnerIDs contains all winners (len > 1 for a split pot).
// WinnerID is always WinnerIDs[0] for backward compatibility.
type GameOverPayload struct {
	WinnerID  uint          `json:"winnerId"`
	WinnerIDs []uint        `json:"winnerIds"`
	Pot       int           `json:"pot"`
	Seats     []SeatPayload `json:"seats,omitempty"`
}
