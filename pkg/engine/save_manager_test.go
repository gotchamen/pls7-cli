package engine

import (
	"os"
	"path/filepath"
	"pls7-cli/pkg/poker"
	"testing"
)

func TestSaveManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create SaveManager
	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Verify save directory was created
	if sm.GetSaveDir() != tempDir {
		t.Errorf("Expected save dir %s, got %s", tempDir, sm.GetSaveDir())
	}

	// Create a test game
	game := createTestGameForSaveManager()

	// Test saving a game
	err = sm.SaveGame(game, "test_save")
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Verify file was created
	savePath := filepath.Join(tempDir, "test_save.json")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}

	// Test loading the game
	loadedGame, err := sm.LoadGame("test_save")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify loaded game matches original basic info
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
	}
}

func TestSaveManagerMultipleSaves(t *testing.T) {
	tempDir := t.TempDir()

	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Create multiple test games
	game1 := createTestGameForSaveManager()
	game1.HandCount = 1

	game2 := createTestGameForSaveManager()
	game2.HandCount = 2

	// Save both games
	err = sm.SaveGame(game1, "game1")
	if err != nil {
		t.Fatalf("Failed to save game1: %v", err)
	}

	err = sm.SaveGame(game2, "game2")
	if err != nil {
		t.Fatalf("Failed to save game2: %v", err)
	}

	// List saves
	saves, err := sm.ListSaves()
	if err != nil {
		t.Fatalf("Failed to list saves: %v", err)
	}

	if len(saves) != 2 {
		t.Errorf("Expected 2 saves, got %d", len(saves))
	}

	// Load both games and verify
	loadedGame1, err := sm.LoadGame("game1")
	if err != nil {
		t.Fatalf("Failed to load game1: %v", err)
	}

	loadedGame2, err := sm.LoadGame("game2")
	if err != nil {
		t.Fatalf("Failed to load game2: %v", err)
	}

	if loadedGame1.HandCount != 1 {
		t.Errorf("Expected game1 hand count 1, got %d", loadedGame1.HandCount)
	}

	if loadedGame2.HandCount != 2 {
		t.Errorf("Expected game2 hand count 2, got %d", loadedGame2.HandCount)
	}
}

func TestSaveManagerErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Test loading non-existent file
	_, err = sm.LoadGame("nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// Test deleting non-existent file
	err = sm.DeleteSave("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent file")
	}

	// Test validating non-existent file
	err = sm.ValidateSaveFile("nonexistent")
	if err == nil {
		t.Error("Expected error when validating non-existent file")
	}
}

func TestSaveManagerFilenameSanitization(t *testing.T) {
	tempDir := t.TempDir()

	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	game := createTestGameForSaveManager()

	// Test various problematic filenames
	problematicFilenames := []string{
		"test/save",
		"test\\save",
		"test:save",
		"test*save",
		"test?save",
		"test\"save",
		"test<save",
		"test>save",
		"test|save",
		"  test  ",
		"...test...",
	}

	for _, filename := range problematicFilenames {
		err = sm.SaveGame(game, filename)
		if err != nil {
			t.Errorf("Failed to save with filename '%s': %v", filename, err)
		}

		// Verify file was created with sanitized name
		saves, err := sm.ListSaves()
		if err != nil {
			t.Errorf("Failed to list saves for filename '%s': %v", filename, err)
		}

		if len(saves) == 0 {
			t.Errorf("No saves found for filename '%s'", filename)
		}
	}
}

func TestSaveManagerInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Create a file with invalid JSON
	invalidJSONPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidJSONPath, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Test loading invalid JSON
	_, err = sm.LoadGame("invalid")
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}

	// Test validating invalid JSON
	err = sm.ValidateSaveFile("invalid")
	if err == nil {
		t.Error("Expected error when validating invalid JSON")
	}
}

// Helper function to create a test game for SaveManager tests
func createTestGameForSaveManager() *Game {
	playerNames := []string{"YOU", "CPU1", "CPU2"}
	initialChips := 15000
	smallBlind := 150
	bigBlind := 300
	difficulty := DifficultyMedium
	rules := &poker.GameRules{
		Name:         "SaveManager Test Game",
		Abbreviation: "SM",
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
	game.StartNewHand()
	game.Phase = PhaseHandOver // Set to HandOver so it can be saved
	return game
}
