// Package poker provides the core data structures and rules for playing poker.
// It includes types for cards, decks, hand evaluation, and game rules, forming
// the foundational building blocks for a poker game engine.
package poker

import (
	"github.com/sirupsen/logrus"
)

// OutsInfo stores the detailed results of an outs calculation. It contains all
// possible outs and categorizes them by the hand rank they would achieve.
type OutsInfo struct {
	// AllOuts is a slice containing all unique cards that can improve the hand.
	AllOuts []Card
	// OutsPerHandRank maps a specific hand rank to the cards that would complete it.
	// For example, OutsPerHandRank[Flush] would list all cards that complete a flush.
	OutsPerHandRank map[HandRank][]Card
}

// drawChecker is a function that checks if a specific draw exists and returns the out cards.
type drawChecker func(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card)

// collectDrawOuts checks a draw type and, if found, adds its outs to the results.
func collectDrawOuts(outsInfo *OutsInfo, allOutsMap map[Card]bool, rank HandRank, checker drawChecker, holeCards, communityCards []Card, seenCards map[Card]bool) {
	if hasDraw, outs := checker(holeCards, communityCards, seenCards); hasDraw {
		outsInfo.OutsPerHandRank[rank] = outs
		logrus.Debugf("CalculateOuts: outsInfo.OutsPerHandRank updated: %+v", outsInfo.OutsPerHandRank)
		for _, out := range outs {
			allOutsMap[out] = true
		}
	}
}

// CalculateOuts determines which cards from the remaining deck would improve the
// player's current hand. It checks for various draws (like flush, straight, etc.)
// and returns an OutsInfo struct containing the identified "out" cards.
func CalculateOuts(holeCards []Card, communityCards []Card, gameRules *GameRules) (bool, *OutsInfo) {
	currentHand, _ := EvaluateHand(holeCards, communityCards, gameRules)
	if currentHand == nil {
		return false, &OutsInfo{
			OutsPerHandRank: make(map[HandRank][]Card),
		}
	}

	outsInfo := &OutsInfo{
		OutsPerHandRank: make(map[HandRank][]Card),
	}
	allOutsMap := make(map[Card]bool)

	// Create a set of all cards currently in play to exclude them from potential outs.
	seenCards := buildSeenCards(holeCards, communityCards)

	// Check for draws in order from highest rank to lowest.
	// Only check for draws to hands better than the current hand.
	drawChecks := []struct {
		rank    HandRank
		checker drawChecker
	}{
		{SkipStraightFlush, hasSkipStraightFlushDraw},
		{StraightFlush, hasStraightFlushDraw},
		{FourOfAKind, hasFourOfAKindDraw},
		{FullHouse, hasFullHouseDraw},
		{Flush, hasFlushDraw},
		{SkipStraight, hasSkipStraightDraw},
		{Straight, hasStraightDraw},
		{ThreeOfAKind, hasThreeOfAKindDraw},
	}

	for _, dc := range drawChecks {
		if currentHand.Rank < dc.rank {
			collectDrawOuts(outsInfo, allOutsMap, dc.rank, dc.checker, holeCards, communityCards, seenCards)
		}
	}

	// Low hand draws (separate logic since the checker has a different signature).
	logrus.Tracef("CalculateOuts: Checking for low hands draws, lowGameEnabled: %v", gameRules.LowHand.Enabled)
	if gameRules.LowHand.Enabled {
		logrus.Debugf("CalculateOuts: Low game enabled, checking for low hand draws")
		maxRank := Rank(gameRules.LowHand.MaxRank)
		lowChecker := func(h []Card, c []Card, s map[Card]bool) (bool, []Card) {
			return hasLowHandDraw(h, c, s, maxRank)
		}
		collectDrawOuts(outsInfo, allOutsMap, HighCard, lowChecker, holeCards, communityCards, seenCards)
	}

	// Consolidate all unique outs into a single slice.
	for card := range allOutsMap {
		outsInfo.AllOuts = append(outsInfo.AllOuts, card)
	}

	return len(outsInfo.AllOuts) > 0, outsInfo
}

