package cli

import (
	"fmt"
	"pls7-cli/pkg/engine"
	"pls7-cli/pkg/poker"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// DisplayGameState prints the current state of the game board and players.
func DisplayGameState(g *engine.Game) {
	if !g.DevMode {
		clearScreen()
	}

	var output string // Concat all output here and print at once not to be mixed with other logs

	phaseName := strings.ToUpper(g.Phase.String())
	output += fmt.Sprintf("\n\n--- %s (%s) | HAND #%d | PHASE: %s | POT: %s | BLINDS: %s/%s ---\n",
		g.Rules.Abbreviation, g.Difficulty, g.HandCount, phaseName,
		FormatNumber(g.Pot), FormatNumber(g.SmallBlind), FormatNumber(g.BigBlind),
	)

	var communityCardStrings []string
	for _, c := range g.CommunityCards {
		communityCardStrings = append(communityCardStrings, c.String())
	}
	output += fmt.Sprintf("Board: %s\n\n", strings.Join(communityCardStrings, " "))

	totalChips := g.Pot
	output += fmt.Sprintln("Players:")
	for i, p := range g.Players {
		// --- NEW: Skip eliminated players from the display ---
		if p.Status == engine.PlayerStatusEliminated {
			continue
		}
		// --- END OF NEW PART ---

		indicator := "  "
		if i == g.DealerPos {
			indicator = "D "
		}
		if i == g.CurrentTurnPos {
			indicator = "> "
		}

		status := ""
		if p.Status == engine.PlayerStatusFolded {
			status = "(Folded)"
		}
		if p.Status == engine.PlayerStatusAllIn && p.CurrentBet == 0 {
			status = "(All In)"
		}

		handInfo := ""
		if !p.IsCPU || g.DevMode {
			var handStrings []string
			for _, c := range p.Hand {
				handStrings = append(handStrings, c.String())
			}
			handInfo = fmt.Sprintf("| Hand: %s", strings.Join(handStrings, " "))

			if g.Phase > engine.PhasePreFlop {
				highRank, lowRank := poker.EvaluateHand(p.Hand, g.CommunityCards, g.Rules)
				rankInfo := fmt.Sprintf(" | High: %s", highRank.String())
				if g.Rules.LowHand.Enabled && lowRank != nil {
					rankInfo += fmt.Sprintf(", Low: %s", lowRank.String())
				}
				handInfo += rankInfo
			}
		}

		actionInfo := ""
		if p.Status != engine.PlayerStatusEliminated {
			actionInfo = fmt.Sprintf(", Current Bet: %-6s", FormatNumber(p.CurrentBet))
			if p.LastActionDesc != "" && i != g.CurrentTurnPos {
				actionInfo += fmt.Sprintf(" - %s", p.LastActionDesc)
			}
		}
		logrus.Debugf(
			"Player %s: Status: %v, Current Bet: %s, Last Action: %s, actionInfo: [%s]",
			p.Name, p.Status, FormatNumber(p.CurrentBet), p.LastActionDesc, actionInfo,
		)

		nameInfo := fmt.Sprintf("%s%s", indicator, p.Name)
		if p.IsCPU && g.DevMode {
			nameInfo = fmt.Sprintf("%s%s (%s)", indicator, p.Name, p.Profile.Name)
		}
		line := fmt.Sprintf("% -30s: Chips: %-9s%s %s %s", nameInfo, FormatNumber(p.Chips), actionInfo, status, handInfo)
		output += fmt.Sprintln(strings.TrimSpace(line))

		// Display outs for the player in dev mode
		if g.CanShowOuts(p) {
			hasOuts, outsInfo := poker.CalculateOuts(p.Hand, g.CommunityCards, g.Rules)
			if hasOuts {
				sort.Slice(outsInfo.AllOuts, func(i, j int) bool {
					if outsInfo.AllOuts[i].Suit != outsInfo.AllOuts[j].Suit {
						return outsInfo.AllOuts[i].Suit < outsInfo.AllOuts[j].Suit
					}
					return outsInfo.AllOuts[i].Rank < outsInfo.AllOuts[j].Rank
				})
				output += formatOuts(outsInfo)

				amountToCall := g.BetToCall - p.CurrentBet
				output += formatEquities(g.Pot, amountToCall, len(outsInfo.AllOuts), g.Phase)
			}
		}

		// Calculate total chips for the game
		if p.Status != engine.PlayerStatusEliminated {
			totalChips += p.Chips
		}
	}

	if totalChips != g.TotalInitialChips {
		logrus.Warnf(
			"Total chips mismatch: expected %s, got %s",
			FormatNumber(g.TotalInitialChips),
			FormatNumber(totalChips),
		)
	} else {
		logrus.Debugf(
			"Total chips match expected value: %s",
			FormatNumber(g.TotalInitialChips),
		)
	}

	output += fmt.Sprintln("-------------------------------------------------")
	fmt.Print(output)
}

// formatOuts formats the outs cards for display.
func formatOuts(outsInfo *poker.OutsInfo) string {
	result := "\tAll Outs: "
	outStrings := make([]string, 0, len(outsInfo.AllOuts))
	for _, c := range outsInfo.AllOuts {
		outStrings = append(outStrings, c.String())
	}
	result += strings.Join(outStrings, ", ")

	if outsInfo.OutsPerHandRank != nil {
		result += "\n\tOuts by Hand Rank:\n"
		for rank, outs := range outsInfo.OutsPerHandRank {
			if len(outs) == 0 {
				continue
			}
			possibleHandRankStr := rank.String()
			if rank.String() == poker.HighCard.String() {
				possibleHandRankStr = possibleHandRankStr + " (for low hand)"
			}
			result += fmt.Sprintf("\t\t%s: ", possibleHandRankStr)
			outRankStrings := make([]string, 0, len(outs))
			for _, c := range outs {
				outRankStrings = append(outRankStrings, c.String())
			}
			result += strings.Join(outRankStrings, ", ") + "\n"
		}
	}
	return result
}

func formatEquities(pot, amountToCall, numOuts int, phase engine.GamePhase) string {
	numCommunityCards := 0
	if phase == engine.PhaseFlop {
		numCommunityCards = 3
	} else if phase == engine.PhaseTurn {
		numCommunityCards = 4
	} else {
		// For Pre-Flop and River, we don't calculate outs or equities
		return ""
	}

	return fmt.Sprintf("\n\t- Break-even equity based on pot odds: %.2f\n\t- Equity: %.2f\n",
		poker.CalculateBreakEvenEquityBasedOnPotOdds(pot, amountToCall),
		poker.CalculateEquity(numCommunityCards, numOuts),
	)
}

// clearScreen clears the console. (Note: This is a simple implementation)
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func FormatShowdownResults(g *engine.Game, distributionResults []engine.DistributionResult) []string {
	var outputLines []string
	outputLines = append(outputLines, "\n--- SHOWDOWN ---")
	outputLines = append(outputLines, fmt.Sprintf("Community Cards: %s", g.CommunityCards))

	winnerMap := make(map[string][]string)
	for _, result := range distributionResults {
		winType := ""
		if strings.HasPrefix(result.HandDesc, "High") || strings.HasPrefix(result.HandDesc, "takes") {
			winType = "High Winner"
		} else if strings.HasPrefix(result.HandDesc, "Scoop") {
			winType = "High/Low Winner"
		} else {
			winType = "Low Winner"
		}
		winnerMap[result.PlayerName] = append(winnerMap[result.PlayerName], winType)
	}

	for _, player := range g.Players {
		if player.Status == engine.PlayerStatusFolded || player.Status == engine.PlayerStatusEliminated {
			continue
		}
		highHand, lowHand := poker.EvaluateHand(player.Hand, g.CommunityCards, g.Rules)

		handDesc := highHand.String()
		if g.Rules.LowHand.Enabled && lowHand != nil {
			var lowHandRanks []string
			for _, c := range lowHand.Cards {
				lowHandRanks = append(lowHandRanks, c.Rank.String())
			}
			if len(lowHandRanks) > 0 && lowHandRanks[0] == "A" {
				lowHandRanks = append(lowHandRanks[1:], lowHandRanks[0])
			}
			handDesc += fmt.Sprintf(" | Low: %s-High", strings.Join(lowHandRanks, "-"))
		}

		winnerStatus := ""
		if statuses, ok := winnerMap[player.Name]; ok {
			winnerStatus = fmt.Sprintf(" (%s)", strings.Join(statuses, " & "))
		}

		outputLines = append(outputLines, fmt.Sprintf("- %-7s: %v -> %s%s", player.Name, player.Hand, handDesc, winnerStatus))
	}

	outputLines = append(outputLines, "\n--- POT DISTRIBUTION ---")
	for _, result := range distributionResults {
		outputLines = append(outputLines, fmt.Sprintf(
			"%s wins %s chips with %s",
			result.PlayerName, FormatNumber(result.AmountWon), result.HandDesc,
		))
	}
	outputLines = append(outputLines, "------------------------")
	return outputLines
}
