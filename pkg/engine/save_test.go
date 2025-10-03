package engine

import (
	"pls7-cli/pkg/poker"
	"testing"
	"time"
)

func TestGameSaveDataSerialization(t *testing.T) {
	// Create a test game
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

	// Complete a hand to reach PhaseHandOver (required for saving)
	game.StartNewHand()
	game.Phase = PhaseHandOver

	// Convert to save data
	saveData := game.ToSaveData()

	if saveData.GameMetadata.HandCount != 1 {
		t.Errorf("Expected hand count 1, got %d", saveData.GameMetadata.HandCount)
	}

	if len(saveData.Players) != 3 {
		t.Errorf("Expected 3 players, got %d", len(saveData.Players))
	}

	// Verify player data
	for i, player := range saveData.Players {
		if player.Name != playerNames[i] {
			t.Errorf("Expected player name %s, got %s", playerNames[i], player.Name)
		}
		// Chips will be different due to blind posting, so just check they're reasonable
		if player.Chips < 0 || player.Chips > initialChips {
			t.Errorf("Player %d chips %d is not in valid range [0, %d]", i, player.Chips, initialChips)
		}
		if player.IsCPU != (i != 0) {
			t.Errorf("Expected CPU status %t for player %d, got %t", i != 0, i, player.IsCPU)
		}
	}
}

func TestGameSaveDataDeserialization(t *testing.T) {
	// Create test save data with simplified structure
	saveData := &GameSaveData{
		Timestamp: time.Now(),
		GameMetadata: GameMetadata{
			HandCount:         1,
			DealerPos:         0,
			SmallBlind:        100,
			BigBlind:          200,
			BlindUpInterval:   0,
			TotalInitialChips: 30000,
		},
		Players: []PlayerSaveData{
			{
				Name:     "YOU",
				Chips:    9700,
				IsCPU:    false,
				Position: 0,
			},
			{
				Name:     "CPU1",
				Chips:    9800,
				IsCPU:    true,
				Position: 1,
				Profile: &AIProfileSaveData{
					Name:               "Tight-Passive",
					PlayHandThreshold:  0.6,
					RaiseHandThreshold: 0.8,
					BluffingFrequency:  0.1,
					AggressionFactor:   0.3,
					MinRaiseMultiplier: 2.0,
					MaxRaiseMultiplier: 4.0,
				},
			},
		},
		GameRules: poker.GameRules{
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
		},
		Settings: GameSettings{
			Difficulty: DifficultyMedium,
			DevMode:    false,
			ShowsOuts:  false,
		},
	}

	// Convert back to game
	game, err := FromSaveData(saveData)
	if err != nil {
		t.Fatalf("Failed to create game from save data: %v", err)
	}

	// Verify game state - should be ready to start new hand
	if game.HandCount != 1 {
		t.Errorf("Expected hand count 1, got %d", game.HandCount)
	}

	if game.Phase != PhaseHandOver {
		t.Errorf("Expected phase HandOver (ready for new hand), got %v", game.Phase)
	}

	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(game.Players))
	}

	// Verify first player
	player := game.Players[0]
	if player.Name != "YOU" {
		t.Errorf("Expected player name YOU, got %s", player.Name)
	}
	if player.Chips != 9700 {
		t.Errorf("Expected player chips 9700, got %d", player.Chips)
	}
	if player.IsCPU {
		t.Error("Expected human player, got CPU")
	}

	// Verify second player
	cpuPlayer := game.Players[1]
	if cpuPlayer.Name != "CPU1" {
		t.Errorf("Expected player name CPU1, got %s", cpuPlayer.Name)
	}
	if !cpuPlayer.IsCPU {
		t.Error("Expected CPU player, got human")
	}
	if cpuPlayer.Profile == nil {
		t.Error("Expected AI profile for CPU player")
	} else if cpuPlayer.Profile.Name != "Tight-Passive" {
		t.Errorf("Expected AI profile name Tight-Passive, got %s", cpuPlayer.Profile.Name)
	}
}