// buildSeenCards creates a set of all cards currently in play.
func buildSeenCards(holeCards, communityCards []Card) map[Card]bool {
	seenCards := make(map[Card]bool)
	for _, c := range holeCards {
		seenCards[c] = true
	}
	for _, c := range communityCards {
		seenCards[c] = true
	}
	return seenCards
}

// hasSkipStraightFlushDraw checks for a draw to a Skip Straight Flush.
// This requires having 4 cards of the same suit that are also 4 of the 5 cards
// needed for a Skip Straight.
func hasSkipStraightFlushDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	suitCounts := make(map[Suit]int)
	for _, c := range pool {
		suitCounts[c.Suit]++
	}

	for suit, count := range suitCounts {
		if count >= 4 { // Need at least 4 cards of the same suit for a draw.
			var suitedCards []Card
			for _, c := range pool {
				if c.Suit == suit {
					suitedCards = append(suitedCards, c)
				}
			}

			// Check for a skip straight draw within the suited cards.
			uniqueRanks := make(map[Rank]bool)
			for _, c := range suitedCards {
				uniqueRanks[c.Rank] = true
			}

			var outs []Card
			// Iterate through all possible ranks to see if adding one completes the hand.
			for r := Rank(2); r <= Ace; r++ {
				if !uniqueRanks[r] {
					outCard := Card{Rank: r, Suit: suit}
					if seenCards[outCard] {
						continue
					}

					// Temporarily add the potential out card and re-evaluate.
					tempPool := append(suitedCards, outCard)
					analysis := newHandAnalysis(tempPool)
					if skipStraightCards, ok := findSkipStraight(analysis); ok {
						// Verify the straight was completed by the card we added.
						found := false
						for _, sc := range skipStraightCards {
							if sc == outCard {
								found = true
								break
							}
						}
						if found {
							outs = append(outs, outCard)
						}
					}
				}
			}
			if len(outs) > 0 {
				return true, outs
			}
		}
	}

	return false, nil
}

// hasStraightFlushDraw checks for a draw to a Straight Flush.
// This requires having 4 cards of the same suit that are also 4 of the 5 cards
// needed for a Straight.
func hasStraightFlushDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	suitCounts := make(map[Suit]int)
	for _, c := range pool {
		suitCounts[c.Suit]++
	}

	for suit, count := range suitCounts {
		if count >= 4 { // Need at least 4 cards of the same suit for a draw.
			var suitedCards []Card
			for _, c := range pool {
				if c.Suit == suit {
					suitedCards = append(suitedCards, c)
				}
			}

			// Check for a straight draw within the suited cards.
			uniqueRanks := make(map[Rank]bool)
			for _, c := range suitedCards {
				uniqueRanks[c.Rank] = true
			}

			var outs []Card
			// Iterate through all possible ranks to see if adding one completes the hand.
			for r := Rank(2); r <= Ace; r++ {
				if !uniqueRanks[r] {
					outCard := Card{Rank: r, Suit: suit}
					if seenCards[outCard] {
						continue
					}

					// Temporarily add the potential out card and re-evaluate.
					tempPool := append(suitedCards, outCard)
					analysis := newHandAnalysis(tempPool)
					if sfCards, ok := findBestStraight(analysis); ok {
						// Verify the straight was completed by the card we added.
						found := false
						for _, sc := range sfCards {
							if sc == outCard {
								found = true
								break
							}
						}
						if found {
							outs = append(outs, outCard)
						}
					}
				}
			}
			if len(outs) > 0 {
				return true, outs
			}
		}
	}

	return false, nil
}

// hasFlushDraw checks for a draw to a Flush.
// This requires having 4 cards of the same suit.
func hasFlushDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	suitCounts := make(map[Suit]int)
	pool := append(holeCards, communityCards...)
	for _, c := range pool {
		suitCounts[c.Suit]++
	}
	logrus.Debugf("hasFlushDraw: Suit counts: %+v", suitCounts)

	for suit, count := range suitCounts {
		if count == 4 {
			var outs []Card
			// Any remaining card of that suit is an out.
			for r := Two; r <= Ace; r++ {
				outCard := Card{Suit: suit, Rank: r}
				if !seenCards[outCard] {
					outs = append(outs, outCard)
					logrus.Debugf("hasFlushDraw: Found flush draw out: %v, current outs: %v", outCard, outs)
				}
			}
			logrus.Debugf("hasFlushDraw: Final outs for flush draw: %v", outs)
			return true, outs
		}
	}

	return false, nil
}

