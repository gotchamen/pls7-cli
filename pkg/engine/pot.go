package engine

import (
	"fmt"
	"pls7-cli/pkg/poker"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// DistributionResult is a data structure that holds the outcome of a pot
// distribution for a single player. It's used to communicate the results
// back to the UI or logger.
type DistributionResult struct {
	PlayerName string // The name of the player who won a share of the pot.
	AmountWon  int    // The total amount of chips won by the player.
	HandDesc   string // A description of the winning hand (e.g., "High: Flush", "Low: 8-7-6-5-4").
}

// PotTier represents a single pot (either the main pot or a side pot) that is
// created when one or more players are all-in. Each tier has a specific amount
// and a list of players who are eligible to win it.
type PotTier struct {
	Amount  int       // The total chip amount in this specific pot tier.
	Players []*Player // The slice of players who are eligible to win this pot tier.
	MaxBet  int       // The maximum bet amount that players in this tier have contributed.
}

// AwardPotToLastPlayer handles the simple scenario where all but one player have
// folded. The remaining player wins the entire pot without a showdown.
func (g *Game) AwardPotToLastPlayer() []DistributionResult {
	var winner *Player
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded && p.Status != PlayerStatusEliminated {
			winner = p
			break
		}
	}

	if winner != nil {
		winner.Chips += g.Pot
		result := DistributionResult{
			PlayerName: winner.Name,
			AmountWon:  g.Pot,
			HandDesc:   "takes the pot as the last remaining player",
		}
		g.Pot = 0
		return []DistributionResult{result}
	}
	return []DistributionResult{}
}

// DistributePot is the core function for calculating and awarding the pot(s) at
// the end of a hand. It correctly handles complex scenarios including multiple
// side pots for all-in players and High-Low split pots.
func (g *Game) DistributePot() []DistributionResult {
	showdownPlayers := g.getShowdownPlayers()
	if len(showdownPlayers) == 0 {
		return nil
	}

	pots := g.buildPotTiers(showdownPlayers)

	winnerChipMap := make(map[string]int)
	winnerHandDescMap := make(map[string]string)

	for _, pot := range pots {
		g.distributeSinglePotTier(pot, winnerChipMap, winnerHandDescMap)
	}

	// Aggregate the winnings into the final result list.
	var results []DistributionResult
	for name, amount := range winnerChipMap {
		results = append(results, DistributionResult{
			PlayerName: name,
			AmountWon:  amount,
			HandDesc:   winnerHandDescMap[name],
		})
	}

	g.Pot = 0
	logrus.Debugf("DistributePot: Final results: %+v", results)
	return results
}

// buildPotTiers creates the main pot and side pots based on all-in bet tiers.
func (g *Game) buildPotTiers(showdownPlayers []*Player) []PotTier {
	// Collect all contributors (including folded players who bet).
	var allContributors []*Player
	for _, p := range g.Players {
		if p.Status != PlayerStatusEliminated && p.TotalBetInHand > 0 {
			allContributors = append(allContributors, p)
		}
	}

	// Create sorted unique bet tiers.
	betTierSet := make(map[int]bool)
	for _, p := range allContributors {
		betTierSet[p.TotalBetInHand] = true
	}
	var sortedTiers []int
	for bet := range betTierSet {
		sortedTiers = append(sortedTiers, bet)
	}
	sort.Ints(sortedTiers)

	logrus.Debugf("DistributePot: Initial Pot: %d, All Contributors: %v, Bet Tiers: %v", g.Pot, getPlayerNames(allContributors), sortedTiers)

	var pots []PotTier
	lastBet := 0

	for _, tierBet := range sortedTiers {
		contribution := tierBet - lastBet
		if contribution <= 0 {
			continue
		}

		numPlayersInTier := 0
		for _, p := range allContributors {
			if p.TotalBetInHand >= tierBet {
				numPlayersInTier++
			}
		}
		tierAmount := contribution * numPlayersInTier

		var eligiblePlayers []*Player
		for _, sp := range showdownPlayers {
			if sp.TotalBetInHand >= tierBet {
				eligiblePlayers = append(eligiblePlayers, sp)
			}
		}

		if tierAmount > 0 && len(eligiblePlayers) > 0 {
			pots = append(pots, PotTier{
				Amount:  tierAmount,
				Players: eligiblePlayers,
				MaxBet:  tierBet,
			})
			logrus.Debugf("  New PotTier created: Amount: %d, MaxBet: %d, Players: %v",
				tierAmount, tierBet, getPlayerNames(eligiblePlayers))
			if len(eligiblePlayers) == 1 {
				logrus.Warnf("  Single player %s eligible for PotTier with amount %d", eligiblePlayers[0].Name, tierAmount)
			}
		}
		lastBet = tierBet
	}
	return pots
}

// distributeSinglePotTier distributes a single pot tier to the winner(s),
// handling Hi-Lo splits when applicable.
func (g *Game) distributeSinglePotTier(pot PotTier, winnerChipMap map[string]int, winnerHandDescMap map[string]string) {
	logrus.Debugf("Distributing PotTier: Amount: %d, MaxBet: %d, Eligible Players: %v", pot.Amount, pot.MaxBet, getPlayerNames(pot.Players))

	highWinners, bestHighHand := findBestHighHand(pot.Players, g)
	lowWinners, bestLowHand := findBestLowHand(pot.Players, g)
	logrus.Debugf("DistributePot: High Winners: %v, Best High Hand: %s", getPlayerNames(highWinners), bestHighHand)
	logrus.Debugf("DistributePot: Low Winners: %v, Best Low Hand: %s", getPlayerNames(lowWinners), bestLowHand)

	if g.Rules.LowHand.Enabled && len(lowWinners) > 0 {
		g.distributeHiLoPot(pot, highWinners, bestHighHand, lowWinners, bestLowHand, winnerChipMap, winnerHandDescMap)
	} else {
		distributeHighOnlyPot(pot, highWinners, bestHighHand, winnerChipMap, winnerHandDescMap)
	}
}

