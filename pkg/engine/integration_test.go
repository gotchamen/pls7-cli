package engine

import (
	"pls7-cli/pkg/poker"
	"testing"
)

// TestSaveLoadIntegration tests the complete save/load workflow
func TestSaveLoadIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test game
	game := createTestGameForIntegration()

	// Complete a hand to reach PhaseHandOver (required for saving)
	game.StartNewHand()
	game.Phase = PhaseHandOver

	// Save the game
	err := SaveGameToFile(game, tempDir, "integration_test")
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Load the game
	loadedGame, err := LoadGameFromFile(tempDir, "integration_test")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify basic game state - loaded game should be ready for new hand
	if loadedGame.Phase != PhaseHandOver {
		t.Errorf("Expected loaded game to be in HandOver phase, got %v", loadedGame.Phase)
	}

	if loadedGame.HandCount != game.HandCount {
		t.Errorf("Expected hand count %d, got %d", game.HandCount, loadedGame.HandCount)
	}

	if len(loadedGame.Players) != len(game.Players) {
		t.Errorf("Expected %d players, got %d", len(game.Players), len(loadedGame.Players))
	}

	// Verify player states
	for i, originalPlayer := range game.Players {
		loadedPlayer := loadedGame.Players[i]
		if originalPlayer.Name != loadedPlayer.Name {
			t.Errorf("Player %d: Expected name %s, got %s", i, originalPlayer.Name, loadedPlayer.Name)
		}
		if originalPlayer.Chips != loadedPlayer.Chips {
			t.Errorf("Player %d: Expected chips %d, got %d", i, originalPlayer.Chips, loadedPlayer.Chips)
		}
		if originalPlayer.IsCPU != loadedPlayer.IsCPU {
			t.Errorf("Player %d: Expected CPU status %t, got %t", i, originalPlayer.IsCPU, loadedPlayer.IsCPU)
		}
	}

	// Verify game settings
	if loadedGame.SmallBlind != game.SmallBlind {
		t.Errorf("Expected small blind %d, got %d", game.SmallBlind, loadedGame.SmallBlind)
	}

	if loadedGame.BigBlind != game.BigBlind {
		t.Errorf("Expected big blind %d, got %d", game.BigBlind, loadedGame.BigBlind)
	}
}

// TestSaveLoadErrorHandling tests error handling in save/load operations
func TestSaveLoadErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	// Test loading non-existent file
	_, err := LoadGameFromFile(tempDir, "nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// Test saving with invalid filename
	game := createTestGameForIntegration()
	game.StartNewHand()
	game.Phase = PhaseHandOver

	err = SaveGameToFile(game, tempDir, "")
	if err != nil {
		t.Error("Expected no error when saving with empty filename")
	}

	// Test saving to non-existent directory
	err = SaveGameToFile(game, "/nonexistent/directory", "test")
	if err == nil {
		t.Error("Expected error when saving to non-existent directory")
	}
}

// TestSaveLoadPerformance tests the performance of save/load operations
func TestSaveLoadPerformance(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test game
	game := createTestGameForIntegration()
	game.StartNewHand()
	game.Phase = PhaseHandOver

	// Test multiple save/load cycles
	for i := 0; i < 10; i++ {
		filename := "perf_test"
		if i > 0 {
			filename = "perf_test_" + string(rune('0'+i))
		}

		// Save
		err := SaveGameToFile(game, tempDir, filename)
		if err != nil {
			t.Fatalf("Failed to save game %d: %v", i, err)
		}

		// Load
		loadedGame, err := LoadGameFromFile(tempDir, filename)
		if err != nil {
			t.Fatalf("Failed to load game %d: %v", i, err)
		}

		// Verify basic state
		if loadedGame.HandCount != game.HandCount {
			t.Errorf("Game %d: Expected hand count %d, got %d", i, game.HandCount, loadedGame.HandCount)
		}

		if len(loadedGame.Players) != len(game.Players) {
			t.Errorf("Game %d: Expected %d players, got %d", i, len(game.Players), len(loadedGame.Players))
		}
	}
}

// Helper function to create a test game for integration tests
func createTestGameForIntegration() *Game {
	playerNames := []string{"YOU", "CPU1", "CPU2", "CPU3"}
	initialChips := 20000
	smallBlind := 200
	bigBlind := 400
	difficulty := DifficultyMedium
	rules := &poker.GameRules{
		Name:         "Integration Test Game",
		Abbreviation: "INT",
		BettingLimit: "no_limit",
		HoleCards: poker.HoleCardRules{
			Count:         2,
			UseConstraint: "any",
		},
		HandRankings: poker.HandRankingsRules{
			UseStandardRankings: true,
		},
		LowHand: poker.LowHandRules{
			Enabled: false,
		},
	}

	game := NewGame(playerNames, initialChips, smallBlind, bigBlind, difficulty, rules, false, false, 0)
	return game
}
