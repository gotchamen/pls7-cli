// Package poker provides the core data structures and rules for playing poker.
// It includes types for cards, decks, hand evaluation, and game rules, forming
// the foundational building blocks for a poker game engine.
package poker

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

// HandRank defines the ranking of a poker hand. The integer values are ordered
// from the lowest rank (HighCard) to the highest (RoyalFlush), which allows for
// direct comparison.
type HandRank int

// HandRank constants represent the possible poker hand rankings.
// These are ordered from weakest to strongest.
const (
	HighCard          HandRank = iota // HighCard represents the lowest-ranking hand, determined by the highest card.
	OnePair                           // OnePair consists of two cards of the same rank.
	TwoPair                           // TwoPair consists of two pairs of different ranks.
	ThreeOfAKind                      // ThreeOfAKind consists of three cards of the same rank.
	Straight                          // Straight consists of five cards of sequential rank.
	SkipStraight                      // SkipStraight is a special hand for PLS7, with ranks in a gapped sequence (e.g., A-J-9-7-5).
	Flush                             // Flush consists of five cards of the same suit.
	FullHouse                         // FullHouse consists of a ThreeOfAKind and a OnePair.
	FourOfAKind                       // FourOfAKind consists of four cards of the same rank.
	StraightFlush                     // StraightFlush consists of five cards of sequential rank and the same suit.
	SkipStraightFlush                 // SkipStraightFlush is a special hand for PLS7, a SkipStraight with all cards of the same suit.
	RoyalFlush                        // RoyalFlush is the highest-ranking hand, an Ace-high StraightFlush.
)

// String returns the string representation of a HandRank (e.g., "High Card", "Royal Flush").
// It implements the fmt.Stringer interface.
func (hr HandRank) String() string {
	return []string{
		"High Card",
		"One Pair",
		"Two Pair",
		"Three of a Kind",
		"Straight",
		"Skip Straight",
		"Flush",
		"Full House",
		"Four of a Kind",
		"Straight Flush",
		"Skip Straight Flush",
		"Royal Flush",
	}[hr]
}

// handRankFromString converts a string representation of a hand rank (e.g., "high_card")
// to its corresponding HandRank constant. It returns the HandRank and a boolean
// indicating if the conversion was successful.
func handRankFromString(s string) (HandRank, bool) {
	switch s {
	case "high_card":
		return HighCard, true
	case "one_pair":
		return OnePair, true
	case "two_pair":
		return TwoPair, true
	case "three_of_a_kind":
		return ThreeOfAKind, true
	case "straight":
		return Straight, true
	case "skip_straight":
		return SkipStraight, true
	case "flush":
		return Flush, true
	case "full_house":
		return FullHouse, true
	case "four_of_a_kind":
		return FourOfAKind, true
	case "straight_flush":
		return StraightFlush, true
	case "skip_straight_flush":
		return SkipStraightFlush, true
	case "royal_flush":
		return RoyalFlush, true
	default:
		return 0, false
	}
}

// HandResult stores the complete details of an evaluated poker hand, including its
// rank, the cards that form the hand, and values for tie-breaking.
type HandResult struct {
	Rank       HandRank // The rank of the hand (e.g., Flush, Straight).
	Cards      []Card   // The best 5 cards that form this hand.
	HighValues []Rank   // A sorted slice of ranks used for tie-breaking. For a pair, this would be [PairRank, Kicker1, Kicker2, Kicker3]. For a flush, it's the ranks of the 5 flush cards.
}

// String returns a detailed string representation of the HandResult,
// suitable for display (e.g., "Two Pair, Kings and Queens, K♠ K♦ Q♥ Q♣ A♣").
func (hr *HandResult) String() string {
	if hr == nil {
		return "N/A"
	}

	switch hr.Rank {
	case RoyalFlush, SkipStraightFlush, StraightFlush, Straight, SkipStraight, FullHouse, Flush, OnePair:
		return fmt.Sprintf("%s, %s", hr.Rank.String(), hr.CardsString())
	case FourOfAKind:
		quadRank := hr.HighValues[0].String()
		return fmt.Sprintf("%s Four of a Kind, %s", quadRank, hr.CardsString())
	case ThreeOfAKind:
		tripleRank := hr.HighValues[0].String()
		return fmt.Sprintf("%s Three of a Kind, %s", tripleRank, hr.CardsString())
	case TwoPair:
		highPair := hr.HighValues[0].String()
		lowPair := hr.HighValues[1].String()
		return fmt.Sprintf("Two Pair, %ss and %ss, %s", highPair, lowPair, hr.CardsString())
	case HighCard:
		topCard := hr.HighValues[0].String()
		return fmt.Sprintf("%s-High, %s", topCard, hr.CardsString())
	default:
		return "Unknown Hand"
	}
}