// distributeHiLoPot splits a pot tier between high and low winners.
func (g *Game) distributeHiLoPot(pot PotTier, highWinners []*Player, bestHighHand *poker.HandResult, lowWinners []*Player, bestLowHand *poker.HandResult, winnerChipMap map[string]int, winnerHandDescMap map[string]string) {
	lowPot := pot.Amount / 2
	highPot := pot.Amount - lowPot
	logrus.Debugf("  Split Pot: lowPot: %d, highPot: %d", lowPot, highPot)

	// Distribute the low half.
	lowShare := lowPot / len(lowWinners)
	var lowHandRanks []string
	for _, c := range bestLowHand.Cards {
		lowHandRanks = append(lowHandRanks, c.Rank.String())
	}
	if len(lowHandRanks) > 0 && lowHandRanks[0] == poker.Ace.String() {
		lowHandRanks = append(lowHandRanks[1:], lowHandRanks[0])
	}
	lowHandDesc := fmt.Sprintf("Low: %s-High", strings.Join(lowHandRanks, "-"))

	for _, winner := range lowWinners {
		winner.Chips += lowShare
		winnerChipMap[winner.Name] += lowShare
		winnerHandDescMap[winner.Name] = lowHandDesc
		logrus.Debugf("    %s wins %d from low pot", winner.Name, lowShare)
	}

	// Distribute the high half.
	highShare := highPot / len(highWinners)
	highHandDesc := fmt.Sprintf("High: %s", bestHighHand.String())
	for _, winner := range highWinners {
		winner.Chips += highShare
		winnerChipMap[winner.Name] += highShare
		if desc, exists := winnerHandDescMap[winner.Name]; exists && strings.HasPrefix(desc, "Low") {
			winnerHandDescMap[winner.Name] = fmt.Sprintf("Scoop! %s, %s", highHandDesc, desc)
		} else {
			winnerHandDescMap[winner.Name] = highHandDesc
		}
		logrus.Debugf("    %s wins %d from high pot", winner.Name, highShare)
	}
}

// distributeHighOnlyPot awards the entire pot tier to the high hand winner(s).
func distributeHighOnlyPot(pot PotTier, highWinners []*Player, bestHighHand *poker.HandResult, winnerChipMap map[string]int, winnerHandDescMap map[string]string) {
	highShare := pot.Amount / len(highWinners)
	highHandDesc := fmt.Sprintf("High: %s (Scoop)", bestHighHand.String())
	for _, winner := range highWinners {
		winner.Chips += highShare
		winnerChipMap[winner.Name] += highShare
		winnerHandDescMap[winner.Name] = highHandDesc
		logrus.Debugf("    %s scoops %d from pot", winner.Name, highShare)
	}
}

// getShowdownPlayers returns a slice of players who are still active in the
// hand and thus eligible to participate in the showdown.
func (g *Game) getShowdownPlayers() []*Player {
	var active []*Player
	for _, p := range g.Players {
		if p.Status != PlayerStatusFolded && p.Status != PlayerStatusEliminated {
			active = append(active, p)
		}
	}
	return active
}

// findBestHighHand iterates through a list of players and determines who has the
// best high hand according to the game's rules. It returns the winning player(s)
// (in case of a tie) and the best hand result.
func findBestHighHand(players []*Player, g *Game) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		highHand, _ := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
		if highHand == nil {
			continue
		}
		if bestHand == nil || compareHandResults(highHand, bestHand) == 1 {
			bestHand = highHand
			winners = []*Player{p}
		} else if compareHandResults(highHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

// findBestLowHand iterates through a list of players and determines who has the
// best qualifying low hand. It returns the winning player(s) and the best low hand.
// If no player has a qualifying low hand, it returns nil.
func findBestLowHand(players []*Player, g *Game) (winners []*Player, bestHand *poker.HandResult) {
	for _, p := range players {
		_, lowHand := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
		if lowHand == nil {
			continue
		}
		// For low hands, a lower result is better.
		if bestHand == nil || compareHandResults(lowHand, bestHand) == -1 {
			bestHand = lowHand
			winners = []*Player{p}
		} else if compareHandResults(lowHand, bestHand) == 0 {
			winners = append(winners, p)
		}
	}
	return
}

// compareHandResults compares two hand results to determine which is stronger.
// It first compares by HandRank, then by HighValues for tie-breaking.
// Returns 1 if h1 > h2, -1 if h1 < h2, 0 if h1 == h2.
func compareHandResults(h1, h2 *poker.HandResult) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}
	// Ranks are the same, compare kickers.
	for i := 0; i < len(h1.HighValues); i++ {
		if h1.HighValues[i] > h2.HighValues[i] {
			return 1
		}
		if h1.HighValues[i] < h2.HighValues[i] {
			return -1
		}
	}
	return 0 // Hands are identical.
}

// getPlayerNames is a helper function for logging, returning a slice of player names.
func getPlayerNames(players []*Player) []string {
	names := make([]string, len(players))
	for i, p := range players {
		names[i] = p.Name
	}
	return names
}