// hasStraightDraw checks for a draw to a Straight. This includes open-ended
// straight draws (e.g., 5-6-7-8) and gutshot draws (e.g., 5-6-8-9).
func hasStraightDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	uniqueRanks := make(map[Rank]bool)
	for _, c := range pool {
		uniqueRanks[c.Rank] = true
	}
	logrus.Debugf("hasStraightDraw: Unique ranks in pool: %+v", uniqueRanks)

	var outs []Card
	// Iterate through all possible ranks to see if adding one completes a straight.
	for r := Two; r <= Ace; r++ {
		if !uniqueRanks[r] {
			// Temporarily add a card of this rank and re-evaluate.
			tempPool := append(pool, Card{Rank: r, Suit: Spade}) // Suit doesn't matter.
			analysis := newHandAnalysis(tempPool)
			if straightCards, ok := findBestStraight(analysis); ok {
				// Verify the straight was completed by the card we added.
				found := false
				for _, sc := range straightCards {
					if sc.Rank == r {
						found = true
						break
					}
				}
				if found {
					// If it completes, all 4 suits of that rank are outs.
					for s := Spade; s <= Club; s++ {
						outCard := Card{Rank: r, Suit: s}
						if !seenCards[outCard] {
							outs = append(outs, outCard)
							logrus.Debugf("hasStraightDraw: Found straight draw out: %v, current outs: %v", outCard, outs)
						}
					}
				}
			}
		}
	}

	logrus.Debugf("hasStraightDraw: Final outs for straight draw: %v", outs)
	return len(outs) > 0, outs
}

// hasSkipStraightDraw checks for a draw to a Skip Straight.
func hasSkipStraightDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	pool := append(holeCards, communityCards...)
	uniqueRanks := make(map[Rank]bool)
	for _, c := range pool {
		uniqueRanks[c.Rank] = true
	}
	logrus.Debugf("hasSkipStraightDraw: Unique ranks in pool: %+v", uniqueRanks)

	var outs []Card
	// Iterate through all possible ranks to see if adding one completes a skip straight.
	for r := Two; r <= Ace; r++ {
		if !uniqueRanks[r] {
			// Temporarily add a card of this rank and re-evaluate.
			tempPool := append(pool, Card{Rank: r, Suit: Spade}) // Suit doesn't matter.
			analysis := newHandAnalysis(tempPool)
			if skipStraightCards, ok := findSkipStraight(analysis); ok {
				// Verify the skip straight was completed by the card we added.
				found := false
				for _, sc := range skipStraightCards {
					if sc.Rank == r {
						found = true
						break
					}
				}
				if found {
					// If it completes, all 4 suits of that rank are outs.
					for s := Spade; s <= Club; s++ {
						outCard := Card{Rank: r, Suit: s}
						if !seenCards[outCard] {
							outs = append(outs, outCard)
							logrus.Debugf("hasSkipStraightDraw: Found skip straight draw out: %v, current outs: %v", outCard, outs)
						}
					}
				}
			}
		}
	}

	logrus.Debugf("hasSkipStraightDraw: Final outs for skip straight draw: %v", outs)
	return len(outs) > 0, outs
}

// hasThreeOfAKindDraw checks for a draw to Three of a Kind, which typically
// means holding a pocket pair and hoping to hit a third card of the same rank.
func hasThreeOfAKindDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	ppFound, ppRank := findPocketPair(holeCards)
	if !ppFound {
		logrus.Debugf("hasThreeOfAKindDraw: No pocket pair found in hole cards: %v", holeCards)
		return false, nil // Must have a pocket pair to have a "trips draw".
	}

	// Check if trips are already made.
	for _, c := range communityCards {
		if c.Rank == ppRank {
			logrus.Debugf("hasThreeOfAKindDraw: Found a community card with rank %v, trips made already", c.Rank)
			return false, nil // Trips already exist.
		}
	}

	// The two remaining cards of the pocket pair's rank are the outs.
	var outs []Card
	for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
		outCard := Card{Rank: ppRank, Suit: suit}
		if !seenCards[outCard] {
			outs = append(outs, outCard)
			logrus.Debugf(
				"hasThreeOfAKindDraw: Found trips draw out: %v, current outs: %v, pool: %v",
				outCard, outs, append(holeCards, communityCards...),
			)
		}
	}

	return len(outs) > 0, outs
}