// CardsString returns a string representation of just the cards in the hand result.
func (hr *HandResult) CardsString() string {
	if hr == nil || len(hr.Cards) == 0 {
		return "No Cards"
	}
	cards := make([]string, len(hr.Cards))
	for i, c := range hr.Cards {
		cards[i] = c.Rank.String()
		cards[i] += c.Suit.String() + " "
	}
	return JoinStrings(cards)
}

// handAnalysis is a private helper struct used during hand evaluation. It stores
// pre-calculated information about a pool of cards, such as rank and suit counts,
// to avoid redundant calculations.
type handAnalysis struct {
	rankCounts map[Rank]int // Maps each rank to its frequency.
	suitCounts map[Suit]int // Maps each suit to its frequency.
	cards      []Card       // The original pool of cards, sorted by rank in descending order.
}

// String provides a string representation of the handAnalysis for debugging purposes.
func (ha *handAnalysis) String() string {
	if ha == nil {
		return "N/A"
	}
	rankStr := "Rank Counts: "
	for rank, count := range ha.rankCounts {
		rankStr += fmt.Sprintf("%s(%d) ", rank.String(), count)
	}
	suitStr := "Suit Counts: "
	for suit, count := range ha.suitCounts {
		suitStr += fmt.Sprintf("%s (%d) ", suit.String(), count)
	}
	return fmt.Sprintf("%s, %s, Cards: %v", rankStr, suitStr, ha.cards)
}

// newHandAnalysis creates and populates a handAnalysis struct from a given pool of cards.
// It sorts the cards by rank descending and calculates rank/suit frequencies.
func newHandAnalysis(pool []Card) *handAnalysis {
	analysis := &handAnalysis{
		rankCounts: make(map[Rank]int),
		suitCounts: make(map[Suit]int),
		cards:      make([]Card, len(pool)),
	}
	copy(analysis.cards, pool)

	// Sort cards by rank descending for consistent processing.
	sort.Slice(analysis.cards, func(i, j int) bool {
		return analysis.cards[i].Rank > analysis.cards[j].Rank
	})

	for _, c := range analysis.cards {
		analysis.rankCounts[c.Rank]++
		analysis.suitCounts[c.Suit]++
	}
	return analysis
}

// getHandIterator selects the appropriate hand combination strategy based on the game rules.
func getHandIterator(rules *GameRules) HandIterator {
	switch rules.HoleCards.UseConstraint {
	case "exact":
		return &ExactCombinationGenerator{}
	default:
		// Default to "any" for safety and backward compatibility.
		if rules.HoleCards.UseConstraint != "any" && rules.HoleCards.UseConstraint != "" {
			logrus.Warnf("Unknown UseConstraint '%s', defaulting to 'any'", rules.HoleCards.UseConstraint)
		}
		return &AnyCombinationGenerator{}
	}
}

