package engine

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"testing"
)

func TestEvaluateHandStrength(t *testing.T) {
	testCases := []struct {
		name              string
		phase             GamePhase
		holeCardsStr      string
		communityCardsStr string
		expectedScore     float64
	}{
		{name: "Pre-Flop - High Pair (Aces)", phase: PhasePreFlop, holeCardsStr: "As Ac 2d", expectedScore: 49},
		{name: "Pre-Flop - Low Pair", phase: PhasePreFlop, holeCardsStr: "2s 2c Ad", expectedScore: 27},
		{name: "Pre-Flop - Suited Connectors", phase: PhasePreFlop, holeCardsStr: "8s 7s 2d", expectedScore: 4},
		{name: "Pre-Flop - Premium Suited High Cards", phase: PhasePreFlop, holeCardsStr: "As Ks Qs", expectedScore: 33},
		{name: "Post-Flop - Full House", phase: PhaseTurn, holeCardsStr: "As Ac Qd", communityCardsStr: "Ah Kc Kh 3d 4s", expectedScore: float64(poker.FullHouse)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Phase:          tc.phase,
				CommunityCards: poker.CardsFromStrings(tc.communityCardsStr),
				Rules:          &poker.GameRules{LowHand: poker.LowHandRules{Enabled: false}},
			}
			player := &Player{Hand: poker.CardsFromStrings(tc.holeCardsStr)}
			score := evaluateHandStrength(g, player)
			if score != tc.expectedScore {
				t.Errorf("Expected score %.2f, but got %.2f", tc.expectedScore, score)
			}
		})
	}
}

// TestCPUAction_ThresholdEdgeCases tests AI decisions at exact threshold boundaries.
func TestCPUAction_ThresholdEdgeCases(t *testing.T) {
	tpProfile := aiProfiles["Tight-Passive"]

	testCases := []struct {
		name           string
		handStrength   float64
		expectedAction ActionType
	}{
		// PlayHandThreshold for TP is 22
		{name: "Pre-Flop - Exactly at play threshold folds", handStrength: 21.99, expectedAction: ActionFold},
		{name: "Pre-Flop - At play threshold calls", handStrength: 22, expectedAction: ActionCall},
		{name: "Pre-Flop - Just above play threshold calls", handStrength: 22.01, expectedAction: ActionCall},
		// RaiseHandThreshold for TP is 28
		{name: "Pre-Flop - Just below raise threshold calls", handStrength: 27.99, expectedAction: ActionCall},
		{name: "Pre-Flop - At raise threshold raises", handStrength: 28, expectedAction: ActionRaise},
		{name: "Pre-Flop - Above raise threshold raises", handStrength: 30, expectedAction: ActionRaise},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Phase:     PhasePreFlop,
				Pot:       100,
				BetToCall: 10,
				BigBlind:  10,
				Rules:     &poker.GameRules{LowHand: poker.LowHandRules{Enabled: false}},
			}
			player := &Player{Profile: &tpProfile}
			g.handEvaluator = func(_ *Game, _ *Player) float64 { return tc.handStrength }

			r := rand.New(rand.NewSource(42))
			action := g.GetCPUAction(player, r)

			if action.Type != tc.expectedAction {
				t.Errorf("Strength %.2f: expected %v, got %v", tc.handStrength, tc.expectedAction, action.Type)
			}
		})
	}
}

// TestCPUAction_PostFlopStrengthBoundaries tests post-flop decisions at hand rank boundaries.
func TestCPUAction_PostFlopStrengthBoundaries(t *testing.T) {
	tpProfile := aiProfiles["Tight-Passive"]

	testCases := []struct {
		name           string
		handStrength   float64
		canCheck       bool
		expectedAction ActionType
	}{
		// Below OnePair → weak hand
		{name: "Weak hand can check", handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionCheck},
		// At OnePair → decent hand
		{name: "OnePair can check", handStrength: float64(poker.OnePair), canCheck: true, expectedAction: ActionCheck},
		{name: "OnePair must call", handStrength: float64(poker.OnePair), canCheck: false, expectedAction: ActionCall},
		// At TwoPair → strong hand, TP has low aggression (0.3) so with seed=100 they'll slow play
		{name: "TwoPair slow play", handStrength: float64(poker.TwoPair), canCheck: false, expectedAction: ActionCall},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Phase:     PhaseFlop,
				Pot:       100,
				BetToCall: 0,
				Rules:     &poker.GameRules{LowHand: poker.LowHandRules{Enabled: false}},
			}
			if !tc.canCheck {
				g.BetToCall = 10
			}
			player := &Player{Profile: &tpProfile, CurrentBet: 0}
			g.handEvaluator = func(_ *Game, _ *Player) float64 { return tc.handStrength }

			// Use a seed that produces r.Float64() > 0.3 (aggression) and > 0.05 (bluff)
			// to trigger slow play and no bluff
			r := rand.New(rand.NewSource(100))
			action := g.GetCPUAction(player, r)

			if action.Type != tc.expectedAction {
				t.Errorf("Strength %.0f canCheck=%v: expected %v, got %v", tc.handStrength, tc.canCheck, tc.expectedAction, action.Type)
			}
		})
	}
}

func TestCPUActionProfileBased(t *testing.T) {
	lagProfile := aiProfiles["Loose-Aggressive"]
	tpProfile := aiProfiles["Tight-Passive"]

	testCases := []struct {
		name           string
		seed           int64
		profile        *AIProfile
		phase          GamePhase
		handStrength   float64
		canCheck       bool
		expectedAction ActionType
	}{
		{name: "LAG AI - Bluffs with weak hand", seed: 2, profile: &lagProfile, phase: PhaseFlop, handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionBet},
		{name: "LAG AI - No Bluff on high random", seed: 12345, profile: &lagProfile, phase: PhaseFlop, handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionCheck},
		{name: "TP AI - No Bluff even with low random", seed: 1, profile: &tpProfile, phase: PhaseFlop, handStrength: float64(poker.HighCard), canCheck: true, expectedAction: ActionCheck},
		{name: "Pre-Flop - Folds below threshold", seed: 1, profile: &tpProfile, phase: PhasePreFlop, handStrength: 21, canCheck: false, expectedAction: ActionFold},
		{name: "Pre-Flop - Raises above threshold", seed: 1, profile: &tpProfile, phase: PhasePreFlop, handStrength: 29, canCheck: false, expectedAction: ActionRaise},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{
				Phase:     tc.phase,
				Pot:       100,
				BetToCall: 0,
				Rules:     &poker.GameRules{LowHand: poker.LowHandRules{Enabled: false}},
			}
			if !tc.canCheck {
				g.BetToCall = 10
			}
			player := &Player{Profile: tc.profile}

			g.handEvaluator = func(g *Game, p *Player) float64 { return tc.handStrength }

			r := rand.New(rand.NewSource(tc.seed))
			action := g.GetCPUAction(player, r)

			if action.Type != tc.expectedAction {
				t.Errorf("Expected action %v, but got %v", tc.expectedAction, action.Type)
			}
		})
	}
}