func TestAIProfileConversion(t *testing.T) {
	// Test AI profile to save data conversion
	originalProfile := &AIProfile{
		Name:               "Test Profile",
		PlayHandThreshold:  0.7,
		RaiseHandThreshold: 0.9,
		BluffingFrequency:  0.2,
		AggressionFactor:   0.5,
		MinRaiseMultiplier: 2.5,
		MaxRaiseMultiplier: 5.0,
	}

	saveData := aiProfileToSaveData(originalProfile)
	if saveData == nil {
		t.Error("Expected non-nil save data")
	}

	// Test save data to AI profile conversion
	convertedProfile := aiProfileFromSaveData(saveData)
	if convertedProfile == nil {
		t.Error("Expected non-nil converted profile")
	}

	// Verify all fields match
	if originalProfile.Name != convertedProfile.Name {
		t.Errorf("Expected name %s, got %s", originalProfile.Name, convertedProfile.Name)
	}
	if originalProfile.PlayHandThreshold != convertedProfile.PlayHandThreshold {
		t.Errorf("Expected PlayHandThreshold %f, got %f", originalProfile.PlayHandThreshold, convertedProfile.PlayHandThreshold)
	}
	if originalProfile.RaiseHandThreshold != convertedProfile.RaiseHandThreshold {
		t.Errorf("Expected RaiseHandThreshold %f, got %f", originalProfile.RaiseHandThreshold, convertedProfile.RaiseHandThreshold)
	}
	if originalProfile.BluffingFrequency != convertedProfile.BluffingFrequency {
		t.Errorf("Expected BluffingFrequency %f, got %f", originalProfile.BluffingFrequency, convertedProfile.BluffingFrequency)
	}
	if originalProfile.AggressionFactor != convertedProfile.AggressionFactor {
		t.Errorf("Expected AggressionFactor %f, got %f", originalProfile.AggressionFactor, convertedProfile.AggressionFactor)
	}
	if originalProfile.MinRaiseMultiplier != convertedProfile.MinRaiseMultiplier {
		t.Errorf("Expected MinRaiseMultiplier %f, got %f", originalProfile.MinRaiseMultiplier, convertedProfile.MinRaiseMultiplier)
	}
	if originalProfile.MaxRaiseMultiplier != convertedProfile.MaxRaiseMultiplier {
		t.Errorf("Expected MaxRaiseMultiplier %f, got %f", originalProfile.MaxRaiseMultiplier, convertedProfile.MaxRaiseMultiplier)
	}
}

func TestJSONSerialization(t *testing.T) {
	// Create test save data with simplified structure
	saveData := &GameSaveData{
		Timestamp: time.Now(),
		GameMetadata: GameMetadata{
			HandCount:         1,
			DealerPos:         0,
			SmallBlind:        100,
			BigBlind:          200,
			BlindUpInterval:   0,
			TotalInitialChips: 20000,
		},
		Players: []PlayerSaveData{
			{
				Name:     "YOU",
				Chips:    9700,
				IsCPU:    false,
				Position: 0,
			},
		},
		GameRules: poker.GameRules{
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
		},
		Settings: GameSettings{
			Difficulty: DifficultyMedium,
			DevMode:    false,
			ShowsOuts:  false,
		},
	}

	// Test JSON serialization
	jsonData, err := saveData.SaveToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// Test JSON deserialization
	loadedSaveData, err := LoadFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}

	if loadedSaveData.GameMetadata.HandCount != saveData.GameMetadata.HandCount {
		t.Errorf("Expected hand count %d, got %d", saveData.GameMetadata.HandCount, loadedSaveData.GameMetadata.HandCount)
	}

	if len(loadedSaveData.Players) != len(saveData.Players) {
		t.Errorf("Expected %d players, got %d", len(saveData.Players), len(loadedSaveData.Players))
	}
}

// TestLoadGameFunctionality tests the complete load functionality
func TestLoadGameFunctionality(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test game and save it
	game := createTestGameForLoad()

	// Save the game
	err := SaveGameToFile(game, tempDir, "test_load")
	if err != nil {
		t.Fatalf("Failed to save test game: %v", err)
	}

	// Load the game
	loadedGame, err := LoadGameFromFile(tempDir, "test_load")
	if err != nil {
		t.Fatalf("Failed to load test game: %v", err)
	}

	// Verify loaded game matches original basic info
	if loadedGame.HandCount != game.HandCount {
		t.Errorf("Expected hand count %d, got %d", game.HandCount, loadedGame.HandCount)
	}

	// Loaded game should be ready to start new hand
	if loadedGame.Phase != PhaseHandOver {
		t.Errorf("Expected phase HandOver (ready for new hand), got %v", loadedGame.Phase)
	}

	if len(loadedGame.Players) != len(game.Players) {
		t.Errorf("Expected %d players, got %d", len(game.Players), len(loadedGame.Players))
	}

	// Verify player states - only basic info should match
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

// Helper function to create a test game for loading tests
func createTestGameForLoad() *Game {
	playerNames := []string{"YOU", "CPU1", "CPU2"}
	initialChips := 10000
	smallBlind := 100
	bigBlind := 200
	difficulty := DifficultyMedium
	rules := &poker.GameRules{
		Name:         "Test Load Game",
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
	game.Phase = PhaseHandOver // Set to HandOver so it can be saved
	return game
}
