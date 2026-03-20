package engine

import (
	"math/rand"
	"pls7-cli/internal/config"
	"reflect"
	"testing"
)

// TestHand_EliminatedPlayersAreSkipped tests that players with zero chips are properly excluded from a new hand.
func TestHand_EliminatedPlayersAreSkipped(t *testing.T) {
	playerNames := []string{"YOU", "CPU1", "CPU2", "CPU3"}
	initialChips := 100000
	rules, err := config.LoadGameRulesFromFile("../../rules/pls7.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}
	g := NewGame(playerNames, initialChips, 500, 1000, DifficultyMedium, rules, true, false, 0)

	// Manually eliminate two players
	g.Players[1].Chips = 0
	g.Players[1].Status = PlayerStatusEliminated
	g.Players[3].Chips = 0
	g.Players[3].Status = PlayerStatusEliminated

	// Start a new hand
	g.StartNewHand()

	// --- Assertion 1: Eliminated players should not have been dealt cards. ---
	// This test will FAIL before the fix.
	if len(g.Players[1].Hand) != 0 {
		t.Errorf("Expected eliminated player CPU 1 to have 0 cards, but got %d", len(g.Players[1].Hand))
	}
	if len(g.Players[3].Hand) != 0 {
		t.Errorf("Expected eliminated player CPU 3 to have 0 cards, but got %d", len(g.Players[3].Hand))
	}

	// --- Assertion 2: Active players should have been dealt cards. ---
	if len(g.Players[0].Hand) != 3 {
		t.Errorf("Expected active player YOU to have 3 cards, but got %d", len(g.Players[0].Hand))
	}
	if len(g.Players[2].Hand) != 3 {
		t.Errorf("Expected active player CPU 2 to have 3 cards, but got %d", len(g.Players[2].Hand))
	}
}

// FoldActionProvider always folds for all players (used in PlaySingleHand tests).
type FoldActionProvider struct{}

func (f *FoldActionProvider) GetAction(_ *Game, _ *Player, _ *rand.Rand) PlayerAction {
	return PlayerAction{Type: ActionFold}
}

// testObserver records GameObserver callbacks for verification.
type testObserver struct {
	phaseStartCount  int
	actionEventCount int
}

func (o *testObserver) OnPhaseStart(_ *Game) {
	o.phaseStartCount++
}

func (o *testObserver) OnPlayerAction(_ *ActionEvent) {
	o.actionEventCount++
}

func TestPlaySingleHand_AllFold_ReturnsLastPlayerWins(t *testing.T) {
	rules, err := config.LoadGameRulesFromFile("../../rules/nlh.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}
	g := NewGame([]string{"YOU", "CPU1", "CPU2"}, 10000, 100, 200, DifficultyEasy, rules, true, false, 0)
	provider := &FoldActionProvider{}
	observer := &testObserver{}

	result := g.PlaySingleHand(provider, observer)

	if result == nil {
		t.Fatal("Expected non-nil HandResult")
	}
	if result.IsShowdown {
		t.Error("Expected non-showdown result when all players fold")
	}
	if len(result.PotResults) == 0 {
		t.Error("Expected at least one pot distribution result")
	}
	if result.PotResults[0].AmountWon == 0 {
		t.Error("Expected winner to receive pot amount > 0")
	}
	if observer.phaseStartCount < 1 {
		t.Errorf("Expected at least 1 OnPhaseStart call, got %d", observer.phaseStartCount)
	}
	if observer.actionEventCount < 1 {
		t.Errorf("Expected at least 1 OnPlayerAction call, got %d", observer.actionEventCount)
	}
}

func TestPlaySingleHand_NilObserver_DoesNotPanic(t *testing.T) {
	rules, err := config.LoadGameRulesFromFile("../../rules/nlh.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}
	g := NewGame([]string{"YOU", "CPU1", "CPU2"}, 10000, 100, 200, DifficultyEasy, rules, true, false, 0)
	provider := &FoldActionProvider{}

	// Should not panic with nil observer
	result := g.PlaySingleHand(provider, nil)

	if result == nil {
		t.Fatal("Expected non-nil HandResult")
	}
}

func TestPlaySingleHand_HandCountIncrements(t *testing.T) {
	rules, err := config.LoadGameRulesFromFile("../../rules/nlh.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}
	g := NewGame([]string{"YOU", "CPU1", "CPU2"}, 10000, 100, 200, DifficultyEasy, rules, true, false, 0)
	provider := &FoldActionProvider{}

	g.PlaySingleHand(provider, nil)
	if g.HandCount != 1 {
		t.Errorf("Expected HandCount 1 after first hand, got %d", g.HandCount)
	}

	g.PlaySingleHand(provider, nil)
	if g.HandCount != 2 {
		t.Errorf("Expected HandCount 2 after second hand, got %d", g.HandCount)
	}
}

func TestPlaySingleHand_BlindEventOnBlindUp(t *testing.T) {
	rules, err := config.LoadGameRulesFromFile("../../rules/nlh.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}
	g := NewGame([]string{"YOU", "CPU1", "CPU2"}, 10000, 100, 200, DifficultyEasy, rules, true, false, 2)
	provider := &FoldActionProvider{}

	// First hand: no blind event
	result1 := g.PlaySingleHand(provider, nil)
	if result1.BlindEvent != nil {
		t.Error("Expected no blind event on first hand")
	}

	// Second hand: no blind event (interval=2 means blind up after hand 2)
	result2 := g.PlaySingleHand(provider, nil)
	if result2.BlindEvent != nil {
		t.Error("Expected no blind event on second hand")
	}

	// Third hand: blind event (hand 3, (3-1)%2==0)
	result3 := g.PlaySingleHand(provider, nil)
	if result3.BlindEvent == nil {
		t.Error("Expected blind event on third hand")
	}
}

// TestBlindUp_IntervalBoundary tests that blinds increase at correct intervals
// and the actual blind amounts double correctly.
func TestBlindUp_IntervalBoundary(t *testing.T) {
	rules, err := config.LoadGameRulesFromFile("../../rules/nlh.yml")
	if err != nil {
		t.Fatalf("Failed to load game rules: %v", err)
	}

	t.Run("Interval=3, blinds double at hands 4,7,10", func(t *testing.T) {
		g := NewGame([]string{"YOU", "CPU1", "CPU2"}, 100000, 100, 200, DifficultyEasy, rules, true, false, 3)
		provider := &FoldActionProvider{}

		// Hands 1-3: no blind increase
		for i := 1; i <= 3; i++ {
			result := g.PlaySingleHand(provider, nil)
			if result.BlindEvent != nil {
				t.Errorf("Hand %d: unexpected blind event", i)
			}
		}
		if g.SmallBlind != 100 {
			t.Errorf("After hand 3, expected SB=100, got %d", g.SmallBlind)
		}

		// Hand 4: blind increase (hand 4, (4-1)%3==0)
		result4 := g.PlaySingleHand(provider, nil)
		if result4.BlindEvent == nil {
			t.Fatal("Hand 4: expected blind event")
		}
		if g.SmallBlind != 200 || g.BigBlind != 400 {
			t.Errorf("Hand 4: expected SB=200/BB=400, got SB=%d/BB=%d", g.SmallBlind, g.BigBlind)
		}

		// Hands 5-6: no increase
		for i := 5; i <= 6; i++ {
			result := g.PlaySingleHand(provider, nil)
			if result.BlindEvent != nil {
				t.Errorf("Hand %d: unexpected blind event", i)
			}
		}

		// Hand 7: blind increase again
		result7 := g.PlaySingleHand(provider, nil)
		if result7.BlindEvent == nil {
			t.Fatal("Hand 7: expected blind event")
		}
		if g.SmallBlind != 400 || g.BigBlind != 800 {
			t.Errorf("Hand 7: expected SB=400/BB=800, got SB=%d/BB=%d", g.SmallBlind, g.BigBlind)
		}
	})

	t.Run("Interval=0 means no blind up ever", func(t *testing.T) {
		g := NewGame([]string{"YOU", "CPU1", "CPU2"}, 100000, 100, 200, DifficultyEasy, rules, true, false, 0)
		provider := &FoldActionProvider{}

		for i := 1; i <= 10; i++ {
			result := g.PlaySingleHand(provider, nil)
			if result.BlindEvent != nil {
				t.Errorf("Hand %d: unexpected blind event with interval=0", i)
			}
		}
		if g.SmallBlind != 100 {
			t.Errorf("Expected SB to stay 100 with interval=0, got %d", g.SmallBlind)
		}
	})
}

func TestNewGame_AssignsCorrectCalculator(t *testing.T) {
	testCases := []struct {
		name               string
		ruleStr            string
		expectedCalculator interface{}
	}{
		{
			name:               "Pot Limit Game",
			ruleStr:            "pls7",
			expectedCalculator: &PotLimitCalculator{},
		},
		{
			name:               "No Limit Game",
			ruleStr:            "nlh",
			expectedCalculator: &NoLimitCalculator{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := config.LoadGameRulesFromFile("../../rules/" + tc.ruleStr + ".yml")
			if err != nil {
				t.Fatalf("Failed to load game rules: %v", err)
			}
			g := NewGame([]string{"YOU", "CPU1"}, 1000, 500, 1000, DifficultyEasy, rules, false, false, 0)

			if g.BettingCalculator == nil {
				t.Fatal("g.BettingCalculator is nil")
			}

			actualType := reflect.TypeOf(g.BettingCalculator)
			expectedType := reflect.TypeOf(tc.expectedCalculator)

			if actualType != expectedType {
				t.Errorf("Expected calculator of type %v, but got %v", expectedType, actualType)
			}
		})
	}
}