// hasFullHouseDraw checks for a draw to a Full House. This is possible if the
// current hand is Three of a Kind (needs a pair) or Two Pair (needs a third card
// to one of the pairs).
func hasFullHouseDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	currentHand, _ := EvaluateHand(holeCards, communityCards, &GameRules{HandRankings: HandRankingsRules{UseStandardRankings: true}})
	if currentHand == nil {
		return false, nil
	}

	pool := append(holeCards, communityCards...)
	rankCounts := make(map[Rank]int)
	for _, c := range pool {
		rankCounts[c.Rank]++
	}
	logrus.Debugf("hasFullHouseDraw: Unique rank counts in pool: %+v", rankCounts)

	var outs []Card
	switch currentHand.Rank {
	case ThreeOfAKind:
		// We have trips. Any card that pairs one of our other cards (from the whole pool) is an out.
		tripRank := currentHand.HighValues[0]
		for rank := range rankCounts {
			// Skip the rank that already forms the trips.
			if rank == tripRank {
				continue
			}

			// Any other rank present in the pool is a potential pair for a full house.
			for s := Spade; s <= Club; s++ {
				outCard := Card{Rank: rank, Suit: s}
				if !seenCards[outCard] {
					outs = append(outs, outCard)
					logrus.Debugf(
						"hasFullHouseDraw: Found full house draw out: %v, current outs: %v, pool: %v",
						outCard, outs, pool,
					)
				}
			}
		}
	case TwoPair:
		// We have two pair. The 2 remaining cards of each pair's rank are outs.
		highPairRank := currentHand.HighValues[0]
		lowPairRank := currentHand.HighValues[1]
		for s := Spade; s <= Club; s++ {
			outCardHigh := Card{Rank: highPairRank, Suit: s}
			if !seenCards[outCardHigh] {
				outs = append(outs, outCardHigh)
			}
			outCardLow := Card{Rank: lowPairRank, Suit: s}
			if !seenCards[outCardLow] {
				outs = append(outs, outCardLow)
			}
		}
	default:
		logrus.Debugf(
			"hasFullHouseDraw: current hand rank is not suitable for full house draw: %v, "+
				"holeCards: %v, communityCards: %v, handRank: %v",
			currentHand, holeCards, communityCards, currentHand.Rank,
		)
		return false, nil // Must have trips or two pair to have a full house draw.
	}

	return len(outs) > 0, outs
}

// hasFourOfAKindDraw checks for a draw to Four of a Kind. This is only
// possible if the current hand is Three of a Kind.
func hasFourOfAKindDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool) (bool, []Card) {
	currentHand, _ := EvaluateHand(holeCards, communityCards, &GameRules{HandRankings: HandRankingsRules{UseStandardRankings: true}})
	if currentHand == nil || currentHand.Rank != ThreeOfAKind {
		logrus.Debugf("hasFourOfAKindDraw: Current hand is not trips, cannot draw four of a kind: %v", currentHand)
		return false, nil
	}

	// The one remaining card of the trips' rank is the out.
	tripRank := currentHand.HighValues[0]
	var outs []Card
	for s := Spade; s <= Club; s++ {
		outCard := Card{Rank: tripRank, Suit: s}
		if !seenCards[outCard] {
			outs = append(outs, outCard)
			logrus.Debugf("hasFourOfAKindDraw: Found four of a kind draw out: %v", outCard)
		}
	}
	return len(outs) > 0, outs
}

