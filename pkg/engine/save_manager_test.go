package engine

import (
	"os"
	"path/filepath"
	"pls7-cli/pkg/poker"
	"testing"
	"time"
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
	game := createTestGame()

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

	// Verify loaded game matches original
	if loadedGame.HandCount != game.HandCount {
		t.Errorf("Expected hand count %d, got %d", game.HandCount, loadedGame.HandCount)
	}

	if len(loadedGame.Players) != len(game.Players) {
		t.Errorf("Expected %d players, got %d", len(game.Players), len(loadedGame.Players))
	}

	// Test listing saves
	saves, err := sm.ListSaves()
	if err != nil {
		t.Fatalf("Failed to list saves: %v", err)
	}

	if len(saves) != 1 {
		t.Errorf("Expected 1 save file, got %d", len(saves))
	}

	if saves[0].Filename != "test_save.json" {
		t.Errorf("Expected filename test_save.json, got %s", saves[0].Filename)
	}

	// Test validating save file
	err = sm.ValidateSaveFile("test_save")
	if err != nil {
		t.Errorf("Save file validation failed: %v", err)
	}

	// Test deleting save file
	err = sm.DeleteSave("test_save")
	if err != nil {
		t.Fatalf("Failed to delete save: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(savePath); !os.IsNotExist(err) {
		t.Error("Save file was not deleted")
	}
}

func TestSaveManagerMultipleSaves(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create SaveManager
	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Create test games with different states
	game1 := createTestGame()
	game1.HandCount = 1

	game2 := createTestGame()
	game2.HandCount = 2

	game3 := createTestGame()
	game3.HandCount = 3

	// Save multiple games
	err = sm.SaveGame(game1, "save1")
	if err != nil {
		t.Fatalf("Failed to save game1: %v", err)
	}

	// Add small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	err = sm.SaveGame(game2, "save2")
	if err != nil {
		t.Fatalf("Failed to save game2: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	err = sm.SaveGame(game3, "save3")
	if err != nil {
		t.Fatalf("Failed to save game3: %v", err)
	}

	// List saves and verify they are sorted by creation time (newest first)
	saves, err := sm.ListSaves()
	if err != nil {
		t.Fatalf("Failed to list saves: %v", err)
	}

	if len(saves) != 3 {
		t.Errorf("Expected 3 save files, got %d", len(saves))
	}

	// Verify sorting (newest first)
	if saves[0].Filename != "save3.json" {
		t.Errorf("Expected newest save to be save3.json, got %s", saves[0].Filename)
	}
	if saves[1].Filename != "save2.json" {
		t.Errorf("Expected second newest save to be save2.json, got %s", saves[1].Filename)
	}
	if saves[2].Filename != "save1.json" {
		t.Errorf("Expected oldest save to be save1.json, got %s", saves[2].Filename)
	}

	// Verify each save can be loaded correctly
	for i, save := range saves {
		loadedGame, err := sm.LoadGame(save.Filename)
		if err != nil {
			t.Fatalf("Failed to load save %d: %v", i, err)
		}

		expectedHandCount := 3 - i // save3=3, save2=2, save1=1
		if loadedGame.HandCount != expectedHandCount {
			t.Errorf("Save %d: Expected hand count %d, got %d", i, expectedHandCount, loadedGame.HandCount)
		}
	}
}

func TestSaveManagerErrorCases(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create SaveManager
	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Test saving with empty filename
	err = sm.SaveGame(createTestGame(), "")
	if err == nil {
		t.Error("Expected error for empty filename")
	}

	// Test loading non-existent file
	_, err = sm.LoadGame("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test deleting non-existent file
	err = sm.DeleteSave("nonexistent")
	if err == nil {
		t.Error("Expected error for deleting non-existent file")
	}

	// Test validating non-existent file
	err = sm.ValidateSaveFile("nonexistent")
	if err == nil {
		t.Error("Expected error for validating non-existent file")
	}
}

func TestSaveManagerFilenameSanitization(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create SaveManager
	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Test various invalid filenames
	testCases := []struct {
		input    string
		expected string
	}{
		{"test/save", "test_save"},
		{"test\\save", "test_save"},
		{"test:save", "test_save"},
		{"test*save", "test_save"},
		{"test?save", "test_save"},
		{"test\"save", "test_save"},
		{"test<save", "test_save"},
		{"test>save", "test_save"},
		{"test|save", "test_save"},
		{"  test  ", "test"},
		{"...test...", "test"},
		// Note: Empty string test is handled separately in TestSaveManagerErrorCases
	}

	for _, tc := range testCases {
		// Test sanitization by trying to save with the input filename
		game := createTestGame()
		err := sm.SaveGame(game, tc.input)
		if err != nil {
			t.Errorf("Failed to save with filename '%s': %v", tc.input, err)
			continue
		}

		// Check if the file was created with expected name
		expectedPath := filepath.Join(tempDir, tc.expected+".json")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created for input '%s'", expectedPath, tc.input)
		}

		// Clean up for next test
		os.Remove(expectedPath)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test SaveGameToFile
	game := createTestGame()
	err := SaveGameToFile(game, tempDir, "convenience_test")
	if err != nil {
		t.Fatalf("SaveGameToFile failed: %v", err)
	}

	// Test LoadGameFromFile
	loadedGame, err := LoadGameFromFile(tempDir, "convenience_test")
	if err != nil {
		t.Fatalf("LoadGameFromFile failed: %v", err)
	}

	if loadedGame.HandCount != game.HandCount {
		t.Errorf("Expected hand count %d, got %d", game.HandCount, loadedGame.HandCount)
	}

	// Test ListSaveFiles
	saves, err := ListSaveFiles(tempDir)
	if err != nil {
		t.Fatalf("ListSaveFiles failed: %v", err)
	}

	if len(saves) != 1 {
		t.Errorf("Expected 1 save file, got %d", len(saves))
	}

	// Test ValidateSaveFile
	err = ValidateSaveFile(tempDir, "convenience_test")
	if err != nil {
		t.Errorf("ValidateSaveFile failed: %v", err)
	}

	// Test DeleteSaveFile
	err = DeleteSaveFile(tempDir, "convenience_test")
	if err != nil {
		t.Fatalf("DeleteSaveFile failed: %v", err)
	}
}

func TestSaveManagerInvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create SaveManager
	sm, err := NewSaveManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create SaveManager: %v", err)
	}

	// Create an invalid JSON file
	invalidPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte("invalid json content"), 0644)
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

// Helper function to create a test game
func createTestGame() *Game {
	playerNames := []string{"YOU", "CPU1", "CPU2"}
	initialChips := 10000
	smallBlind := 100
	bigBlind := 200
	difficulty := DifficultyMedium
	rules := &poker.GameRules{
		Name:         "Test Game",
		Abbreviation: "TEST",
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
	return game
}
