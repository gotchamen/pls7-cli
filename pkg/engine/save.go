package engine

import (
	"encoding/json"
	"fmt"
	"pls7-cli/pkg/poker"
	"time"
)

// GameSaveData represents the simplified state of a poker game that can be
// serialized to JSON and saved to disk. This structure contains only the essential
// information needed to start a new hand with the same players and settings.
type GameSaveData struct {
	// Version tracks the save file format version for compatibility checking.
	Version string `json:"version"`
	// Timestamp records when the save was created.
	Timestamp time.Time `json:"timestamp"`
	// GameMetadata contains the core game state information.
	GameMetadata GameMetadata `json:"game_metadata"`
	// Players contains the state of all players in the game.
	Players []PlayerSaveData `json:"players"`
	// GameRules contains the complete game rules configuration.
	GameRules poker.GameRules `json:"game_rules"`
	// Settings contains the game configuration settings.
	Settings GameSettings `json:"settings"`
}

// GameMetadata contains the core game state information that changes during gameplay.
// Simplified to only store essential information for starting a new hand.
type GameMetadata struct {
	// HandCount tracks the number of hands played in the current game session.
	HandCount int `json:"hand_count"`
	// DealerPos is the index in the Players slice corresponding to the player with the dealer button.
	DealerPos int `json:"dealer_pos"`
	// SmallBlind is the size of the small blind for the current hand.
	SmallBlind int `json:"small_blind"`
	// BigBlind is the size of the big blind for the current hand.
	BigBlind int `json:"big_blind"`
	// BlindUpInterval is the number of hands after which the blinds increase. 0 disables this.
	BlindUpInterval int `json:"blind_up_interval"`
	// TotalInitialChips stores the sum of all players' starting chips.
	TotalInitialChips int `json:"total_initial_chips"`
}

// PlayerSaveData contains the state of a single player that needs to be saved.
// Simplified to only store essential player information for starting a new hand.
type PlayerSaveData struct {
	// Name is the unique identifier for the player.
	Name string `json:"name"`
	// Chips is the player's current stack size.
	Chips int `json:"chips"`
	// IsCPU is true if the player is controlled by the AI.
	IsCPU bool `json:"is_cpu"`
	// Position is the player's seat at the table.
	Position int `json:"position"`
	// Profile contains the AI behavior parameters if the player is a CPU.
	Profile *AIProfileSaveData `json:"profile,omitempty"`
}

// AIProfileSaveData contains the AI behavior parameters in a JSON-serializable format.
type AIProfileSaveData struct {
	// Name is the identifier for the profile.
	Name string `json:"name"`
	// PlayHandThreshold is the minimum hand strength score required for the AI to consider playing a hand pre-flop.
	PlayHandThreshold float64 `json:"play_hand_threshold"`
	// RaiseHandThreshold is the minimum hand strength score required for the AI to open with a raise pre-flop.
	RaiseHandThreshold float64 `json:"raise_hand_threshold"`
	// BluffingFrequency is the probability (0.0 to 1.0) that the AI will attempt a bluff with a weak hand.
	BluffingFrequency float64 `json:"bluffing_frequency"`
	// AggressionFactor is the probability (0.0 to 1.0) that the AI will choose to bet or raise instead of check or call.
	AggressionFactor float64 `json:"aggression_factor"`
	// MinRaiseMultiplier is the minimum multiplier for a raise amount.
	MinRaiseMultiplier float64 `json:"min_raise_multiplier"`
	// MaxRaiseMultiplier is the maximum multiplier for a raise amount.
	MaxRaiseMultiplier float64 `json:"max_raise_multiplier"`
}

// GameSettings contains the game configuration settings that affect gameplay.
type GameSettings struct {
	// Difficulty determines the skill level of the AI opponents.
	Difficulty Difficulty `json:"difficulty"`
	// DevMode enables development-specific features like detailed logging.
	DevMode bool `json:"dev_mode"`
	// ShowsOuts enables a helper feature for human players to see their potential "outs" cards.
	ShowsOuts bool `json:"shows_outs"`
}

// ToSaveData converts a Game instance to GameSaveData for serialization.
// Only saves essential information for starting a new hand.
func (g *Game) ToSaveData() *GameSaveData {
	// Convert players - only save basic player information
	players := make([]PlayerSaveData, len(g.Players))
	for i, player := range g.Players {
		players[i] = PlayerSaveData{
			Name:     player.Name,
			Chips:    player.Chips,
			IsCPU:    player.IsCPU,
			Position: player.Position,
			Profile:  aiProfileToSaveData(player.Profile),
		}
	}

	// Create simplified game metadata
	gameMetadata := GameMetadata{
		HandCount:         g.HandCount,
		DealerPos:         g.DealerPos,
		SmallBlind:        g.SmallBlind,
		BigBlind:          g.BigBlind,
		BlindUpInterval:   g.BlindUpInterval,
		TotalInitialChips: g.TotalInitialChips,
	}

	// Create settings
	settings := GameSettings{
		Difficulty: g.Difficulty,
		DevMode:    g.DevMode,
		ShowsOuts:  g.ShowsOuts,
	}

	return &GameSaveData{
		Version:      "2.0", // Updated version for simplified format
		Timestamp:    time.Now(),
		GameMetadata: gameMetadata,
		Players:      players,
		GameRules:    *g.Rules,
		Settings:     settings,
	}
}

