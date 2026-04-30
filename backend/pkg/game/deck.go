package game

import (
	"math/rand"
	"time"
)

// Suit represents a card suit.
type Suit string

const (
	SuitHearts   Suit = "hearts"
	SuitDiamonds Suit = "diamonds"
	SuitClubs    Suit = "clubs"
	SuitSpades   Suit = "spades"
)

// Rank represents a card rank.
type Rank string

const (
	Rank2 Rank = "2"
	Rank3 Rank = "3"
	Rank4 Rank = "4"
	Rank5 Rank = "5"
	Rank6 Rank = "6"
	Rank7 Rank = "7"
	Rank8 Rank = "8"
	Rank9 Rank = "9"
	RankT Rank = "T"
	RankJ Rank = "J"
	RankQ Rank = "Q"
	RankK Rank = "K"
	RankA Rank = "A"
)

// Card is a playing card.
type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

var allRanks = []Rank{Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, RankT, RankJ, RankQ, RankK, RankA}
var allSuits = []Suit{SuitHearts, SuitDiamonds, SuitClubs, SuitSpades}

// rankValue returns the numeric value of a rank (2=2, T=10, A=14).
func rankValue(r Rank) int {
	switch r {
	case Rank2:
		return 2
	case Rank3:
		return 3
	case Rank4:
		return 4
	case Rank5:
		return 5
	case Rank6:
		return 6
	case Rank7:
		return 7
	case Rank8:
		return 8
	case Rank9:
		return 9
	case RankT:
		return 10
	case RankJ:
		return 11
	case RankQ:
		return 12
	case RankK:
		return 13
	case RankA:
		return 14
	}
	return 0
}

// Deck is a shuffleable collection of cards.
type Deck struct {
	cards []Card
	rng   *rand.Rand
}

// NewDeck returns a freshly ordered 52-card deck with its own RNG seeded from
// the current time.
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for _, suit := range allSuits {
		for _, rank := range allRanks {
			cards = append(cards, Card{Rank: rank, Suit: suit})
		}
	}
	return &Deck{
		cards: cards,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Shuffle randomises the deck in-place using Fisher-Yates.
func (d *Deck) Shuffle() {
	d.rng.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// Deal removes and returns the top n cards from the deck.
func (d *Deck) Deal(n int) []Card {
	if n > len(d.cards) {
		n = len(d.cards)
	}
	dealt := make([]Card, n)
	copy(dealt, d.cards[:n])
	d.cards = d.cards[n:]
	return dealt
}

// Remaining returns the number of cards left in the deck.
func (d *Deck) Remaining() int {
	return len(d.cards)
}
