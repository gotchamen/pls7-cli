package cmd

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"pls7-cli/internal/cli"
	"pls7-cli/internal/config"
	"pls7-cli/internal/util"
	"pls7-cli/pkg/engine"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ruleStr         string // To hold the --rule flag value (load rules/{rule}.yml when the game starts)
	difficultyStr   string // To hold the flag value
	devMode         bool   // To hold the --dev flag value
	showOuts        bool   // To hold the --outs flag value (this does not work if devMode is true, as it will always show outs in dev mode)
	blindUpInterval int    // To hold the --blind-up flag value
	initialChips    int    // To hold the --initial-chips flag value
	smallBlind      int    // To hold the --small-blind flag value
	loadGame        bool   // To hold the --load flag value (load saved game)
	loadFile        string // To hold the --load-file flag value (specific filename to load)
	saveDir         string // To hold the --save-dir flag value (directory for save files)
)

// CLIActionProvider implements the ActionProvider interface using the CLI.
type CLIActionProvider struct{}

func (p *CLIActionProvider) GetAction(g *engine.Game, _ *engine.Player, _ *rand.Rand) engine.PlayerAction {
	return cli.PromptForAction(g)
}

// CPUActionProvider implements the ActionProvider interface for CPU players.
type CPUActionProvider struct{}

func (p *CPUActionProvider) GetAction(g *engine.Game, pl *engine.Player, r *rand.Rand) engine.PlayerAction {
	return g.GetCPUAction(pl, r)
}

// CombinedActionProvider decides which provider to use based on player type.
type CombinedActionProvider struct{}

// GetAction method for CombinedActionProvider
func (p *CombinedActionProvider) GetAction(g *engine.Game, player *engine.Player, r *rand.Rand) engine.PlayerAction {
	if player.IsCPU {
		time.Sleep(g.CPUThinkTime())
		return g.GetCPUAction(player, r)
	}
	return cli.PromptForAction(g)
}

