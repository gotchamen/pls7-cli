package engine

import (
	"testing"
)

func TestAdvanceTurn(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000, "NLH")
	// Initial turn is after BB, so it should be YOU (pos 0)
	g.CurrentTurnPos = 0

	// Advance turn to CPU1 (pos 1)
	g.AdvanceTurn()
	if g.CurrentTurnPos != 1 {
		t.Errorf("Expected CurrentTurnPos to be 1, got %d", g.CurrentTurnPos)
	}

	// Advance turn to CPU2 (pos 2)
	g.AdvanceTurn()
	if g.CurrentTurnPos != 2 {
		t.Errorf("Expected CurrentTurnPos to be 2, got %d", g.CurrentTurnPos)
	}

	// Advance turn back to YOU (pos 0)
	g.AdvanceTurn()
	if g.CurrentTurnPos != 0 {
		t.Errorf("Expected CurrentTurnPos to be 0, got %d", g.CurrentTurnPos)
	}
}

func TestAdvanceTurn_SkipsEliminatedPlayer(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000, "NLH")
	g.CurrentTurnPos = 0
	g.Players[1].Status = PlayerStatusEliminated

	// Advance turn should skip CPU1 and go to CPU2
	g.AdvanceTurn()
	if g.CurrentTurnPos != 2 {
		t.Errorf("Expected CurrentTurnPos to be 2, got %d", g.CurrentTurnPos)
	}
}

func TestCurrentPlayer(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 500, 1000, "NLH")
	g.CurrentTurnPos = 1

	currentPlayer := g.CurrentPlayer()
	if currentPlayer.Name != "CPU1" {
		t.Errorf("Expected current player to be CPU1, but got %s", currentPlayer.Name)
	}
}
