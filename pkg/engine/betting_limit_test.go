package engine

import "testing"

// MockCalculator is a mock implementation of BettingLimitCalculator for testing.
type MockCalculator struct {
	min int
	max int
}

// CalculateBettingLimits for the mock, returns predefined values.
func (m *MockCalculator) CalculateBettingLimits(g *Game) (int, int) {
	return m.min, m.max
}

// TestBettingLimitCalculatorInterface ensures that any calculator can be used by the Game.
func TestBettingLimitCalculatorInterface(t *testing.T) {
	// This test primarily serves to enforce the existence and signature of the interface.
	var calculator BettingLimitCalculator
	calculator = &MockCalculator{min: 100, max: 1000}

	// Create a dummy game, as the calculator might need it.
	g := &Game{}

	min, max := calculator.CalculateBettingLimits(g)

	if min != 100 {
		t.Errorf("expected min bet of 100, got %d", min)
	}
	if max != 1000 {
		t.Errorf("expected max bet of 1000, got %d", max)
	}
}

// TestPotLimitCalculator tests the pot-limit betting logic.
func TestPotLimitCalculator(t *testing.T) {
	// Scenario:
	// Pot: 1500 (from blinds)
	// SB: CPU1, BB: CPU2
	// Player's turn (YOU), has 10000 chips.
	// Bet to call is 1000.
	// Last raise amount was BB post, so 1000.
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000, "PLS")
	g.Pot = 1500
	g.BetToCall = 1000
	g.LastRaiseAmount = 1000
	g.CurrentTurnPos = 0 // YOU's turn
	g.Players[0].CurrentBet = 0

	calculator := &PotLimitCalculator{}
	min, max := calculator.CalculateBettingLimits(g)

	// Min Raise: BetToCall (1000) + LastRaiseAmount (1000) = 2000
	// Max Raise (Pot Limit): Call (1000) + Pot (1500) + Call (1000) = 3500
	expectedMin := 2000
	expectedMax := 3500

	if min != expectedMin {
		t.Errorf("expected min raise to be %d, got %d", expectedMin, min)
	}
	if max != expectedMax {
		t.Errorf("expected max raise to be %d, got %d", expectedMax, max)
	}
}

// TestNoLimitCalculator tests the no-limit betting logic.
func TestNoLimitCalculator(t *testing.T) {
	// Scenario:
	// SB: CPU1, BB: CPU2
	// Player's turn (YOU), has 10000 chips.
	// Bet to call is 1000.
	// Last raise amount was 1000.
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000, "NLH")
	g.BetToCall = 1000
	g.LastRaiseAmount = 1000
	g.CurrentTurnPos = 0 // YOU's turn
	g.Players[0].CurrentBet = 0
	g.Players[0].Chips = 10000

	calculator := &NoLimitCalculator{}
	min, max := calculator.CalculateBettingLimits(g)

	// Min Raise: BetToCall (1000) + LastRaiseAmount (1000) = 2000
	// Max Raise (No Limit): All-in, which is the player's total chips (10000)
	expectedMin := 2000
	expectedMax := 10000

	if min != expectedMin {
		t.Errorf("expected min raise to be %d, got %d", expectedMin, min)
	}
	if max != expectedMax {
		t.Errorf("expected max raise to be %d, got %d", expectedMax, max)
	}
}
