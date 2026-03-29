package cli

import (
	"bufio"
	"fmt"
	"os"
	"pls7-cli/pkg/engine"
	"strconv"
	"strings"
)

// PromptForAction requests the player to choose an action during their turn.
func PromptForAction(g *engine.Game) engine.PlayerAction {
	DisplayGameState(g)

	// for loop to keep prompting until a valid action is chosen
	for {
		player := g.Players[g.CurrentTurnPos]
		canCheck := player.CurrentBet == g.BetToCall
		amountToCall := g.BetToCall - player.CurrentBet

		var prompt strings.Builder
		prompt.WriteString("Choose your action: ")

		raisingMeaningless := g.IsRaisingMeaningless()

		if canCheck {
			if raisingMeaningless {
				prompt.WriteString("chec(k), (f)old > ")
			} else {
				prompt.WriteString("chec(k), (b)et, (f)old > ")
			}
		} else {
			// If amountToCall is negative, it means remaining players have bet all-in with less than the current bet.
			// So the player does not need to act anything, call.
			if amountToCall < 0 {
				return engine.PlayerAction{Type: engine.ActionCall}
			}

			prompt.WriteString(fmt.Sprintf("(c)all %s, ", FormatNumber(amountToCall)))
			// Only show raise option if the player has enough chips and raising is meaningful.
			if !raisingMeaningless {
				minRaise, _ := g.CalculateBettingLimits()
				if player.Chips > amountToCall && player.CurrentBet+player.Chips >= minRaise {
					prompt.WriteString("(r)aise, ")
				}
			}
			prompt.WriteString("(f)old > ")
		}

		fmt.Print(prompt.String())
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "f":
			return engine.PlayerAction{Type: engine.ActionFold}
		case "k":
			if canCheck {
				return engine.PlayerAction{Type: engine.ActionCheck}
			}
		case "c":
			if !canCheck {
				return engine.PlayerAction{Type: engine.ActionCall}
			}
		case "b":
			if canCheck && !raisingMeaningless {
				return promptForAmount(g, engine.ActionBet)
			}
		case "r":
			if !canCheck && !raisingMeaningless {
				return promptForAmount(g, engine.ActionRaise)
			}
		}

		fmt.Println("Invalid action.")
	}
}

// promptForAmount requests the betting/raising amount.
func promptForAmount(g *engine.Game, actionType engine.ActionType) engine.PlayerAction {
	for {
		minBet, maxBet := g.CalculateBettingLimits()
		actionName := "bet"
		if actionType == engine.ActionRaise {
			actionName = "raise to"
		}

		fmt.Printf(
			"Enter amount to %s (min: %s, max: %s): ",
			actionName, FormatNumber(minBet), FormatNumber(maxBet),
		)

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		amount, err := strconv.Atoi(strings.TrimSpace(input))

		if err != nil || amount < minBet || amount > maxBet {
			fmt.Println("Invalid amount. Please try again.")
		} else {
			return engine.PlayerAction{Type: actionType, Amount: amount}
		}
	}
}