// EvaluateHand is the main evaluation function. It takes a player's hole cards and the
// community cards and, based on the provided game rules, determines the best possible
// high hand and, if applicable, the best possible low hand.
//
// The evaluation process is as follows:
//
// 1. High Hand Evaluation:
//   - The function first combines the hole cards and community cards into a single pool.
//   - It then iterates through a list of hand ranks, ordered from highest to lowest
//     (e.g., Royal Flush, then Straight Flush, etc.), which is determined by the game rules.
//   - For each rank, it calls a specific `find...` helper function (e.g., `findStraightFlush`)
//     to check if that hand can be made from the pool of cards.
//   - If a hand is found, it is compared to the best hand found so far (`highResult`).
//     If the new hand is better, it replaces `highResult`. This process continues through
//     all hand ranks to ensure the absolute best hand is found.
//
// 2. Low Hand Evaluation (only for Hi-Lo games):
//   - If the game rules enable low hands, it calls `findBestLowHand`.
//   - This function attempts to find the best qualifying low hand (e.g., 8-low or better)
//     from the card pool, independent of the high hand result.
//
// Parameters:
//   - holeCards: The player's private cards.
//   - communityCards: The shared cards on the board.
//   - gameRules: The ruleset defining which hands are valid and their rankings.
//
// Returns:
//   - highResult: A HandResult for the best high hand, or nil if no hand could be formed.
//   - lowResult: A HandResult for the best low hand (if enabled by rules), or nil.
func EvaluateHand(holeCards []Card, communityCards []Card, gameRules *GameRules) (highResult *HandResult, lowResult *HandResult) {
	// 1. Select the combination generation strategy based on the game rules.
	iterator := getHandIterator(gameRules)

	// 2. Generate all possible 5-card hand combinations using the selected strategy.
	all5CardCombos := iterator.Generate(holeCards, communityCards, gameRules)

	if all5CardCombos == nil {
		logrus.Warnf("EvaluateHand: No card combinations could be generated with the given hole and community cards.")
		return nil, nil // No valid high hand could be formed, and by extension no low hand.
	}

	// 3. Evaluate each 5-card combination to find the best high hand.
	var bestHand *HandResult
	for _, combo := range all5CardCombos {
		handResult := evaluateSingleHand(combo, gameRules)
		if handResult != nil {
			if bestHand == nil || compareHandResults(handResult, bestHand) > 0 {
				bestHand = handResult
			}
		}
	}
	highResult = bestHand

	// 4. From the same combinations, find the best low hand if the game rules enable it.
	if gameRules.LowHand.Enabled {
		var bestLowHand *HandResult
		for _, combo := range all5CardCombos {
			if isQualifyingLowHand(combo, Rank(gameRules.LowHand.MaxRank)) {
				// This combo is a valid low hand. We create a HandResult for it
				// so we can use the standard comparison logic.
				currentLowHand := &HandResult{
					Rank:       HighCard, // Low hands are ranked as HighCard for comparison.
					Cards:      combo,
					HighValues: getLowHandHighValues(combo),
				}

				if bestLowHand == nil || compareLowHands(currentLowHand, bestLowHand) > 0 {
					bestLowHand = currentLowHand
				}
			}
		}
		lowResult = bestLowHand
	}

	return highResult, lowResult
}

// isQualifyingLowHand checks if a 5-card hand meets the criteria for a low hand.
func isQualifyingLowHand(cards []Card, maxRank Rank) bool {
	if len(cards) != 5 {
		return false
	}
	usedRanks := make(map[Rank]bool)
	for _, card := range cards {
		if card.Rank > maxRank && card.Rank != Ace {
			return false // A card is too high.
		}
		if usedRanks[card.Rank] {
			return false // Contains a pair, not a valid low hand.
		}
		usedRanks[card.Rank] = true
	}
	return true
}

// compareLowHands compares two low hands. It returns 1 if h1 is better (lower) than h2,
// -1 if h2 is better, and 0 if they are identical.
func compareLowHands(h1, h2 *HandResult) int {
	for i := 0; i < len(h1.HighValues); i++ {
		v1 := getLowRankValue(h1.HighValues[i])
		v2 := getLowRankValue(h2.HighValues[i])
		if v1 < v2 {
			return 1 // h1 is better because its card is lower.
		}
		if v1 > v2 {
			return -1 // h2 is better.
		}
	}
	return 0 // Hands are identical.
}

// getLowHandHighValues returns the ranks of the cards sorted for low-hand comparison (highest to lowest).
func getLowHandHighValues(cards []Card) []Rank {
	sortedCards := make([]Card, 5)
	copy(sortedCards, cards)
	// Sort descending by low-rank value (Ace=1, Two=2, etc.)
	sort.Slice(sortedCards, func(i, j int) bool {
		return getLowRankValue(sortedCards[i].Rank) > getLowRankValue(sortedCards[j].Rank)
	})
	return []Rank{
		sortedCards[0].Rank,
		sortedCards[1].Rank,
		sortedCards[2].Rank,
		sortedCards[3].Rank,
		sortedCards[4].Rank,
	}
}

// handChecker is a function that checks if the given hand analysis matches a specific hand rank.
type handChecker func(analysis *handAnalysis) *HandResult

// handCheckers maps each HandRank to its checker function.
var handCheckers = map[HandRank]handChecker{
	RoyalFlush:        checkRoyalFlush,
	SkipStraightFlush: checkSkipStraightFlush,
	StraightFlush:     checkStraightFlush,
	FourOfAKind:       checkFourOfAKind,
	FullHouse:         checkFullHouse,
	Flush:             checkFlush,
	SkipStraight:      checkSkipStraight,
	Straight:          checkStraight,
	ThreeOfAKind:      checkThreeOfAKind,
	TwoPair:           checkTwoPair,
	OnePair:           checkOnePair,
	HighCard:          checkHighCard,
}