func runGame(cmd *cobra.Command, _ []string) {
	util.InitLogger(devMode)

	var g *engine.Game
	var err error

	// Check if we should load a saved game (--load flag was specified)
	if loadGame {
		// Set default save directory if not specified
		if saveDir == "" {
			saveDir = "saves"
		}

		// If no specific filename provided, load the most recent save file
		if loadFile == "" {
			fmt.Printf("Loading most recent saved game...\n")
			g, err = engine.LoadGameFromFile(saveDir, "")
		} else {
			fmt.Printf("Loading saved game from %s...\n", loadFile)
			g, err = engine.LoadGameFromFile(saveDir, loadFile)
		}
		if err != nil {
			fmt.Printf("❌ Failed to load saved game: %v\n", err)
			fmt.Printf("💡 Make sure you have saved games in the '%s' directory.\n", saveDir)
			fmt.Printf("💡 You can start a new game by running: go run main.go\n")
			os.Exit(1)
		}

		fmt.Printf("Game loaded successfully! Starting new hand with Hand #%d\n",
			g.HandCount+1) // Show next hand number since we're starting fresh
		fmt.Printf("Players: %d, Total chips in play: %s\n",
			len(g.Players), cli.FormatNumber(g.TotalInitialChips))
	} else {
		// Create new game
		// Load game rules
		rules, err := config.LoadGameRulesFromOptions(ruleStr)
		if err != nil {
			logrus.Fatalf("Failed to load game rules: %v", err)
		}

		fmt.Printf("======== %s ========\n", rules.Name)

		playerNames := []string{"YOU", "CPU 1", "CPU 2", "CPU 3", "CPU 4", "CPU 5"}

		var difficulty engine.Difficulty
		switch difficultyStr {
		case "easy":
			difficulty = engine.DifficultyEasy
		case "medium":
			difficulty = engine.DifficultyMedium
		case "hard":
			difficulty = engine.DifficultyHard
		default:
			logrus.Warnf("Invalid difficulty '%s' specified. Defaulting to medium.", difficultyStr)
			difficulty = engine.DifficultyMedium
		}

		g = engine.NewGame(playerNames, initialChips, smallBlind, smallBlind*2, difficulty, rules, devMode, showOuts, blindUpInterval)
	}

	actionProvider := &CombinedActionProvider{}

	// Main Game Loop (multi-hand)
	for {
		// Always start a new hand - loaded games are ready to start fresh
		blindEvent := g.StartNewHand()
		if blindEvent != nil {
			message := fmt.Sprintf("\n*** Blinds are now %s/%s ***\n", cli.FormatNumber(blindEvent.SmallBlind), cli.FormatNumber(blindEvent.BigBlind))
			fmt.Println(message)
		}
		// Clear the loadFile flag after starting the first hand
		loadFile = ""

		cli.DisplayGameState(g)

		// Single Hand Loop
		for g.Phase != engine.PhaseShowdown && g.Phase != engine.PhaseHandOver {
			if g.CountNonFoldedPlayers() <= 1 {
				break
			}
			g.PrepareNewBettingRound()

			// New Turn-by-turn Betting Loop
			for !g.IsBettingRoundOver() {
				player := g.CurrentPlayer()
				var action engine.PlayerAction

				if player.Status != engine.PlayerStatusPlaying {
					g.AdvanceTurn()
					continue
				}

				action = actionProvider.GetAction(g, player, g.Rand)

				_, event := g.ProcessAction(player, action)
				if event != nil {
					var eventMessage string
					switch event.Action {
					case engine.ActionFold:
						eventMessage = fmt.Sprintf("%s folds.", event.PlayerName)
					case engine.ActionCheck:
						eventMessage = fmt.Sprintf("%s checks.", event.PlayerName)
					case engine.ActionCall:
						eventMessage = fmt.Sprintf("%s calls %s.", event.PlayerName, cli.FormatNumber(event.Amount))
					case engine.ActionBet:
						eventMessage = fmt.Sprintf("%s bets %s.", event.PlayerName, cli.FormatNumber(event.Amount))
					case engine.ActionRaise:
						eventMessage = fmt.Sprintf("%s raises to %s.", event.PlayerName, cli.FormatNumber(event.Amount))
					}
					if eventMessage != "" {
						fmt.Println(eventMessage)
					}
				}
				g.AdvanceTurn()
			}
			g.Advance()
		}

		// Conclude the hand
		if g.CountNonFoldedPlayers() > 1 {
			showdownMessages := cli.FormatShowdownResults(g)
			for _, msg := range showdownMessages {
				fmt.Println(msg)
			}
		} else {
			results := g.AwardPotToLastPlayer()
			fmt.Println("--- POT AWARDED ---")
			for _, result := range results {
				fmt.Printf(
					"%s wins %s chips with %s\n",
					result.PlayerName, cli.FormatNumber(result.AmountWon), result.HandDesc,
				)
			}
			fmt.Println("------------------------")
		}

		cleanupMessages := g.CleanupHand()
		for _, msg := range cleanupMessages {
			fmt.Println(msg)
		}

		if g.Players[0].Status == engine.PlayerStatusEliminated {
			fmt.Println("You have been eliminated. GAME OVER.")
			break
		}

		if g.CountRemainingPlayers() <= 1 {
			fmt.Println("--- GAME OVER ---")
			break
		}

		fmt.Print("Press ENTER to start the next hand, type 's' to save, or type 'q' to exit > ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q":
			fmt.Println("Thanks for playing!")
			return
		case "s":
			if saveDir == "" {
				saveDir = "saves"
			}

			// Generate timestamp-based filename automatically
			saveFilename := fmt.Sprintf("save_%s", time.Now().Format("20060102_150405"))

			err := engine.SaveGameToFile(g, saveDir, saveFilename)
			if err != nil {
				fmt.Printf("❌ Failed to save game: %v\n", err)
				fmt.Print("Press ENTER to continue...")
				reader.ReadString('\n')
			} else {
				fmt.Printf("✅ Game saved successfully as %s.json\n", saveFilename)
				fmt.Printf("🔄 You can load this game later with: go run main.go --load\n")
				fmt.Print("Press ENTER to continue...")
				reader.ReadString('\n')
			}
			continue
		default:
			// Continue to next hand
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pls7",
	Short: "Starts a new game of Poker",
	Long:  `Starts a new game of Poker (PLS7, PLS, NLH) with 1 player and 5 CPUs.`, // Corrected escaping for backticks and quotes within the string literal. The original string was fine.
	Run:   runGame,
}

// savesCmd represents the saves subcommand
var savesCmd = &cobra.Command{
	Use:   "saves",
	Short: "Manage saved games",
	Long:  `List, validate, or delete saved game files.`,
}

// listCmd represents the list subcommand
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved games",
	Long:  `List all saved games in the save directory with their metadata.`,
	Run:   listSaves,
}

// validateCmd represents the validate subcommand
var validateCmd = &cobra.Command{
	Use:   "validate [filename]",
	Short: "Validate a saved game file",
	Long:  `Validate a saved game file to check if it can be loaded properly.`,
	Args:  cobra.ExactArgs(1),
	Run:   validateSave,
}

// deleteCmd represents the delete subcommand
var deleteCmd = &cobra.Command{
	Use:   "delete [filename]",
	Short: "Delete a saved game file",
	Long:  `Delete a saved game file from the save directory.`,
	Args:  cobra.ExactArgs(1),
	Run:   deleteSave,
}

// listSaves lists all saved games
func listSaves(_ *cobra.Command, _ []string) {
	saves, err := engine.ListSaveFiles(saveDir)
	if err != nil {
		logrus.Fatalf("Failed to list saves: %v", err)
	}

	if len(saves) == 0 {
		fmt.Printf("No saved games found in directory: %s\n", saveDir)
		return
	}

	fmt.Printf("Saved games in %s:\n", saveDir)
	fmt.Println("==========================================")
	for i, save := range saves {
		fmt.Printf("%d. %s\n", i+1, save.Filename)
		fmt.Printf("   Created: %s\n", save.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Size: %d bytes\n", save.Size)
		if save.GameMetadata != nil {
			fmt.Printf("   Hand: #%d\n", save.GameMetadata.HandCount)
			fmt.Printf("   Blinds: %s/%s\n", cli.FormatNumber(save.GameMetadata.SmallBlind), cli.FormatNumber(save.GameMetadata.BigBlind))
		}
		fmt.Println()
	}
}

// validateSave validates a saved game file
func validateSave(_ *cobra.Command, args []string) {
	filename := args[0]

	err := engine.ValidateSaveFile(saveDir, filename)
	if err != nil {
		fmt.Printf("❌ Save file '%s' is invalid: %v\n", filename, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Save file '%s' is valid and can be loaded.\n", filename)
}

// deleteSave deletes a saved game file
func deleteSave(_ *cobra.Command, args []string) {
	filename := args[0]

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete '%s'? (y/N): ", filename)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return
	}

	err := engine.DeleteSaveFile(saveDir, filename)
	if err != nil {
		fmt.Printf("Failed to delete save file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Save file '%s' deleted successfully.\n", filename)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Add subcommands
	rootCmd.AddCommand(savesCmd)
	savesCmd.AddCommand(listCmd)
	savesCmd.AddCommand(validateCmd)
	savesCmd.AddCommand(deleteCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&ruleStr, "rule", "r", "pls7", "Game rule to use (pls7, pls, nlh).")
	rootCmd.Flags().StringVarP(&difficultyStr, "difficulty", "d", "medium", "Set AI difficulty (easy, medium, hard)")
	rootCmd.Flags().BoolVar(&devMode, "dev", false, "Enable development mode for verbose logging.")
	rootCmd.Flags().BoolVar(&showOuts, "outs", false, "Shows outs for players if found (temporarily draws fixed good hole cards).")
	rootCmd.Flags().IntVar(&blindUpInterval, "blind-up", 2, "Sets the number of rounds for blind up. 0 means no blind up.")
	rootCmd.Flags().IntVar(&initialChips, "initial-chips", 300000, "Initial chips for each player.")
	rootCmd.Flags().IntVar(&smallBlind, "small-blind", 500, "Small blind amount.")
	rootCmd.Flags().BoolVarP(&loadGame, "load", "l", false, "Load the most recent saved game.")
	rootCmd.Flags().StringVar(&loadFile, "load-file", "", "Load a specific saved game file.")
	rootCmd.Flags().StringVar(&saveDir, "save-dir", "saves", "Directory to store save files.")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if initialChips <= 0 {
			return fmt.Errorf("initial-chips는 0보다 커야 합니다. 입력값: %d", initialChips)
		}
		if smallBlind <= 0 {
			return fmt.Errorf("small-blind는 0보다 커야 합니다. 입력값: %d", smallBlind)
		}
		return nil
	}
}
