package engine

import (
	"fmt"
	"pls7-cli/pkg/poker"

	"github.com/sirupsen/logrus"
)

// playerHoleCardsForDebug is a map used for debugging and testing purposes. It allows
// developers to force specific hole cards for the human player ("YOU") for different
// game variants to test specific scenarios, such as hand evaluation or outs calculation.
var playerHoleCardsForDebug = map[string]map[string]string{
	"PLS7": {
		"3As":        "As Ah Ad", // For testing outs for Four of a Kind
		"AQT-suited": "As Qs Ts", // For testing outs for Flush, Straight, and Skip Straight
		"AAK":        "As Ah Ks", // For testing outs for Three of a Kind
		"A23-suited": "As 2s 3s", // For testing outs for Straight, Flush, and low hand scenarios
	},
	"PLS": {
		"3As":        "As Ah Ad",
		"AQT-suited": "As Qs Ts",
		"AAK":        "As Ah Ks",
		"AKQ-suited": "As Ks Qs",
	},
	"NLH": {
		"AA":        "As Ah",
		"KK":        "Ks Kh",
		"AK-suited": "As Ks",
		"KQ-suited": "Ks Qs",
	},
}

// ProcessAction is a core state-mutating function that updates the game based on a
// single player's action. It handles the logic for folding, checking, calling,
// betting, and raising, and updates the player and game states accordingly.
//
// It returns a boolean indicating if an aggressive action (bet or raise) was taken,
// which is used to track the flow of the betting round, and an ActionEvent for logging.
func (g *Game) ProcessAction(player *Player, action PlayerAction) (wasAggressive bool, event *ActionEvent) {
	g.ActionsTakenThisRound++
	event = &ActionEvent{PlayerName: player.Name, Action: action.Type}

	switch action.Type {
	case ActionFold:
		player.Status = PlayerStatusFolded
		player.LastActionDesc = "Fold"
	case ActionCheck:
		player.LastActionDesc = "Check"
	case ActionCall:
		amountToCall := g.BetToCall - player.CurrentBet
		event.Amount = amountToCall
		g.postBet(player, amountToCall)
		desc := fmt.Sprintf("Call %d", amountToCall)
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
	case ActionBet:
		g.ActionsTakenThisRound = 1 // This player is the new aggressor.
		event.Amount = action.Amount
		g.LastRaiseAmount = action.Amount
		g.postBet(player, action.Amount)
		g.BetToCall = player.CurrentBet
		desc := fmt.Sprintf("Bet %d", action.Amount)
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		g.Aggressor = player
		return true, event
	case ActionRaise:
		g.ActionsTakenThisRound = 1 // This player is the new aggressor.
		event.Amount = action.Amount
		amountToPost := action.Amount - player.CurrentBet
		previousBetToCall := g.BetToCall
		g.postBet(player, amountToPost)
		g.BetToCall = player.CurrentBet
		g.LastRaiseAmount = g.BetToCall - previousBetToCall
		desc := fmt.Sprintf("Raise to %d", action.Amount)
		if player.Status == PlayerStatusAllIn {
			desc += " (All-in)"
		}
		player.LastActionDesc = desc
		g.Aggressor = player
		return true, event
	}
	return false, event
}

// CleanupHand performs post-hand maintenance. It checks for and marks any players
// who have been eliminated (run out of chips) and checks for a game-over condition.
func (g *Game) CleanupHand() []string {
	var events []string
	events = append(events, "\n--- End of Hand ---")
	for _, p := range g.Players {
		if p.Chips == 0 && p.Status != PlayerStatusEliminated {
			p.Status = PlayerStatusEliminated
			events = append(events, fmt.Sprintf("%s has been eliminated!", p.Name))
		}
	}

	// Check if only one player is left in the entire game.
	if g.CountRemainingPlayers() <= 1 {
		for _, p := range g.Players {
			if p.Status != PlayerStatusEliminated {
				events = append(events, fmt.Sprintf("%s wins the game!", p.Name))
				break
			}
		}
	}
	return events
}

// CountRemainingPlayers counts players who have not been eliminated from the game.
// This is used to check for the end-of-game condition.
func (g *Game) CountRemainingPlayers() int {
	count := 0
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			count++
		}
	}
	return count
}

// CountNonFoldedPlayers counts players who are still active in the current hand,
// including those who are all-in.
func (g *Game) CountNonFoldedPlayers() int {
	count := 0
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying || p.Status == PlayerStatusAllIn {
			count++
		}
	}
	return count
}

// CountPlayersAbleToAct counts players who are still able to make betting decisions
// in the current round (i.e., not folded and not all-in).
func (g *Game) CountPlayersAbleToAct() int {
	count := 0
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying {
			count++
		}
	}
	return count
}