// evaluateSingleHand takes exactly 5 cards and determines their rank.
func evaluateSingleHand(cards []Card, gameRules *GameRules) *HandResult {
	if len(cards) != 5 {
		logrus.Warnf("evaluateSingleHand called with %d cards, but expected 5", len(cards))
		return nil
	}

	analysis := newHandAnalysis(cards)
	handRankOrder := getHandRanks(&gameRules.HandRankings)

	for _, rank := range handRankOrder {
		if checker, ok := handCheckers[rank]; ok {
			if result := checker(analysis); result != nil {
				return result
			}
		}
	}
	return nil
}

func checkRoyalFlush(analysis *handAnalysis) *HandResult {
	if sfCards, ok := findStraightFlush(analysis); ok && sfCards[0].Rank == Ace {
		return &HandResult{Rank: RoyalFlush, Cards: sfCards, HighValues: []Rank{sfCards[0].Rank}}
	}
	return nil
}

func checkSkipStraightFlush(analysis *handAnalysis) *HandResult {
	if ssfCards, ok := findSkipStraightFlush(analysis); ok {
		return &HandResult{Rank: SkipStraightFlush, Cards: ssfCards, HighValues: []Rank{ssfCards[0].Rank}}
	}
	return nil
}

func checkStraightFlush(analysis *handAnalysis) *HandResult {
	if sfCards, ok := findStraightFlush(analysis); ok {
		return &HandResult{Rank: StraightFlush, Cards: sfCards, HighValues: []Rank{sfCards[0].Rank}}
	}
	return nil
}

func checkFourOfAKind(analysis *handAnalysis) *HandResult {
	if quadRank, ok := findBestNOfAKind(analysis.rankCounts, 4); ok {
		if found, kickers := findKickers(analysis.cards, []Rank{quadRank}, 1); found {
			quadCards := findCardsByRank(analysis.cards, quadRank, 4)
			return &HandResult{Rank: FourOfAKind, Cards: append(quadCards, kickers...), HighValues: []Rank{quadRank, kickers[0].Rank}}
		}
	}
	return nil
}

func checkFullHouse(analysis *handAnalysis) *HandResult {
	if tripleRank, pairRank, ok := findBestFullHouse(analysis.rankCounts); ok {
		tripleCards := findCardsByRank(analysis.cards, tripleRank, 3)
		pairCards := findCardsByRank(analysis.cards, pairRank, 2)
		return &HandResult{Rank: FullHouse, Cards: append(tripleCards, pairCards...), HighValues: []Rank{tripleRank, pairRank}}
	}
	return nil
}

func checkFlush(analysis *handAnalysis) *HandResult {
	if flushCards, ok := findBestFlush(analysis); ok {
		return &HandResult{
			Rank:       Flush,
			Cards:      flushCards,
			HighValues: []Rank{flushCards[0].Rank, flushCards[1].Rank, flushCards[2].Rank, flushCards[3].Rank, flushCards[4].Rank},
		}
	}
	return nil
}

func checkSkipStraight(analysis *handAnalysis) *HandResult {
	if ssCards, ok := findSkipStraight(analysis); ok {
		return &HandResult{Rank: SkipStraight, Cards: ssCards, HighValues: []Rank{ssCards[0].Rank}}
	}
	return nil
}

func checkStraight(analysis *handAnalysis) *HandResult {
	if straightCards, ok := findBestStraight(analysis); ok {
		return &HandResult{Rank: Straight, Cards: straightCards, HighValues: []Rank{straightCards[0].Rank}}
	}
	return nil
}

func checkThreeOfAKind(analysis *handAnalysis) *HandResult {
	if tripleRank, ok := findBestNOfAKind(analysis.rankCounts, 3); ok {
		if found, kickers := findKickers(analysis.cards, []Rank{tripleRank}, 2); found {
			tripleCards := findCardsByRank(analysis.cards, tripleRank, 3)
			return &HandResult{
				Rank:       ThreeOfAKind,
				Cards:      append(tripleCards, kickers...),
				HighValues: []Rank{tripleRank, kickers[0].Rank, kickers[1].Rank},
			}
		}
	}
	return nil
}

