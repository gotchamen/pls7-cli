package engine

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"sort"
	"time"
)

// byRank is a helper type that implements the sort.Interface for a slice of
// poker.Rank, allowing them to be sorted. It sorts in descending order (Ace high).
type byRank []poker.Rank

func (a byRank) Len() int           { return len(a) }
func (a byRank) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRank) Less(i, j int) bool { return a[i] > a[j] } // Sort descending

// aiProfiles contains a set of predefined AI personalities that dictate how a CPU
// player behaves. Each profile has different thresholds for playing, raising,
// and bluffing, creating varied opponent styles.
var aiProfiles = map[string]AIProfile{
	"Tight-Aggressive": {
		Name:               "Tight-Aggressive",
		PlayHandThreshold:  20,   // Plays only the top 20% of starting hands.
		RaiseHandThreshold: 25,   // Raises with the top 15% of hands.
		BluffingFrequency:  0.15, // Bluffs occasionally.
		AggressionFactor:   0.7,  // Highly likely to bet or raise with strong hands.
		MinRaiseMultiplier: 2.5,
		MaxRaiseMultiplier: 4.0,
	},
	"Loose-Aggressive": {
		Name:               "Loose-Aggressive",
		PlayHandThreshold:  10,   // Plays a wide range of hands (top 40%).
		RaiseHandThreshold: 20,   // Raises often.
		BluffingFrequency:  0.35, // Bluffs frequently.
		AggressionFactor:   0.9,  // Very aggressive.
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 3.5,
	},
	"Tight-Passive": {
		Name:               "Tight-Passive",
		PlayHandThreshold:  22,   // Very selective with starting hands.
		RaiseHandThreshold: 28,   // Rarely raises, only with premium hands.
		BluffingFrequency:  0.05, // Almost never bluffs.
		AggressionFactor:   0.3,  // Prefers to call rather than bet or raise.
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 2.5,
	},
	"Loose-Passive": {
		Name:               "Loose-Passive",
		PlayHandThreshold:  8,    // Plays many hands (calling station).
		RaiseHandThreshold: 24,   // Rarely raises.
		BluffingFrequency:  0.10, // Bluffs infrequently.
		AggressionFactor:   0.2,  // Very passive, calls often, folds to aggression.
		MinRaiseMultiplier: 2.0,
		MaxRaiseMultiplier: 3.0,
	},
}

// GetCPUAction determines the action for an AI-controlled player based on their
// assigned profile and the current game state.
func (g *Game) GetCPUAction(player *Player, r *rand.Rand) PlayerAction {
	strength := g.handEvaluator(g, player)
	time.Sleep(g.CPUThinkTime())

	if g.Phase == PhasePreFlop {
		return g.getPreFlopAction(player, strength)
	}
	return g.getPostFlopAction(player, r, strength)
}

// getPreFlopAction decides the CPU action before the flop based on hand strength thresholds.
func (g *Game) getPreFlopAction(player *Player, strength float64) PlayerAction {
	if strength < player.Profile.PlayHandThreshold {
		return PlayerAction{Type: ActionFold}
	}
	if strength >= player.Profile.RaiseHandThreshold {
		return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
	}
	return PlayerAction{Type: ActionCall}
}

// getPostFlopAction decides the CPU action after the flop based on hand strength,
// bluffing probability, and aggression factor.
func (g *Game) getPostFlopAction(player *Player, r *rand.Rand, strength float64) PlayerAction {
	canCheck := player.CurrentBet == g.BetToCall

	// Bluffing: attempt with weak hands based on profile frequency.
	if r.Float64() < player.Profile.BluffingFrequency && strength < float64(poker.OnePair) {
		return g.bluffAction(canCheck)
	}

	// Strong hands: value bet/raise or slow play.
	if strength >= float64(poker.TwoPair) {
		return g.strongHandAction(player, r)
	}

	// Decent hands: play cheaply.
	if strength >= float64(poker.OnePair) {
		if canCheck {
			return PlayerAction{Type: ActionCheck}
		}
		return PlayerAction{Type: ActionCall}
	}

	// Weak hands: check or fold based on pot odds.
	return g.weakHandAction(player, canCheck)
}

func (g *Game) bluffAction(canCheck bool) PlayerAction {
	if canCheck {
		return PlayerAction{Type: ActionBet, Amount: g.Pot / 2}
	}
	return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
}

func (g *Game) strongHandAction(player *Player, r *rand.Rand) PlayerAction {
	if r.Float64() < player.Profile.AggressionFactor {
		return PlayerAction{Type: ActionRaise, Amount: g.minRaiseAmount() * 2}
	}
	return PlayerAction{Type: ActionCall}
}

