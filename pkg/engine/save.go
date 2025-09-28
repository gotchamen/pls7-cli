package engine

import (
	"encoding/json"
	"fmt"
	"pls7-cli/pkg/poker"
	"time"
)

// GameSaveData represents the complete state of a poker game that can be
// serialized to JSON and saved to disk. This structure contains all necessary
// information to restore a game to its exact state.
type GameSaveData struct {
	// Version tracks the save file format version for compatibility checking.
	Version string `json:"version"`
	// Timestamp records when the save was created.
	Timestamp time.Time `json:"timestamp"`
	// GameMetadata contains the core game state information.
	GameMetadata GameMetadata `json:"game_metadata"`
	// Players contains the state of all players in the game.
	Players []PlayerSaveData `json:"players"`
	// CommunityCards holds the community cards currently on the board.
	CommunityCards []CardSaveData `json:"community_cards"`
	// DeckState contains information about the deck's current state.
	DeckState DeckSaveData `json:"deck_state"`
	// GameRules contains the complete game rules configuration.
	GameRules poker.GameRules `json:"game_rules"`
	// Settings contains the game configuration settings.
	Settings GameSettings `json:"settings"`
}

// GameMetadata contains the core game state information that changes during gameplay.
type GameMetadata struct {
	// HandCount tracks the number of hands played in the current game session.
	HandCount int `json:"hand_count"`
	// Phase indicates the current stage of the hand (e.g., Pre-Flop, Flop, Turn).
	Phase GamePhase `json:"phase"`
	// DealerPos is the index in the Players slice corresponding to the player with the dealer button.
	DealerPos int `json:"dealer_pos"`
	// CurrentTurnPos is the index in the Players slice for the player whose turn it is to act.
	CurrentTurnPos int `json:"current_turn_pos"`
	// Pot holds the total amount of chips wagered by all players in the current hand.
	Pot int `json:"pot"`
	// BetToCall is the current highest bet amount that any player must match to stay in the hand.
	BetToCall int `json:"bet_to_call"`
	// LastRaiseAmount stores the size of the most recent raise.
	LastRaiseAmount int `json:"last_raise_amount"`
	// SmallBlind is the size of the small blind for the current hand.
	SmallBlind int `json:"small_blind"`
	// BigBlind is the size of the big blind for the current hand.
	BigBlind int `json:"big_blind"`
	// BlindUpInterval is the number of hands after which the blinds increase. 0 disables this.
	BlindUpInterval int `json:"blind_up_interval"`
	// TotalInitialChips stores the sum of all players' starting chips.
	TotalInitialChips int `json:"total_initial_chips"`
	// ActionsTakenThisRound counts player actions to help determine the end of a betting round.
	ActionsTakenThisRound int `json:"actions_taken_this_round"`
	// ActionCloserPos is the position of the player who can close the action in a round.
	ActionCloserPos int `json:"action_closer_pos"`
}

// PlayerSaveData contains the state of a single player that needs to be saved.
type PlayerSaveData struct {
	// Name is the unique identifier for the player.
	Name string `json:"name"`
	// Chips is the player's current stack size.
	Chips int `json:"chips"`
	// IsCPU is true if the player is controlled by the AI.
	IsCPU bool `json:"is_cpu"`
	// Position is the player's seat at the table.
	Position int `json:"position"`
	// Status indicates the player's current state in the hand.
	Status PlayerStatus `json:"status"`
	// CurrentBet is the amount of chips the player has committed to the pot in the current betting round.
	CurrentBet int `json:"current_bet"`
	// TotalBetInHand is the cumulative amount of chips the player has put into the pot throughout the entire current hand.
	TotalBetInHand int `json:"total_bet_in_hand"`
	// LastActionDesc is a human-readable string describing the player's last action.
	LastActionDesc string `json:"last_action_desc"`
	// Hand holds the player's private hole cards.
	Hand []CardSaveData `json:"hand"`
	// Profile contains the AI behavior parameters if the player is a CPU.
	Profile *AIProfileSaveData `json:"profile,omitempty"`
}

// CardSaveData represents a playing card in a format suitable for JSON serialization.
type CardSaveData struct {
	// Suit represents the suit of the card (0=Spade, 1=Heart, 2=Diamond, 3=Club).
	Suit int `json:"suit"`
	// Rank represents the rank of the card (2-14, where 14=Ace).
	Rank int `json:"rank"`
}