// hasLowHandDraw checks for a draw to a qualifying low hand (e.g., 8-low or better).
// This requires having 4 qualifying low cards, needing one more. The definition of
// a low card is determined by the maxRank parameter.
func hasLowHandDraw(holeCards []Card, communityCards []Card, seenCards map[Card]bool, maxRank Rank) (bool, []Card) {
	pool := append(holeCards, communityCards...)

	uniqueLowCards := make([]Card, 0, 8)
	usedRanks := make(map[Rank]bool)

	// Find all unique cards that qualify for a low hand based on maxRank.
	for _, card := range pool {
		if isLowCard(card, maxRank) && !usedRanks[card.Rank] {
			uniqueLowCards = append(uniqueLowCards, card)
			usedRanks[card.Rank] = true
		}
	}

	// If we have exactly 4 unique low cards, we have a draw.
	if len(uniqueLowCards) != 4 {
		logrus.Debugf("hasLowHandDraw: cannot draw since lowCards count is not 4: %v", usedRanks)
		return false, nil
	}

	var outs []Card
	// Any low rank we don't have yet is an out.
	for r := Two; r <= maxRank; r++ {
		if !usedRanks[r] {
			for s := Spade; s <= Club; s++ {
				outCard := Card{Rank: r, Suit: s}
				if !seenCards[outCard] {
					outs = append(outs, outCard)
					logrus.Debugf("hasLowHandDraw: Found low draw out: %v, current outs: %v", outCard, outs)
				}
			}
		}
	}
	// Check for Ace separately, as it's always a low card.
	if !usedRanks[Ace] {
		for s := Spade; s <= Club; s++ {
			outCard := Card{Rank: Ace, Suit: s}
			if !seenCards[outCard] {
				outs = append(outs, outCard)
				logrus.Debugf("hasLowHandDraw: Found low draw out: %v, current outs: %v", outCard, outs)
			}
		}
	}

	return len(outs) > 0, outs
}

// isLowCard checks if a card qualifies for a low hand based on a given maxRank.
// Ace is always considered a low card.
func isLowCard(c Card, maxRank Rank) bool {
	return c.Rank <= maxRank || c.Rank == Ace
}

// findPocketPair checks hole cards for a pair.
// It returns true and the rank of the pair if found.
func findPocketPair(holeCards []Card) (bool, Rank) {
	if len(holeCards) < 2 {
		return false, 0
	}
	rankCounts := make(map[Rank]int)
	for _, c := range holeCards {
		rankCounts[c.Rank]++
		if rankCounts[c.Rank] == 2 {
			return true, c.Rank
		}
	}
	logrus.Debugf("findPocketPair: Hole cards: %v, Pocket pair rank: 0", holeCards)
	return false, 0
}

// CalculateBreakEvenEquityBasedOnPotOdds calculates the minimum equity required to
// profitably make a call, based on the current pot size and the amount to call.
// Equity = Amount to Call / (Current Pot Size + Amount to Call)
func CalculateBreakEvenEquityBasedOnPotOdds(pot int, amountToCall int) float64 {
	if amountToCall <= 0 {
		return 0
	}
	totalPot := pot + amountToCall
	return float64(amountToCall) / float64(totalPot)
}

// CalculateEquityWithCards is a convenience function that first calculates outs
// and then uses the "Rule of 2 and 4" to estimate hand equity.
func CalculateEquityWithCards(ourHand, communityCards []Card) float64 {
	// Use standard rules for outs calculation, as custom rules might not apply to equity estimation.
	gameRules := &GameRules{
		HandRankings: HandRankingsRules{UseStandardRankings: true},
		LowHand:      LowHandRules{Enabled: false, MaxRank: 0},
	}
	// Note that outs are calculated without low hands draw.
	hasOuts, outsInfo := CalculateOuts(ourHand, communityCards, gameRules)
	if !hasOuts {
		return 0
	}
	return CalculateEquity(len(communityCards), len(outsInfo.AllOuts))
}

// CalculateEquity estimates the probability of winning a hand based on the number
// of outs and the current phase of the game (flop or turn). It uses the "Rule of
// 2 and 4":
// - On the flop: Equity ≈ Number of Outs * 4%
// - On the turn: Equity ≈ Number of Outs * 2%
// This is a widely used heuristic for quick equity estimation.
func CalculateEquity(numCommunityCards, numOuts int) float64 {
	if numOuts == 0 {
		return 0
	}

	switch numCommunityCards {
	case 3: // Flop
		return float64(numOuts*4) / 100.0
	case 4: // Turn
		return float64(numOuts*2) / 100.0
	default:
		logrus.Warnf("CalculateEquity: Invalid number of community cards (%d) for equity calculation.", numCommunityCards)
		return 0
	}
}