func (g *Game) weakHandAction(player *Player, canCheck bool) PlayerAction {
	if canCheck {
		return PlayerAction{Type: ActionCheck}
	}
	potOdds := float64(g.BetToCall) / float64(g.Pot+g.BetToCall)
	if potOdds < player.Profile.BluffingFrequency*0.5 {
		return PlayerAction{Type: ActionCall}
	}
	return PlayerAction{Type: ActionFold}
}

// Pre-flop hand strength scoring constants.
const (
	pairBonus          = 15.0 // Base bonus for holding a pair in the hole.
	suitedBonus        = 2.0  // Bonus for having suited hole cards.
	threeCardConnector = 5.0  // Bonus for three consecutive cards (e.g., 7-8-9).
	twoCardConnector   = 2.0  // Bonus for two consecutive cards (e.g., 7-8).
	highCloseBonus     = 1.0  // Bonus for high cards that are close in rank.
	highCardGapMax     = 5    // Max gap between highest and lowest card for the high-close bonus.
)

// highCardPoints maps face card ranks to their pre-flop scoring value.
var highCardPoints = map[poker.Rank]float64{
	poker.Ace: 10, poker.King: 8, poker.Queen: 7, poker.Jack: 6, poker.Ten: 5,
}

// evaluateHandStrength calculates a numerical score for a player's hand to guide
// AI decision-making. Post-flop uses the actual hand rank; pre-flop uses a custom heuristic.
func evaluateHandStrength(g *Game, player *Player) float64 {
	if g.Phase > PhasePreFlop {
		highHand, _ := poker.EvaluateHand(player.Hand, g.CommunityCards, g.Rules)
		if highHand != nil {
			return float64(highHand.Rank)
		}
		return 0
	}
	return evaluatePreFlopStrength(player.Hand)
}

// evaluatePreFlopStrength scores hole cards based on high-card values, pairs,
// suited cards, and connectivity.
func evaluatePreFlopStrength(hand []poker.Card) float64 {
	var score float64

	// High card points.
	for _, c := range hand {
		score += highCardPoints[c.Rank]
	}

	// Pair bonus.
	score += calculatePairBonus(hand)

	// Suited bonus.
	if hasSuitedCards(hand) {
		score += suitedBonus
	}

	// Connectivity bonus.
	score += calculateConnectivityBonus(hand)

	return score
}

// calculatePairBonus returns the bonus for holding a pair in the hole cards.
func calculatePairBonus(hand []poker.Card) float64 {
	if len(hand) >= 3 {
		if hand[0].Rank == hand[1].Rank || hand[0].Rank == hand[2].Rank || hand[1].Rank == hand[2].Rank {
			pairRank := hand[0].Rank
			if hand[1].Rank == hand[2].Rank {
				pairRank = hand[1].Rank
			}
			return pairBonus + float64(pairRank)
		}
	} else if len(hand) == 2 {
		if hand[0].Rank == hand[1].Rank {
			return pairBonus + float64(hand[0].Rank)
		}
	}
	return 0
}

// hasSuitedCards checks if any two hole cards share a suit.
func hasSuitedCards(hand []poker.Card) bool {
	if len(hand) >= 3 {
		return hand[0].Suit == hand[1].Suit || hand[0].Suit == hand[2].Suit || hand[1].Suit == hand[2].Suit
	}
	if len(hand) == 2 {
		return hand[0].Suit == hand[1].Suit
	}
	return false
}

// calculateConnectivityBonus scores hole cards based on sequential rank proximity.
func calculateConnectivityBonus(hand []poker.Card) float64 {
	var bonus float64
	if len(hand) >= 3 {
		ranks := []poker.Rank{hand[0].Rank, hand[1].Rank, hand[2].Rank}
		sort.Sort(byRank(ranks))

		if ranks[0] == ranks[1]+1 && ranks[1] == ranks[2]+1 {
			bonus += threeCardConnector
		} else if (ranks[0] == ranks[1]+1) || (ranks[1] == ranks[2]+1) {
			bonus += twoCardConnector
		}

		if ranks[0] >= poker.Ten && (ranks[0]-ranks[2] < poker.Rank(highCardGapMax)) {
			bonus += highCloseBonus
		}
	} else if len(hand) == 2 {
		ranks := []poker.Rank{hand[0].Rank, hand[1].Rank}
		sort.Sort(byRank(ranks))
		if ranks[0] == ranks[1]+1 {
			bonus += twoCardConnector
		}
		if ranks[0] >= poker.Ten && (ranks[0]-ranks[1] < poker.Rank(highCardGapMax)) {
			bonus += highCloseBonus
		}
	}
	return bonus
}
