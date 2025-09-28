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

	// Start a hand to have some game state
	game.StartNewHand()

	// Convert to save data
	saveData := game.ToSaveData()

	// Verify basic structure
	if saveData.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", saveData.Version)
	}

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
	// Create test save data
	saveData := &GameSaveData{
		Version:   "1.0",
		Timestamp: time.Now(),
		GameMetadata: GameMetadata{
			HandCount:      1,
			Phase:          PhasePreFlop,
			DealerPos:      0,
			CurrentTurnPos: 1,
			Pot:            300,
			BetToCall:      200,
			SmallBlind:     100,
			BigBlind:       200,
		},
		Players: []PlayerSaveData{
			{
				Name:           "YOU",
				Chips:          9700,
				IsCPU:          false,
				Position:       0,
				Status:         PlayerStatusPlaying,
				CurrentBet:     200,
				TotalBetInHand: 200,
				Hand: []CardSaveData{
					{Suit: 0, Rank: 14}, // Ace of Spades
					{Suit: 1, Rank: 13}, // King of Hearts
				},
			},
			{
				Name:           "CPU1",
				Chips:          9800,
				IsCPU:          true,
				Position:       1,
				Status:         PlayerStatusPlaying,
				CurrentBet:     200,
				TotalBetInHand: 200,
				Hand: []CardSaveData{
					{Suit: 2, Rank: 12}, // Queen of Diamonds
					{Suit: 3, Rank: 11}, // Jack of Clubs
				},
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
		CommunityCards: []CardSaveData{},
		DeckState: DeckSaveData{
			RemainingCardsCount: 48,
			Seed:                1234567890,
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

	// Verify game state
	if game.HandCount != 1 {
		t.Errorf("Expected hand count 1, got %d", game.HandCount)
	}

	if game.Phase != PhasePreFlop {
		t.Errorf("Expected phase PreFlop, got %v", game.Phase)
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
	if len(player.Hand) != 2 {
		t.Errorf("Expected 2 hole cards, got %d", len(player.Hand))
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

func TestCardConversion(t *testing.T) {
	// Test card to save data conversion
	originalCards := []poker.Card{
		{Suit: poker.Spade, Rank: poker.Ace},
		{Suit: poker.Heart, Rank: poker.King},
		{Suit: poker.Diamond, Rank: poker.Queen},
		{Suit: poker.Club, Rank: poker.Jack},
	}

	saveData := cardsToSaveData(originalCards)
	if len(saveData) != 4 {
		t.Errorf("Expected 4 cards in save data, got %d", len(saveData))
	}

	// Test save data to card conversion
	convertedCards := cardsFromSaveData(saveData)
	if len(convertedCards) != 4 {
		t.Errorf("Expected 4 converted cards, got %d", len(convertedCards))
	}

	// Verify each card matches
	for i, original := range originalCards {
		converted := convertedCards[i]
		if original.Suit != converted.Suit {
			t.Errorf("Card %d: Expected suit %v, got %v", i, original.Suit, converted.Suit)
		}
		if original.Rank != converted.Rank {
			t.Errorf("Card %d: Expected rank %v, got %v", i, original.Rank, converted.Rank)
		}
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
	// Create test save data
	saveData := &GameSaveData{
		Version:   "1.0",
		Timestamp: time.Now(),
		GameMetadata: GameMetadata{
			HandCount:      1,
			Phase:          PhasePreFlop,
			DealerPos:      0,
			CurrentTurnPos: 1,
			Pot:            300,
			BetToCall:      200,
			SmallBlind:     100,
			BigBlind:       200,
		},
		Players: []PlayerSaveData{
			{
				Name:           "YOU",
				Chips:          9700,
				IsCPU:          false,
				Position:       0,
				Status:         PlayerStatusPlaying,
				CurrentBet:     200,
				TotalBetInHand: 200,
				Hand: []CardSaveData{
					{Suit: 0, Rank: 14},
					{Suit: 1, Rank: 13},
				},
			},
		},
		CommunityCards: []CardSaveData{},
		DeckState: DeckSaveData{
			RemainingCardsCount: 48,
			Seed:                1234567890,
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

	// Verify loaded data matches original
	if loadedSaveData.Version != saveData.Version {
		t.Errorf("Expected version %s, got %s", saveData.Version, loadedSaveData.Version)
	}

	if loadedSaveData.GameMetadata.HandCount != saveData.GameMetadata.HandCount {
		t.Errorf("Expected hand count %d, got %d", saveData.GameMetadata.HandCount, loadedSaveData.GameMetadata.HandCount)
	}

	if len(loadedSaveData.Players) != len(saveData.Players) {
		t.Errorf("Expected %d players, got %d", len(saveData.Players), len(loadedSaveData.Players))
	}
}
