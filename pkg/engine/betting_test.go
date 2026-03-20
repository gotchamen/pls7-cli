package engine

import (
	"math/rand"
	"pls7-cli/pkg/poker"
	"testing"
)

// TestActionProvider provides a predefined sequence of actions for testing.
type TestActionProvider struct {
	Actions []PlayerAction
	index   int
}

func (p *TestActionProvider) GetAction(_ *Game, _ *Player, _ *rand.Rand) PlayerAction {
	if p.index < len(p.Actions) {
		action := p.Actions[p.index]
		p.index++
		return action
	}
	return PlayerAction{Type: ActionFold} // Default to fold
}

// newGameForBettingTests creates a game for betting tests.
// If ruleAbbr is omitted, defaults to "PLS".
func newGameForBettingTests(playerNames []string, initialChips int, smallBlind int, bigBlind int, ruleAbbr ...string) *Game {
	abbr := "PLS"
	if len(ruleAbbr) > 0 {
		abbr = ruleAbbr[0]
	}

	rules := &poker.GameRules{
		Abbreviation: abbr,
	}
	switch abbr {
	case "NLH":
		rules.HoleCards = poker.HoleCardRules{Count: 2}
		rules.LowHand = poker.LowHandRules{Enabled: false}
		rules.BettingLimit = "no_limit"
	case "PLS7":
		rules.HoleCards = poker.HoleCardRules{Count: 3}
		rules.LowHand = poker.LowHandRules{Enabled: true, MaxRank: 7}
		rules.BettingLimit = "pot_limit"
	default: // PLS
		rules.HoleCards = poker.HoleCardRules{Count: 3}
		rules.LowHand = poker.LowHandRules{Enabled: false}
		rules.BettingLimit = "pot_limit"
	}
	return NewGame(playerNames, initialChips, smallBlind, bigBlind, DifficultyMedium, rules, true, false, 0)
}

// all players have matched the bet, isBettingActionRequired should return false.
func TestIsBettingActionRequired_MatchedBets_False(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000, "NLH")
	g.StartNewHand()
	// Force a state where all active players have matched the bet
	g.BetToCall = g.BigBlind
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.Status = PlayerStatusPlaying
			p.CurrentBet = g.BigBlind
		}
	}
	if g.isBettingActionRequired() {
		t.Fatalf("expected no further betting required when all bets are matched")
	}
}

// when a player still needs to call, isBettingActionRequired should return true.
func TestIsBettingActionRequired_PlayerNeedsToCall_True(t *testing.T) {
	g := newGameForBettingTests([]string{"YOU", "CPU1", "CPU2"}, 10000, 500, 1000, "NLH")
	g.StartNewHand()
	g.BetToCall = g.BigBlind
	// YOU still needs to call
	g.Players[0].Status = PlayerStatusPlaying
	g.Players[0].CurrentBet = g.SmallBlind
	// Others have matched
	g.Players[1].Status = PlayerStatusPlaying
	g.Players[1].CurrentBet = g.BigBlind
	g.Players[2].Status = PlayerStatusPlaying
	g.Players[2].CurrentBet = g.BigBlind

	if !g.isBettingActionRequired() {
		t.Fatalf("expected betting to be required when a player must still call")
	}
}

func TestIsBettingRoundOver(t *testing.T) {
	t.Run("Round not over - bets not matched", func(t *testing.T) {
		g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 500, 1000, "NLH")
		g.Players[0].CurrentBet = 100
		g.Players[1].CurrentBet = 200
		g.BetToCall = 200
		g.ActionsTakenThisRound = 2 // Assume both players acted
		if g.IsBettingRoundOver() {
			t.Error("Expected betting round to NOT be over")
		}
	})

	t.Run("Round over - all bets matched", func(t *testing.T) {
		g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 500, 1000, "NLH")
		g.Players[0].Status = PlayerStatusPlaying
		g.Players[1].Status = PlayerStatusPlaying
		g.Players[0].CurrentBet = 200
		g.Players[1].CurrentBet = 200
		g.BetToCall = 200
		g.ActionsTakenThisRound = 2 // Manually set that both players have acted
		if !g.IsBettingRoundOver() {
			t.Error("Expected betting round to BE over")
		}
	})

	t.Run("Round over - one player left", func(t *testing.T) {
		g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 500, 1000, "NLH")
		g.Players[0].Status = PlayerStatusPlaying
		g.Players[1].Status = PlayerStatusFolded
		// This should be true regardless of actions taken
		if !g.IsBettingRoundOver() {
			t.Error("Expected betting round to BE over when only one player remains")
		}
	})

	t.Run("Round over - all-in player cannot act on a raise", func(t *testing.T) {
		g := newGameForBettingTests([]string{"YOU", "CPU1"}, 10000, 500, 1000, "NLH")
		g.Players[0].Status = PlayerStatusAllIn
		g.Players[0].CurrentBet = 100
		g.Players[1].Status = PlayerStatusPlaying
		g.Players[1].CurrentBet = 200
		g.BetToCall = 200
		g.ActionsTakenThisRound = 1 // Only the active player needs to have acted
		if !g.IsBettingRoundOver() {
			t.Error("Expected betting round to BE over when a player is all-in and cannot call a raise")
		}
	})
}