// StartNewHand resets the game state to begin a new hand. This involves resetting
// players' statuses and bets, shuffling the deck, moving the dealer button,
// posting blinds, and dealing new hole cards.
func (g *Game) StartNewHand() (event *BlindEvent) {
	g.HandCount++

	// Increase blinds if the blind-up interval has been reached.
	if g.BlindUpInterval > 0 && g.HandCount > 1 && (g.HandCount-1)%g.BlindUpInterval == 0 {
		g.SmallBlind *= 2
		g.BigBlind *= 2
		event = &BlindEvent{SmallBlind: g.SmallBlind, BigBlind: g.BigBlind}
	}

	// Reset game state for the new hand.
	g.Phase = PhasePreFlop
	g.Deck = poker.NewDeck()
	g.Deck.Shuffle(g.Rand)
	g.CommunityCards = []poker.Card{}
	g.Pot = 0
	g.LastRaiseAmount = 0

	g.DealerPos = g.FindNextActivePlayer(g.DealerPos)

	// Reset each player's state for the new hand.
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.Hand = []poker.Card{}
			p.CurrentBet = 0
			p.TotalBetInHand = 0
			p.Status = PlayerStatusPlaying
			p.LastActionDesc = ""
		}
	}

	// Post blinds.
	sbPos := g.FindNextActivePlayer(g.DealerPos)
	bbPos := g.FindNextActivePlayer(sbPos)
	g.postBet(g.Players[sbPos], g.SmallBlind)
	g.postBet(g.Players[bbPos], g.BigBlind)

	g.BetToCall = g.BigBlind
	g.CurrentTurnPos = g.FindNextActivePlayer(bbPos)

	// Deal hole cards.
	// In dev/debug mode, specific cards can be dealt to the human player.
	ruleAbbr := g.Rules.Abbreviation
	if g.DevMode {
		you := g.Players[0]
		if you.Status == PlayerStatusPlaying {
			// Deal specific debug cards to the human player.
			if debugHand, ok := playerHoleCardsForDebug[ruleAbbr]; ok {
				// A default hand from the map is chosen here, e.g., "3As" or "AA".
				// This can be modified for specific testing needs.
				var handStr string
				if ruleAbbr == "NLH" {
					handStr = debugHand["AA"]
				} else {
					handStr = debugHand["3As"]
				}
				playerHoleCards := poker.CardsFromStrings(handStr)
				for _, card := range playerHoleCards {
					dealtCard, err := g.Deck.DealForDebug(card)
					if err == nil {
						you.Hand = append(you.Hand, dealtCard)
					}
				}
			} else {
				logrus.Warnf("Unsupported rule abbreviation for debug hands: %s", ruleAbbr)
			}
		}
		// Deal remaining cards randomly to CPUs.
		for i := 1; i < len(g.Players); i++ {
			for j := 0; j < g.Rules.HoleCards.Count; j++ {
				if g.Players[i].Status == PlayerStatusPlaying {
					card, _ := g.Deck.Deal()
					g.Players[i].Hand = append(g.Players[i].Hand, card)
				}
			}
		}
	} else {
		// In a normal game, deal cards to all players in order.
		for i := 0; i < g.Rules.HoleCards.Count; i++ {
			for pos, p := range g.Players {
				if p.Status == PlayerStatusPlaying {
					card, _ := g.Deck.Deal()
					g.Players[pos].Hand = append(g.Players[pos].Hand, card)
				}
			}
		}
	}

	return event
}

// PlaySingleHand executes one complete hand of poker and returns the result.
// It uses the provided ActionProvider to get player actions and notifies the
// optional GameObserver of state changes during the hand.
// After this method returns, the game is ready for the next hand or inter-hand
// control (save, quit, etc.) by the caller.
func (g *Game) PlaySingleHand(actionProvider ActionProvider, observer GameObserver) *HandResult {
	result := &HandResult{}

	// Start the hand (deals cards, posts blinds, advances dealer).
	result.BlindEvent = g.StartNewHand()

	// Notify observer of initial game state.
	if observer != nil {
		observer.OnPhaseStart(g)
	}

	// Betting rounds loop.
	for g.Phase != PhaseShowdown && g.Phase != PhaseHandOver {
		if g.CountNonFoldedPlayers() <= 1 {
			break
		}
		g.PrepareNewBettingRound()

		// Turn-by-turn betting loop.
		for !g.IsBettingRoundOver() {
			player := g.CurrentPlayer()
			if player.Status != PlayerStatusPlaying {
				g.AdvanceTurn()
				continue
			}

			action := actionProvider.GetAction(g, player, g.Rand)
			_, event := g.ProcessAction(player, action)

			if observer != nil && event != nil {
				observer.OnPlayerAction(event)
			}

			g.AdvanceTurn()
		}
		g.Advance()
	}

	// Conclude the hand.
	if g.CountNonFoldedPlayers() > 1 {
		result.IsShowdown = true
		result.PotResults = g.DistributePot()
	} else {
		result.IsShowdown = false
		result.PotResults = g.AwardPotToLastPlayer()
	}

	result.CleanupMessages = g.CleanupHand()
	return result
}