func checkTwoPair(analysis *handAnalysis) *HandResult {
	if highPair, lowPair, ok := findBestTwoPair(analysis.rankCounts); ok {
		if found, kickers := findKickers(analysis.cards, []Rank{highPair, lowPair}, 1); found {
			highPairCards := findCardsByRank(analysis.cards, highPair, 2)
			lowPairCards := findCardsByRank(analysis.cards, lowPair, 2)
			allCards := append(highPairCards, lowPairCards...)
			allCards = append(allCards, kickers...)
			return &HandResult{Rank: TwoPair, Cards: allCards, HighValues: []Rank{highPair, lowPair, kickers[0].Rank}}
		}
	}
	return nil
}

func checkOnePair(analysis *handAnalysis) *HandResult {
	if pairRank, ok := findBestNOfAKind(analysis.rankCounts, 2); ok {
		if found, kickers := findKickers(analysis.cards, []Rank{pairRank}, 3); found {
			pairCards := findCardsByRank(analysis.cards, pairRank, 2)
			return &HandResult{
				Rank:       OnePair,
				Cards:      append(pairCards, kickers...),
				HighValues: []Rank{pairRank, kickers[0].Rank, kickers[1].Rank, kickers[2].Rank},
			}
		}
	}
	return nil
}

func checkHighCard(analysis *handAnalysis) *HandResult {
	return &HandResult{
		Rank:  HighCard,
		Cards: analysis.cards[:5],
		HighValues: []Rank{
			analysis.cards[0].Rank, analysis.cards[1].Rank, analysis.cards[2].Rank, analysis.cards[3].Rank, analysis.cards[4].Rank,
		},
	}
}

// findSkipStraightFlush checks for a Skip Straight Flush. It first identifies a
// potential flush and then checks if the flushed cards form a Skip Straight.
func findSkipStraightFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			// Extract all cards of the potential flush suit.
			flushCards := make([]Card, 0, count)
			for _, card := range analysis.cards {
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			// Analyze these flushed cards to see if they form a Skip Straight.
			flushAnalysis := newHandAnalysis(flushCards)
			if ssfCards, ok := findSkipStraight(flushAnalysis); ok {
				return ssfCards, true
			}
		}
	}
	return nil, false
}

// findBestLowHand identifies the best possible "N-or-better" low hand from the card pool.
// A low hand consists of five unique cards with ranks at or below `maxRank` (e.g., 8),
// with Aces counting as low. The "best" low hand is the one with the lowest high card
// (e.g., 7-5-4-3-2 is better than 8-4-3-2-A).
//
// The process is:
// 1. Collect all unique cards from the pool that are eligible for a low hand.
// 2. If there are fewer than 5 such cards, no low hand is possible.
// 3. Sort the eligible cards ascending (Ace is lowest) to find the best combination.
// 4. The 5 lowest cards form the best possible low hand.
func findBestLowHand(analysis *handAnalysis, maxRank Rank) (*HandResult, bool) {
	uniqueLowCards := make([]Card, 0, 8)
	usedRanks := make(map[Rank]bool)

	// Collect all unique cards that qualify for a low hand.
	for _, card := range analysis.cards {
		isLowCard := card.Rank <= maxRank || card.Rank == Ace
		if isLowCard && !usedRanks[card.Rank] {
			uniqueLowCards = append(uniqueLowCards, card)
			usedRanks[card.Rank] = true
		}
	}

	// A valid low hand requires at least 5 qualifying cards.
	if len(uniqueLowCards) < 5 {
		return nil, false
	}

	// Sort the qualifying cards by rank ascending (Ace is lowest) to find the best hand.
	sort.Slice(uniqueLowCards, func(i, j int) bool {
		return getLowRankValue(uniqueLowCards[i].Rank) < getLowRankValue(uniqueLowCards[j].Rank)
	})

	// The best 5-card low hand is the 5 lowest unique cards.
	bestLowCards := uniqueLowCards[:5]

	// Sort the final 5 cards descending for tie-breaking purposes, with Ace treated as low.
	sort.Slice(bestLowCards, func(i, j int) bool {
		return getLowRankValue(bestLowCards[i].Rank) > getLowRankValue(bestLowCards[j].Rank)
	})

	return &HandResult{
		Rank:  HighCard, // Low hands are ranked as HighCard but compared by their low values.
		Cards: bestLowCards,
		HighValues: []Rank{
			bestLowCards[0].Rank,
			bestLowCards[1].Rank,
			bestLowCards[2].Rank,
			bestLowCards[3].Rank,
			bestLowCards[4].Rank,
		},
	}, true
}

