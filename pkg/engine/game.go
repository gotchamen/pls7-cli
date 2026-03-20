package engine

import (
	"fmt"
	"math/rand"
	"os"
	"pls7-cli/pkg/poker"
	"time"

	"github.com/sirupsen/logrus"
)

// GamePhase defines the current stage of a poker hand, from the initial deal
// through multiple betting rounds to the final showdown.
type GamePhase int

// GamePhase constants represent the sequential phases of a poker hand.
const (
	PhasePreFlop  GamePhase = iota // PhasePreFlop is the first betting round, occurring after hole cards are dealt.
	PhaseFlop                      // PhaseFlop is the second betting round, after the first three community cards are dealt.
	PhaseTurn                      // PhaseTurn is the third betting round, after the fourth community card is dealt.
	PhaseRiver                     // PhaseRiver is the fourth and final betting round, after the fifth community card is dealt.
	PhaseShowdown                  // PhaseShowdown occurs after all betting is complete, where remaining players reveal their hands to determine the winner.
	PhaseHandOver                  // PhaseHandOver is a state indicating the hand is complete, the pot has been awarded, and the game is ready to start a new hand.
)

// String returns the human-readable name of the game phase.
func (gp GamePhase) String() string {
	return []string{"Pre-Flop", "Flop", "Turn", "River", "Showdown", "Hand Over"}[gp]
}

// Game is the central struct that encapsulates the entire state of a single poker hand.
// It acts as the orchestrator, managing players, cards, betting, and the progression
// through game phases. An instance of Game is a self-contained universe for one hand of poker.
type Game struct {
	// Players holds the slice of all players participating in the game session.
	// The order in this slice represents the seating arrangement at the table.
	Players []*Player
	// Deck is the deck of cards for the current hand. It is created new and shuffled
	// at the beginning of each hand.
	Deck *poker.Deck
	// CommunityCards are the shared cards dealt face-up on the board.
	CommunityCards []poker.Card
	// Pot holds the total amount of chips wagered by all players in the current hand.
	Pot int
	// DealerPos is the index in the Players slice corresponding to the player with the dealer button.
	DealerPos int
	// CurrentTurnPos is the index in the Players slice for the player whose turn it is to act.
	CurrentTurnPos int
	// Phase indicates the current stage of the hand (e.g., Pre-Flop, Flop, Turn).
	Phase GamePhase
	// BetToCall is the current highest bet amount that any player must match to stay in the hand.
	BetToCall int
	// LastRaiseAmount stores the size of the most recent raise, which is crucial for
	// calculating the minimum legal amount for a subsequent raise.
	LastRaiseAmount int
	// HandCount tracks the number of hands played in the current game session.
	HandCount int
	// SmallBlind is the size of the small blind for the current hand.
	SmallBlind int
	// BigBlind is the size of the big blind for the current hand.
	BigBlind int
	// Difficulty determines the skill level of the AI opponents.
	Difficulty Difficulty
	// handEvaluator is a function used to determine hand strength, primarily for AI decisions.
	// It can be replaced in tests for predictable outcomes.
	handEvaluator func(g *Game, player *Player) float64
	// DevMode enables development-specific features like detailed logging or predictable card dealing.
	DevMode bool
	// ShowsOuts enables a helper feature for human players to see their potential "outs" cards.
	ShowsOuts bool
	// Rules contains the complete set of rules for the specific poker variant being played.
	Rules *poker.GameRules
	// Rand is the single source of randomness for the entire game, used for shuffling and AI decisions.
	Rand *rand.Rand
	// BlindUpInterval is the number of hands after which the blinds increase. 0 disables this.
	BlindUpInterval int
	// BettingCalculator is an interface that calculates valid bet/raise sizes based on the game's betting limit.
	BettingCalculator BettingLimitCalculator
	// Aggressor points to the player who made the last aggressive action (bet or raise).
	// This is key to determining when a betting round ends.
	Aggressor *Player
	// ActionCloserPos is the position of the player who can close the action in a round
	// if no one raises. Pre-flop, this is the Big Blind. Post-flop, it's the first active
	// player to the left of the dealer.
	ActionCloserPos int
	// ActionsTakenThisRound counts player actions to help determine the end of a betting round.
	ActionsTakenThisRound int
	// TotalInitialChips stores the sum of all players' starting chips, used for sanity checks
	// to ensure chip conservation.
	TotalInitialChips int
	// DebugHandKey overrides the default debug hand selection when set (DevMode only).
	// Empty string means use the default for the game variant.
	DebugHandKey string
}

// CPUThinkTime returns the delay used to simulate CPU "thinking" for a more
// realistic game pace. In development mode, this delay is zero.
func (g *Game) CPUThinkTime() time.Duration {
	if g.DevMode {
		return 0 // No delay in dev mode.
	}
	return 500 * time.Millisecond // Default delay.
}

