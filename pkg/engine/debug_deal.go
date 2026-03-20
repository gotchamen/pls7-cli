package engine

import (
	"pls7-cli/pkg/poker"

	"github.com/sirupsen/logrus"
)

// playerHoleCardsForDebug maps game variant abbreviations to named debug hands.
// Each debug hand is a string of space-separated card codes (e.g., "As Ah Ad").
// Used only in DevMode to force specific hole cards for testing scenarios.
var playerHoleCardsForDebug = map[string]map[string]string{
	"PLS7": {
		"3As":        "As Ah Ad", // For testing outs for Four of a Kind
		"AQT-suited": "As Qs Ts", // For testing outs for Flush, Straight, and Skip Straight
		"AAK":        "As Ah Ks", // For testing outs for Three of a Kind
		"A23-suited": "As 2s 3s", // For testing outs for Straight, Flush, and low hand scenarios
	},
	"PLS": {
		"3As":        "As Ah Ad",
		"AQT-suited": "As Qs Ts",
		"AAK":        "As Ah Ks",
		"AKQ-suited": "As Ks Qs",
	},
	"PLO": {
		"AAKKds":  "As Ah Ks Kh", // #1 premium: top two pairs, double-suited nut flush draws
		"AAJTds":  "As Ah Js Th", // #2 premium: aces + Broadway wrap connectivity, double-suited
		"AAQQds":  "As Ah Qs Qh", // #3 premium: two premium pairs, double-suited
		"AAJJds":  "As Ah Js Jh", // #4 premium: aces + jacks, double-suited
		"KQJTds":  "Ks Kh Qs Qh", // #5 premium: max straight connectivity (Broadway wrap), double-suited
	},
	"PLO8": {
		"AAKKds":    "As Ah Ks Kh", // Top two pairs, double-suited nut flush draws
		"AAJTds":    "As Ah Js Th", // Aces + Broadway wrap, double-suited
		"AA23ds":    "As Ah 2s 3h", // Aces + nut low draw (A-2-3), suited ace for nut flush
		"AA23suited": "As 2s 3s Kh", // Nut low draw + nut flush draw in one suit
		"AA45ds":    "As Ah 4s 5h", // Aces + low draw + straight potential
	},
	"NLH": {
		"AA":        "As Ah",
		"KK":        "Ks Kh",
		"AK-suited": "As Ks",
		"KQ-suited": "Ks Qs",
	},
}

// defaultDebugHandKey is the debug hand key to use for each game variant.
var defaultDebugHandKey = map[string]string{
	"PLS7": "3As",
	"PLS":  "3As",
	"PLO":  "AAKKds",
	"PLO8": "AAKKds",
	"NLH":  "AA",
}

// GetDebugHands returns the registered debug hands for a given variant.
// Returns nil if the variant has no debug hands.
func GetDebugHands(variant string) map[string]string {
	return playerHoleCardsForDebug[variant]
}

// GetAllDebugHands returns the full debug hands map (variant → hand key → cards).
func GetAllDebugHands() map[string]map[string]string {
	return playerHoleCardsForDebug
}

// GetDefaultDebugHandKey returns the default debug hand key for a variant.
func GetDefaultDebugHandKey(variant string) (string, bool) {
	key, ok := defaultDebugHandKey[variant]
	return key, ok
}

// dealHoleCards dispatches to the appropriate dealing strategy based on DevMode.
func (g *Game) dealHoleCards() {
	if g.DevMode {
		g.dealDebugHoleCards()
	} else {
		g.dealNormalHoleCards()
	}
}

// dealNormalHoleCards deals cards to all active players in standard round-robin order.
func (g *Game) dealNormalHoleCards() {
	for i := 0; i < g.Rules.HoleCards.Count; i++ {
		for pos, p := range g.Players {
			if p.Status == PlayerStatusPlaying {
				card, _ := g.Deck.Deal()
				g.Players[pos].Hand = append(g.Players[pos].Hand, card)
			}
		}
	}
}

// dealDebugHoleCards deals predetermined hole cards to the human player (Players[0])
// and random cards to CPU players. Used only in DevMode.
func (g *Game) dealDebugHoleCards() {
	ruleAbbr := g.Rules.Abbreviation
	you := g.Players[0]

	if you.Status == PlayerStatusPlaying {
		debugHands, ok := playerHoleCardsForDebug[ruleAbbr]
		if !ok {
			logrus.Warnf("Unsupported rule abbreviation for debug hands: %s", ruleAbbr)
		} else {
			handKey := g.DebugHandKey
			if handKey == "" {
				var exists bool
				handKey, exists = defaultDebugHandKey[ruleAbbr]
				if !exists {
					logrus.Warnf("No default debug hand key for variant: %s, using first available", ruleAbbr)
					for k := range debugHands {
						handKey = k
						break
					}
				}
			}
			for _, card := range poker.CardsFromStrings(debugHands[handKey]) {
				dealtCard, err := g.Deck.DealForDebug(card)
				if err == nil {
					you.Hand = append(you.Hand, dealtCard)
				}
			}
		}
	}

	// Deal remaining cards randomly to CPUs.
	for i := 1; i < len(g.Players); i++ {
		for j := 0; j < g.Rules.HoleCards.Count; j++ {
			if g.Players[i].Status == PlayerStatusPlaying {
				card, _ := g.Deck.Deal()
				g.Players[i].Hand = append(g.Players[i].Hand, card)
			}
		}
	}
}