// FindNextActivePlayer finds the index of the next player at the table who has
// not been eliminated from the game.
func (g *Game) FindNextActivePlayer(startPos int) int {
	pos := (startPos + 1) % len(g.Players)
	for {
		if g.Players[pos].Status != PlayerStatusEliminated {
			return pos
		}
		pos = (pos + 1) % len(g.Players)
	}
}

// postBet is an internal helper function to process a player's bet. It moves chips
// from the player's stack to the pot and updates the player's bet amounts and status.
func (g *Game) postBet(player *Player, amount int) {
	if player.Chips < amount {
		amount = player.Chips // Player is going all-in for less.
	}
	player.Chips -= amount
	player.CurrentBet += amount
	player.TotalBetInHand += amount
	g.Pot += amount
	if player.Chips == 0 {
		player.Status = PlayerStatusAllIn
	}
}

// Advance moves the game state to the next phase (e.g., from Flop to Turn),
// dealing community cards as required.
func (g *Game) Advance() {
	switch g.Phase {
	case PhasePreFlop:
		g.Phase = PhaseFlop
		g.dealCommunityCards(3)
	case PhaseFlop:
		g.Phase = PhaseTurn
		g.dealCommunityCards(1)
	case PhaseTurn:
		g.Phase = PhaseRiver
		g.dealCommunityCards(1)
	case PhaseRiver:
		g.Phase = PhaseShowdown
	case PhaseShowdown:
		g.Phase = PhaseHandOver
	default:
		panic("Undefined game phase in Advance()")
	}
}

// dealCommunityCards deals n cards from the deck to the community cards on the board.
func (g *Game) dealCommunityCards(n int) {
	for i := 0; i < n; i++ {
		card, _ := g.Deck.Deal()
		g.CommunityCards = append(g.CommunityCards, card)
	}
}

// isBettingActionRequired checks if a betting round is necessary. A round can be
// skipped if all but one player is all-in.
func (g *Game) isBettingActionRequired() bool {
	// If there is at least one player who can act and still needs to match a bet,
	// then betting action is required.
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying && p.CurrentBet < g.BetToCall {
			return true
		}
	}
	return false
}

// PrepareNewBettingRound resets the state for the start of a new betting round
// (e.g., after the flop is dealt). It clears players' current bets and determines
// who acts first.
func (g *Game) PrepareNewBettingRound() {
	g.Aggressor = nil
	g.ActionsTakenThisRound = 0

	if g.Phase == PhasePreFlop {
		// Pre-flop is special: blinds are already posted, and action starts after the big blind.
		bbPos := g.FindNextActivePlayer(g.FindNextActivePlayer(g.DealerPos))
		g.ActionCloserPos = bbPos
		return
	}

	// For post-flop rounds, reset bets and start with the first active player after the dealer.
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated {
			p.CurrentBet = 0
			p.LastActionDesc = ""
		}
	}
	g.BetToCall = 0
	g.LastRaiseAmount = 0
	g.CurrentTurnPos = g.FindNextActivePlayer(g.DealerPos)
	g.ActionCloserPos = g.FindPreviousActivePlayer(g.CurrentTurnPos)
}

// FindPreviousActivePlayer finds the index of the previous player at the table
// who has not been eliminated from the game.
func (g *Game) FindPreviousActivePlayer(startPos int) int {
	pos := (startPos - 1 + len(g.Players)) % len(g.Players)
	for {
		if g.Players[pos].Status != PlayerStatusEliminated {
			return pos
		}
		pos = (pos - 1 + len(g.Players)) % len(g.Players)
	}
}

// IsBettingRoundOver determines if the current betting round has concluded.
// A round ends when all active players have had a turn and all bets have been matched.
func (g *Game) IsBettingRoundOver() bool {
	// Round is over if only one or zero players can act.
	if g.CountNonFoldedPlayers() <= 1 {
		return true
	}

	// All players who are able to act must have taken an action.
	if g.ActionsTakenThisRound < g.CountPlayersAbleToAct() {
		return false
	}

	// All players still in the hand must have matching bets (or be all-in).
	for _, p := range g.Players {
		if p.Status == PlayerStatusPlaying {
			if p.CurrentBet < g.BetToCall {
				return false // A player still needs to call a bet.
			}
		}
	}

	return true
}

// CurrentPlayer returns a pointer to the player whose turn it is to act.
func (g *Game) CurrentPlayer() *Player {
	return g.Players[g.CurrentTurnPos]
}

// AdvanceTurn moves the action to the next active player in the hand.
func (g *Game) AdvanceTurn() {
	g.CurrentTurnPos = g.FindNextActivePlayer(g.CurrentTurnPos)
}
