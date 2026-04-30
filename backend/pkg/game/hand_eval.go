package game

import (
	"sort"
)

// HandRank is the strength category of a poker hand (higher = stronger).
type HandRank int

const (
	HighCard      HandRank = 0
	OnePair       HandRank = 1
	TwoPair       HandRank = 2
	ThreeOfAKind  HandRank = 3
	Straight      HandRank = 4
	Flush         HandRank = 5
	FullHouse     HandRank = 6
	FourOfAKind   HandRank = 7
	StraightFlush HandRank = 8
)

// EvalResult is the result of evaluating a poker hand.
type EvalResult struct {
	Rank        HandRank
	Tiebreakers []int  // descending list of values for tiebreaking within same rank
	Description string // human-readable (e.g. "Full House")
	BestCards   []Card // the 5 best cards
}

// compareEval returns positive if a > b, negative if a < b, 0 if equal.
func compareEval(a, b EvalResult) int {
	if a.Rank != b.Rank {
		return int(a.Rank) - int(b.Rank)
	}
	for i := 0; i < len(a.Tiebreakers) && i < len(b.Tiebreakers); i++ {
		if a.Tiebreakers[i] != b.Tiebreakers[i] {
			return a.Tiebreakers[i] - b.Tiebreakers[i]
		}
	}
	return 0
}

// Evaluate returns the best 5-card hand from the given cards (accepts 5–7).
func Evaluate(cards []Card) EvalResult {
	if len(cards) == 5 {
		return eval5(cards)
	}

	var best EvalResult
	first := true
	combinations(cards, 5, func(combo []Card) {
		r := eval5(combo)
		if first || compareEval(r, best) > 0 {
			best = r
			first = false
		}
	})
	return best
}

// combinations calls fn for every k-element combination from src.
func combinations(src []Card, k int, fn func([]Card)) {
	n := len(src)
	indices := make([]int, k)
	for i := range indices {
		indices[i] = i
	}
	for {
		combo := make([]Card, k)
		for i, idx := range indices {
			combo[i] = src[idx]
		}
		fn(combo)

		i := k - 1
		for i >= 0 && indices[i] == i+n-k {
			i--
		}
		if i < 0 {
			break
		}
		indices[i]++
		for j := i + 1; j < k; j++ {
			indices[j] = indices[j-1] + 1
		}
	}
}

// eval5 evaluates exactly 5 cards and returns the hand rank.
func eval5(cards []Card) EvalResult {
	vals := make([]int, 5)
	for i, c := range cards {
		vals[i] = rankValue(c.Rank)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))

	isFlush := samesuit(cards)
	isStraight, straightHigh := checkStraight(vals)

	if isFlush && isStraight {
		desc := "Straight Flush"
		if straightHigh == 14 {
			desc = "Royal Flush"
		}
		return EvalResult{StraightFlush, []int{straightHigh}, desc, sortCards(cards)}
	}

	groups := groupByRank(vals)
	sorted := sortedGroups(groups)

	if sorted[0].count == 4 {
		kickers := kickers(vals, sorted[0].value, 1)
		return EvalResult{FourOfAKind, append([]int{sorted[0].value}, kickers...), "Four of a Kind", sortCards(cards)}
	}
	if sorted[0].count == 3 && sorted[1].count == 2 {
		return EvalResult{FullHouse, []int{sorted[0].value, sorted[1].value}, "Full House", sortCards(cards)}
	}
	if isFlush {
		return EvalResult{Flush, vals, "Flush", sortCards(cards)}
	}
	if isStraight {
		return EvalResult{Straight, []int{straightHigh}, "Straight", sortCards(cards)}
	}
	if sorted[0].count == 3 {
		kickers := kickers(vals, sorted[0].value, 2)
		return EvalResult{ThreeOfAKind, append([]int{sorted[0].value}, kickers...), "Three of a Kind", sortCards(cards)}
	}
	if sorted[0].count == 2 && sorted[1].count == 2 {
		high, low := sorted[0].value, sorted[1].value
		if high < low {
			high, low = low, high
		}
		kickers := kickers(vals, -1, 1, high, low)
		return EvalResult{TwoPair, append([]int{high, low}, kickers...), "Two Pair", sortCards(cards)}
	}
	if sorted[0].count == 2 {
		kickers := kickers(vals, sorted[0].value, 3)
		return EvalResult{OnePair, append([]int{sorted[0].value}, kickers...), "One Pair", sortCards(cards)}
	}
	return EvalResult{HighCard, vals, "High Card", sortCards(cards)}
}

type group struct {
	value int
	count int
}

func groupByRank(vals []int) map[int]int {
	m := make(map[int]int)
	for _, v := range vals {
		m[v]++
	}
	return m
}

func sortedGroups(groups map[int]int) []group {
	result := make([]group, 0, len(groups))
	for v, c := range groups {
		result = append(result, group{v, c})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].count != result[j].count {
			return result[i].count > result[j].count
		}
		return result[i].value > result[j].value
	})
	return result
}

func samesuit(cards []Card) bool {
	s := cards[0].Suit
	for _, c := range cards[1:] {
		if c.Suit != s {
			return false
		}
	}
	return true
}

func checkStraight(vals []int) (bool, int) {
	unique := uniqueSorted(vals)
	if len(unique) < 5 {
		return false, 0
	}
	// Check regular straight
	if unique[0]-unique[4] == 4 {
		return true, unique[0]
	}
	// Check wheel: A-2-3-4-5
	if unique[0] == 14 && unique[1] == 5 && unique[2] == 4 && unique[3] == 3 && unique[4] == 2 {
		return true, 5
	}
	return false, 0
}

func uniqueSorted(vals []int) []int {
	seen := make(map[int]bool)
	var result []int
	for _, v := range vals {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result
}

// kickers returns the top n values from vals that are not in the exclude set.
// Pass -1 as first exclude to use the variadic excludes only.
func kickers(vals []int, exclude int, n int, alsoExclude ...int) []int {
	excludeSet := make(map[int]bool)
	if exclude >= 0 {
		excludeSet[exclude] = true
	}
	for _, e := range alsoExclude {
		excludeSet[e] = true
	}
	var result []int
	for _, v := range vals {
		if !excludeSet[v] {
			result = append(result, v)
		}
		if len(result) == n {
			break
		}
	}
	return result
}

func sortCards(cards []Card) []Card {
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return rankValue(sorted[i].Rank) > rankValue(sorted[j].Rank)
	})
	return sorted
}