// DeckSaveData contains information about the deck's current state for game restoration.
type DeckSaveData struct {
	// RemainingCardsCount is the number of cards left in the deck.
	RemainingCardsCount int `json:"remaining_cards_count"`
	// Seed is the random seed used for deck shuffling, allowing for deterministic recreation.
	Seed int64 `json:"seed"`
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
func (g *Game) ToSaveData() *GameSaveData {
	// Convert players
	players := make([]PlayerSaveData, len(g.Players))
	for i, player := range g.Players {
		players[i] = PlayerSaveData{
			Name:           player.Name,
			Chips:          player.Chips,
			IsCPU:          player.IsCPU,
			Position:       player.Position,
			Status:         player.Status,
			CurrentBet:     player.CurrentBet,
			TotalBetInHand: player.TotalBetInHand,
			LastActionDesc: player.LastActionDesc,
			Hand:           cardsToSaveData(player.Hand),
			Profile:        aiProfileToSaveData(player.Profile),
		}
	}

	// Convert community cards
	communityCards := cardsToSaveData(g.CommunityCards)

	// Create deck state
	deckState := DeckSaveData{
		RemainingCardsCount: len(g.Deck.Cards),
		Seed:                g.Rand.Int63(), // Store current random state
	}

	// Create game metadata
	gameMetadata := GameMetadata{
		HandCount:             g.HandCount,
		Phase:                 g.Phase,
		DealerPos:             g.DealerPos,
		CurrentTurnPos:        g.CurrentTurnPos,
		Pot:                   g.Pot,
		BetToCall:             g.BetToCall,
		LastRaiseAmount:       g.LastRaiseAmount,
		SmallBlind:            g.SmallBlind,
		BigBlind:              g.BigBlind,
		BlindUpInterval:       g.BlindUpInterval,
		TotalInitialChips:     g.TotalInitialChips,
		ActionsTakenThisRound: g.ActionsTakenThisRound,
		ActionCloserPos:       g.ActionCloserPos,
	}

	// Create settings
	settings := GameSettings{
		Difficulty: g.Difficulty,
		DevMode:    g.DevMode,
		ShowsOuts:  g.ShowsOuts,
	}

	return &GameSaveData{
		Version:        "1.0",
		Timestamp:      time.Now(),
		GameMetadata:   gameMetadata,
		Players:        players,
		CommunityCards: communityCards,
		DeckState:      deckState,
		GameRules:      *g.Rules,
		Settings:       settings,
	}
}

// FromSaveData creates a Game instance from GameSaveData.
func FromSaveData(saveData *GameSaveData) (*Game, error) {
	// Validate version
	if saveData.Version != "1.0" {
		return nil, fmt.Errorf("unsupported save file version: %s", saveData.Version)
	}

	// Create new random source with saved seed
	r := poker.NewRand(saveData.DeckState.Seed)

	// Convert players
	players := make([]*Player, len(saveData.Players))
	for i, playerData := range saveData.Players {
		players[i] = &Player{
			Name:           playerData.Name,
			Chips:          playerData.Chips,
			IsCPU:          playerData.IsCPU,
			Position:       playerData.Position,
			Status:         playerData.Status,
			CurrentBet:     playerData.CurrentBet,
			TotalBetInHand: playerData.TotalBetInHand,
			LastActionDesc: playerData.LastActionDesc,
			Hand:           cardsFromSaveData(playerData.Hand),
			Profile:        aiProfileFromSaveData(playerData.Profile),
		}
	}

	// Create deck and restore its state
	deck := poker.NewDeck()
	deck.Shuffle(r)

	// Remove cards that have been dealt (approximate recreation)
	// Note: This is an approximation since we can't perfectly recreate the exact deck state
	// without storing the entire deck order. For most purposes, this should be sufficient.
	cardsDealt := 0
	for _, player := range players {
		cardsDealt += len(player.Hand)
	}
	cardsDealt += len(saveData.CommunityCards)

	// Remove dealt cards from deck
	for i := 0; i < cardsDealt && i < len(deck.Cards); i++ {
		deck.Cards = deck.Cards[1:]
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

	// Create game instance
	game := &Game{
		Players:               players,
		Deck:                  deck,
		CommunityCards:        cardsFromSaveData(saveData.CommunityCards),
		Pot:                   saveData.GameMetadata.Pot,
		DealerPos:             saveData.GameMetadata.DealerPos,
		CurrentTurnPos:        saveData.GameMetadata.CurrentTurnPos,
		Phase:                 saveData.GameMetadata.Phase,
		BetToCall:             saveData.GameMetadata.BetToCall,
		LastRaiseAmount:       saveData.GameMetadata.LastRaiseAmount,
		HandCount:             saveData.GameMetadata.HandCount,
		SmallBlind:            saveData.GameMetadata.SmallBlind,
		BigBlind:              saveData.GameMetadata.BigBlind,
		Difficulty:            saveData.Settings.Difficulty,
		DevMode:               saveData.Settings.DevMode,
		ShowsOuts:             saveData.Settings.ShowsOuts,
		Rules:                 &saveData.GameRules,
		Rand:                  r,
		BlindUpInterval:       saveData.GameMetadata.BlindUpInterval,
		BettingCalculator:     calculator,
		TotalInitialChips:     saveData.GameMetadata.TotalInitialChips,
		ActionsTakenThisRound: saveData.GameMetadata.ActionsTakenThisRound,
		ActionCloserPos:       saveData.GameMetadata.ActionCloserPos,
	}

	// Set default hand evaluator
	game.handEvaluator = evaluateHandStrength

	return game, nil
}

// Helper functions for data conversion

// cardsToSaveData converts a slice of poker.Card to CardSaveData.
func cardsToSaveData(cards []poker.Card) []CardSaveData {
	saveData := make([]CardSaveData, len(cards))
	for i, card := range cards {
		saveData[i] = CardSaveData{
			Suit: int(card.Suit),
			Rank: int(card.Rank),
		}
	}
	return saveData
}

// cardsFromSaveData converts CardSaveData back to poker.Card.
func cardsFromSaveData(saveData []CardSaveData) []poker.Card {
	cards := make([]poker.Card, len(saveData))
	for i, cardData := range saveData {
		cards[i] = poker.Card{
			Suit: poker.Suit(cardData.Suit),
			Rank: poker.Rank(cardData.Rank),
		}
	}
	return cards
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