// getLowRankValue returns the numeric value of a rank for low hand comparisons,
// where Ace is treated as 1.
func getLowRankValue(r Rank) int {
	if r == Ace {
		return 1
	}
	return int(r)
}

// findStraightFlush checks for a Straight Flush. It first identifies a potential
// flush and then checks if the flushed cards form a regular Straight.
func findStraightFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			flushCards := make([]Card, 0, count)
			for _, card := range analysis.cards {
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			flushAnalysis := newHandAnalysis(flushCards)
			if sfCards, ok := findBestStraight(flushAnalysis); ok {
				return sfCards, true
			}
		}
	}
	return nil, false
}

// findSkipStraight checks for a Skip Straight. This is a special PLS7 hand
// with a specific gapped sequence of 5 cards (e.g., K-J-9-7-5).
func findSkipStraight(analysis *handAnalysis) ([]Card, bool) {
	logrus.Tracef("findSkipStraight: Analyzing handAnalysis: %+v", analysis)

	uniqueRanksAceHigh := make([]Rank, 0)
	seenRanks := make(map[Rank]bool)
	hasAce := false

	for _, card := range analysis.cards {
		if !seenRanks[card.Rank] {
			uniqueRanksAceHigh = append(uniqueRanksAceHigh, card.Rank)
			seenRanks[card.Rank] = true
			if card.Rank == Ace {
				hasAce = true
			}
		}
	}
	sort.Slice(uniqueRanksAceHigh, func(i, j int) bool { return uniqueRanksAceHigh[i] > uniqueRanksAceHigh[j] })
	listOfUniqueRanks := [][]Rank{uniqueRanksAceHigh}

	// If Ace is present, create a second list treating Ace as 1 (for low-end straights)
	if hasAce {
		logrus.Tracef("findSkipStraight: Ace found, creating alternative rank list treating Ace as 1.")
		uniqueRanksAceLow := make([]Rank, 0)
		uniqueRanksAceLow = append(uniqueRanksAceLow, uniqueRanksAceHigh[1:]...) // Copy all except Ace
		uniqueRanksAceLow = append(uniqueRanksAceLow, uniqueRanksAceHigh[0])     // Add Ace at the end
		listOfUniqueRanks = append(listOfUniqueRanks, uniqueRanksAceLow)
	}
	logrus.Tracef("findSkipStraight: listOfUniqueRanks: %+v", listOfUniqueRanks)

	for _, uniqueRanks := range listOfUniqueRanks {
		// In PLS7, a Skip Straight's highest card must be 9 or greater.
		if len(uniqueRanks) > 0 && uniqueRanks[0] < 9 {
			logrus.Tracef(
				"findSkipStraight: Skipping analysis for uniqueRanks starting with %v, as it is less than 9.",
				uniqueRanks[0],
			)
			continue // Skip analysis if the highest rank is less than 9
		}
		for i := 0; i <= len(uniqueRanks)-5; i++ {
			biggest := uniqueRanks[i]      // The biggest rank in a Skip Straight
			smallest := uniqueRanks[i] - 8 // The smallest rank in a Skip Straight is 8 ranks below the top rank
			// Only biggest is an odd number, smallest less than Two can be treated as Ace
			if smallest < Two && biggest%2 == 1 {
				smallest = Ace
				logrus.Tracef("findSkipStraight: Adjusting smallest rank to Ace as it is less than Two and biggest is an odd number.")
			}
			possibleSkipStraight := []Rank{
				uniqueRanks[i],
				uniqueRanks[i] - 2,
				uniqueRanks[i] - 4,
				uniqueRanks[i] - 6,
				smallest,
			}
			logrus.Tracef("findSkipStraight: Checking possible Skip Straight: %v", possibleSkipStraight)
			isSkipStraight := true
			for _, c := range possibleSkipStraight {
				if !containsRank(uniqueRanks, c) {
					isSkipStraight = false
					logrus.Tracef(
						"findSkipStraight: Rank %v not found in uniqueRanks: %v, setting isSkipStraight to false.",
						c, uniqueRanks,
					)
					break
				}
			}
			if isSkipStraight {
				topRank := uniqueRanks[i]
				logrus.Tracef("findSkipStraight: Found Skip Straight! topRank: %v, ranks: %v", topRank, possibleSkipStraight)
				return findCardsForStraight(analysis.cards, possibleSkipStraight), true
			}
		}
	}
	logrus.Tracef("findSkipStraight: No Skip Straight found.")
	return nil, false
}