// FromSaveData creates a Game instance from GameSaveData.
// Creates a new game with the same players and settings, ready to start a new hand.
func FromSaveData(saveData *GameSaveData) (*Game, error) {
	// Validate version - support both old and new formats
	if saveData.Version != "1.0" && saveData.Version != "2.0" {
		return nil, fmt.Errorf("unsupported save file version: %s", saveData.Version)
	}

	// Create new random source
	r := poker.NewRand(time.Now().UnixNano())

	// Convert players - create fresh player objects
	players := make([]*Player, len(saveData.Players))
	for i, playerData := range saveData.Players {
		players[i] = &Player{
			Name:     playerData.Name,
			Chips:    playerData.Chips,
			IsCPU:    playerData.IsCPU,
			Position: playerData.Position,
			Profile:  aiProfileFromSaveData(playerData.Profile),
		}
	}

	// Select appropriate betting calculator
	var calculator BettingLimitCalculator
	switch saveData.GameRules.BettingLimit {
	case "pot_limit":
		calculator = &PotLimitCalculator{}
	case "no_limit":
		calculator = &NoLimitCalculator{}
	default:
		return nil, fmt.Errorf("unknown betting limit type: %s", saveData.GameRules.BettingLimit)
	}

	// Create game instance - ready to start a new hand
	game := &Game{
		Players:           players,
		DealerPos:         saveData.GameMetadata.DealerPos,
		SmallBlind:        saveData.GameMetadata.SmallBlind,
		BigBlind:          saveData.GameMetadata.BigBlind,
		Difficulty:        saveData.Settings.Difficulty,
		DevMode:           saveData.Settings.DevMode,
		ShowsOuts:         saveData.Settings.ShowsOuts,
		Rules:             &saveData.GameRules,
		Rand:              r,
		BlindUpInterval:   saveData.GameMetadata.BlindUpInterval,
		BettingCalculator: calculator,
		TotalInitialChips: saveData.GameMetadata.TotalInitialChips,
		HandCount:         saveData.GameMetadata.HandCount,
		// Initialize new hand state
		Phase:                 PhaseHandOver, // Ready to start new hand
		CurrentTurnPos:        -1,            // Will be set when starting new hand
		Pot:                   0,             // Fresh pot
		BetToCall:             0,             // Fresh betting state
		LastRaiseAmount:       0,             // Fresh betting state
		ActionsTakenThisRound: 0,             // Fresh betting state
		ActionCloserPos:       -1,            // Will be set when starting new hand
	}

	// Set default hand evaluator
	game.handEvaluator = evaluateHandStrength

	return game, nil
}

// aiProfileToSaveData converts AIProfile to AIProfileSaveData.
func aiProfileToSaveData(profile *AIProfile) *AIProfileSaveData {
	if profile == nil {
		return nil
	}
	return &AIProfileSaveData{
		Name:               profile.Name,
		PlayHandThreshold:  profile.PlayHandThreshold,
		RaiseHandThreshold: profile.RaiseHandThreshold,
		BluffingFrequency:  profile.BluffingFrequency,
		AggressionFactor:   profile.AggressionFactor,
		MinRaiseMultiplier: profile.MinRaiseMultiplier,
		MaxRaiseMultiplier: profile.MaxRaiseMultiplier,
	}
}

// aiProfileFromSaveData converts AIProfileSaveData back to AIProfile.
func aiProfileFromSaveData(saveData *AIProfileSaveData) *AIProfile {
	if saveData == nil {
		return nil
	}
	return &AIProfile{
		Name:               saveData.Name,
		PlayHandThreshold:  saveData.PlayHandThreshold,
		RaiseHandThreshold: saveData.RaiseHandThreshold,
		BluffingFrequency:  saveData.BluffingFrequency,
		AggressionFactor:   saveData.AggressionFactor,
		MinRaiseMultiplier: saveData.MinRaiseMultiplier,
		MaxRaiseMultiplier: saveData.MaxRaiseMultiplier,
	}
}

// SaveToJSON serializes GameSaveData to JSON format.
func (gsd *GameSaveData) SaveToJSON() ([]byte, error) {
	return json.MarshalIndent(gsd, "", "  ")
}

// LoadFromJSON deserializes JSON data to GameSaveData.
func LoadFromJSON(data []byte) (*GameSaveData, error) {
	var saveData GameSaveData
	err := json.Unmarshal(data, &saveData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}
	return &saveData, nil
}