// NewGame is the constructor for the Game object. It initializes the game state,
// creates players, assigns AI profiles, and sets up the rules for the specified
// poker variant.
func NewGame(
	playerNames []string,
	initialChips int,
	smallBlind int,
	bigBlind int,
	difficulty Difficulty,
	rules *poker.GameRules,
	isDev bool,
	showsOuts bool,
	blindUpInterval int,
) *Game {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	players := make([]*Player, len(playerNames))
	cpuProfilesToAssign, err := cpuProfiles(difficulty, len(playerNames)-1)
	if err != nil {
		logrus.Errorf("Failed to get CPU profiles: %v", err)
		os.Exit(1)
	}

	if len(playerNames)-1 != len(cpuProfilesToAssign) {
		logrus.Errorf(
			"Mismatch in number of CPU profiles and players. %d != %d - 1",
			len(cpuProfilesToAssign), len(playerNames),
		)
		os.Exit(1)
	}

	// Create player objects, assigning AI profiles to CPUs.
	for i, name := range playerNames {
		isCPU := name != "YOU"
		players[i] = &Player{
			Name:     name,
			Chips:    initialChips,
			IsCPU:    isCPU,
			Position: i,
		}

		if isCPU {
			if profile, ok := aiProfiles[cpuProfilesToAssign[i-1]]; ok {
				players[i].Profile = &profile
			} else {
				logrus.Errorf("Unknown AI profile: %s", cpuProfilesToAssign[i-1])
				os.Exit(1)
			}
		}
	}

	// Select the appropriate betting calculator based on the game rules.
	var calculator BettingLimitCalculator
	switch rules.BettingLimit {
	case "pot_limit":
		calculator = &PotLimitCalculator{}
	case "no_limit":
		calculator = &NoLimitCalculator{}
	default:
		logrus.Fatalf("Unknown betting limit type: %s", rules.BettingLimit)
	}

	g := &Game{
		Players:           players,
		DealerPos:         -1, // Dealer position is set at the start of the first hand.
		SmallBlind:        smallBlind,
		BigBlind:          bigBlind,
		Difficulty:        difficulty,
		DevMode:           isDev,
		ShowsOuts:         showsOuts,
		Rules:             rules,
		Rand:              r,
		BlindUpInterval:   blindUpInterval,
		BettingCalculator: calculator,
		TotalInitialChips: initialChips * len(playerNames),
	}
	// Set the default hand evaluator function.
	g.handEvaluator = evaluateHandStrength
	return g
}

// String provides a formatted string representation of the current game state,
// useful for debugging and logging.
func (g *Game) String() string {
	dealerName := "N/A"
	if g.DealerPos >= 0 && g.DealerPos < len(g.Players) {
		dealerName = g.Players[g.DealerPos].Name
	}

	turnPlayerName := "N/A"
	if g.CurrentTurnPos >= 0 && g.CurrentTurnPos < len(g.Players) {
		turnPlayerName = g.Players[g.CurrentTurnPos].Name
	}

	return fmt.Sprintf(
		"[Game State]:\n"+
			"- Hand #%d, Phase: %s, Difficulty: %v\n"+
			"- Pot: %d, BetToCall: %d, LastRaise: %d\n"+
			"- Dealer: %s, Turn: %s\n"+
			"- Community: %v\n"+
			"- Players: %+v\n",
		g.HandCount, g.Phase, g.Difficulty, g.Pot, g.BetToCall, g.LastRaiseAmount,
		dealerName, turnPlayerName, g.CommunityCards, g.Players,
	)
}

// CalculateBettingLimits delegates the calculation of valid bet and raise sizes
// to the game's configured BettingLimitCalculator.
func (g *Game) CalculateBettingLimits() (minRaiseTotal int, maxRaiseTotal int) {
	return g.BettingCalculator.CalculateBettingLimits(g)
}

// CanShowOuts determines if the "show outs" helper should be displayed for a player.
// It is typically only enabled for the human player in development or easy modes.
func (g *Game) CanShowOuts(p *Player) bool {
	humanPlayerInPlay := p.Name == "YOU" && p.Status != PlayerStatusFolded
	availablePhase := g.Phase == PhaseFlop || g.Phase == PhaseTurn
	optionEnabled := g.DevMode || g.ShowsOuts
	return humanPlayerInPlay && optionEnabled && availablePhase
}

// minRaiseAmount calculates the minimum total bet required for a valid raise.
func (g *Game) minRaiseAmount() int {
	minRaiseIncrease := g.LastRaiseAmount
	if minRaiseIncrease == 0 {
		minRaiseIncrease = g.BetToCall
	}
	if g.BetToCall == 0 {
		minRaiseIncrease = g.BigBlind
	}
	return g.BetToCall + minRaiseIncrease
}

// cpuProfiles returns a slice of AI profile names to be assigned to CPU players,
// based on the selected game difficulty and the number of CPUs.
func cpuProfiles(difficulty Difficulty, numCPUs int) ([]string, error) {
	if numCPUs < 1 || numCPUs > 5 {
		return []string{}, fmt.Errorf("numCPUs must be between 1 and 5, got %d", numCPUs)
	}

	switch difficulty {
	case DifficultyEasy:
		// Easy difficulty features more passive opponents.
		return []string{
			"Loose-Passive", "Loose-Passive",
			"Loose-Passive", "Loose-Passive", "Loose-Passive",
		}[:numCPUs], nil
	case DifficultyMedium:
		// Medium difficulty introduces a mix of passive styles.
		return []string{
			"Loose-Passive", "Loose-Passive",
			"Tight-Passive", "Tight-Passive", "Tight-Passive",
		}[:numCPUs], nil
	case DifficultyHard:
		// Hard difficulty features more aggressive and varied opponents.
		return []string{
			"Tight-Passive",
			"Loose-Aggressive", "Loose-Aggressive",
			"Tight-Aggressive", "Tight-Aggressive",
		}[:numCPUs], nil
	default:
		return []string{}, fmt.Errorf("unknown difficulty: %v", difficulty)
	}
}