// findBestFullHouse finds the best possible Full House (three of a kind and a pair).
// It looks for the highest-ranked three of a kind, then the highest-ranked pair
// from the remaining cards.
func findBestFullHouse(rankCounts map[Rank]int) (Rank, Rank, bool) {
	var bestTripleRank Rank = -1
	var bestPairRank Rank = -1

	// Find the highest rank with at least 3 cards.
	for rank, count := range rankCounts {
		if count >= 3 {
			if rank > bestTripleRank {
				bestTripleRank = rank
			}
		}
	}
	if bestTripleRank == -1 {
		return -1, -1, false
	}

	// Find the highest rank with at least 2 cards, excluding the triple.
	for rank, count := range rankCounts {
		if count >= 2 && rank != bestTripleRank {
			if rank > bestPairRank {
				bestPairRank = rank
			}
		}
	}
	if bestPairRank == -1 {
		return -1, -1, false
	}

	return bestTripleRank, bestPairRank, true
}

// findBestFlush finds the best possible Flush (five cards of the same suit).
// It returns the 5 highest-ranked cards of the most common suit.
func findBestFlush(analysis *handAnalysis) ([]Card, bool) {
	for suit, count := range analysis.suitCounts {
		if count >= 5 {
			flushCards := make([]Card, 0, count)
			// Since analysis.cards is pre-sorted high-to-low, the first cards
			// of the flush suit we find are the highest ones.
			for _, card := range analysis.cards {
				if card.Suit == suit {
					flushCards = append(flushCards, card)
				}
			}
			return flushCards[:5], true
		}
	}
	return nil, false
}

// findBestStraight finds the best possible Straight (five cards of sequential rank).
// It handles both standard straights and the A-2-3-4-5 "wheel" straight.
func findBestStraight(analysis *handAnalysis) ([]Card, bool) {
	uniqueRanks := make([]Rank, 0, len(analysis.rankCounts))
	for rank := range analysis.rankCounts {
		uniqueRanks = append(uniqueRanks, rank)
	}
	sort.Slice(uniqueRanks, func(i, j int) bool { return uniqueRanks[i] > uniqueRanks[j] })

	if len(uniqueRanks) < 5 {
		return nil, false
	}

	// Special case: Check for the A-2-3-4-5 "wheel" straight.
	if containsRank(uniqueRanks, Ace) &&
		containsRank(uniqueRanks, Five) &&
		containsRank(uniqueRanks, Four) &&
		containsRank(uniqueRanks, Three) &&
		containsRank(uniqueRanks, Two) {
		return findCardsForStraight(analysis.cards, []Rank{Five, Four, Three, Two, Ace}), true
	}

	// Check for other straights, starting from the highest rank.
	for i := 0; i <= len(uniqueRanks)-5; i++ {
		isStraight := true
		for j := 0; j < 4; j++ {
			if uniqueRanks[i+j] != uniqueRanks[i+j+1]+1 {
				isStraight = false
				break
			}
		}
		if isStraight {
			topRank := uniqueRanks[i]
			ranks := []Rank{topRank, topRank - 1, topRank - 2, topRank - 3, topRank - 4}
			return findCardsForStraight(analysis.cards, ranks), true
		}
	}
	return nil, false
}

// findCardsForStraight constructs a 5-card hand from a pool of cards, given a slice
// of 5 ranks that are known to form a straight. It picks one card for each rank.
func findCardsForStraight(pool []Card, ranks []Rank) []Card {
	straightCards := make([]Card, 0, 5)
	usedRanks := make(map[Rank]bool)
	for _, rank := range ranks {
		for _, card := range pool {
			if card.Rank == rank && !usedRanks[rank] {
				straightCards = append(straightCards, card)
				usedRanks[rank] = true
				break
			}
		}
	}
	return straightCards
}

// containsRank is a helper to check if a slice of ranks contains a target rank.
func containsRank(ranks []Rank, target Rank) bool {
	for _, r := range ranks {
		if r == target {
			return true
		}
	}
	return false
}

// findBestNOfAKind finds the highest-ranked set of N cards of the same rank.
// For example, with n=4, it finds the best Four of a Kind.
func findBestNOfAKind(rankCounts map[Rank]int, n int) (Rank, bool) {
	bestRank := Rank(-1)
	found := false
	for rank, count := range rankCounts {
		if count >= n {
			if rank > bestRank {
				bestRank = rank
				found = true
			}
		}
	}
	return bestRank, found
}

// findBestTwoPair finds the two highest-ranked pairs in the hand.
func findBestTwoPair(rankCounts map[Rank]int) (Rank, Rank, bool) {
	var pairs []Rank
	for rank, count := range rankCounts {
		if count >= 2 {
			pairs = append(pairs, rank)
		}
	}
	if len(pairs) < 2 {
		return -1, -1, false
	}
	// Sort pairs descending to find the highest two.
	sort.Slice(pairs, func(i, j int) bool { return pairs[i] > pairs[j] })
	return pairs[0], pairs[1], true
}

// findCardsByRank finds the first 'n' cards of a specific rank from a pool.
func findCardsByRank(pool []Card, rank Rank, n int) []Card {
	cards := make([]Card, 0, n)
	for _, c := range pool {
		if c.Rank == rank {
			cards = append(cards, c)
			if len(cards) == n {
				break
			}
		}
	}
	return cards
}

// findKickers finds 'n' kicker cards from a sorted pool of cards, excluding
// any ranks specified in excludeRanks. It returns the highest available cards
// that are not part of the main hand (e.g., not part of the pair in a OnePair hand).
func findKickers(sortedPool []Card, excludeRanks []Rank, n int) (bool, []Card) {
	kickers := make([]Card, 0, n)
	excludeMap := make(map[Rank]bool)
	for _, r := range excludeRanks {
		excludeMap[r] = true
	}
	for _, c := range sortedPool {
		if !excludeMap[c.Rank] {
			kickers = append(kickers, c)
			if len(kickers) == n {
				break
			}
		}
	}
	// Returns true only if the requested number of kickers was found.
	return len(kickers) == n, kickers
}

// compareHandResults compares two HandResult objects to determine which is stronger.
// It first compares by HandRank, then by HighValues for tie-breaking.
// Returns 1 if h1 > h2, -1 if h1 < h2, 0 if h1 == h2.
func compareHandResults(h1, h2 *HandResult) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}
	// Ranks are the same, compare kickers.
	for i := 0; i < len(h1.HighValues); i++ {
		if h1.HighValues[i] > h2.HighValues[i] {
			return 1
		}
		if h1.HighValues[i] < h2.HighValues[i] {
			return -1
		}
	}
	return 0 // Hands are identical.
}

// getHandRanks determines the order of hand ranks to be evaluated based on the game rules.
// It can either use the standard poker ranking or a custom ranking defined in the rules.
func getHandRanks(rules *HandRankingsRules) []HandRank {
	var handRankOrder []HandRank

	if rules.UseStandardRankings {
		// Standard poker hand rankings (from highest to lowest).
		handRankOrder = []HandRank{
			RoyalFlush,
			StraightFlush,
			FourOfAKind,
			FullHouse,
			Flush,
			Straight,
			ThreeOfAKind,
			TwoPair,
			OnePair,
			HighCard,
		}
	} else {
		// Start with a base set of standard ranks to be modified.
		baseOrder := []HandRank{
			RoyalFlush,
			StraightFlush,
			FourOfAKind,
			FullHouse,
			Flush,
			Straight,
			ThreeOfAKind,
			TwoPair,
			OnePair,
			HighCard,
		}
		handRankOrder = make([]HandRank, len(baseOrder))
		copy(handRankOrder, baseOrder)

		// Insert custom rankings into the order.
		for _, customRank := range rules.CustomRankings {
			hr, ok := handRankFromString(customRank.Name)
			if !ok {
				logrus.Warnf("Unknown custom hand ranking name: %s", customRank.Name)
				continue
			}

			insertAfterHr, ok := handRankFromString(customRank.InsertAfterRank)
			if !ok {
				logrus.Warnf("Unknown insert_after_rank name: %s for custom rank %s", customRank.InsertAfterRank, customRank.Name)
				continue
			}

			// Find the index where the custom rank should be inserted.
			insertIndex := -1
			for i, rank := range handRankOrder {
				if rank == insertAfterHr {
					insertIndex = i + 1 // Insert after the matched rank.
					break
				}
			}

			if insertIndex != -1 {
				// Insert the custom rank into the slice.
				handRankOrder = append(handRankOrder[:insertIndex], append([]HandRank{hr}, handRankOrder[insertIndex:]...)...)
			} else {
				logrus.Warnf("Could not find insertion point for custom rank %s after %s. Appending to end.", customRank.Name, customRank.InsertAfterRank)
				handRankOrder = append(handRankOrder, hr) // Fallback to appending.
			}
		}
	}

	return handRankOrder
}
