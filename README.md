# pls7-cli

A simple CLI for Pot Limit Sampyong 7 or Better (PLS7) Poker

## What is PLS7?

Pot Limit Sampyong 7 or Better (PLS7, or Sampyong Hi-Lo) is a variant of poker that combines elements of traditional poker with unique rules and gameplay mechanics. It is played with a standard deck of cards and involves betting, bluffing, and strategic decision-making.

- [Guide - English](https://philipjkim.github.io/posts/20250729-pls7-english-guide/)
- [Guide - Korean](https://philipjkim.github.io/posts/20250724-sampyeong-holdem-guide-v1-4/)

## Installation

This guide will walk you through setting up the Go environment and the project itself.

### 1. Go Language Installation

You need Go version 1.23 or higher to run this application.

#### For macOS Users

The easiest way to install Go on a Mac is by using [Homebrew](https://brew.sh/).

1.  If you don't have Homebrew, open your Terminal and install it with the following command:
    ```bash
    /bin/bash -c "$(curl -fsSL [https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh](https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh))"
    ```
2.  Once Homebrew is installed, install Go with this simple command:
    ```bash
    brew install go
    ```
3.  Verify the installation by checking the version:
    ```bash
    go version
    ```

Alternatively, you can download the official installer from the [Go download page](https://go.dev/dl/).

#### For Windows Users

The recommended way to install Go on Windows is by using the official MSI installer.

1.  Visit the [Go download page](https://go.dev/dl/) and download the MSI installer for Windows.
2.  Run the downloaded installer file. The setup wizard will guide you through the installation process.
3.  The installer will automatically add the Go binary to your system's PATH environment variable.
4.  To verify the installation, open a new Command Prompt or PowerShell window and type:
    ```bash
    go version
    ```

### 2. Project Setup

Once Go is installed on your system, follow these steps to set up the project.

1.  Open your terminal or command prompt.
2.  Clone the repository to your local machine (replace the URL with the actual repository URL):
    ```bash
    git clone [https://github.com/your-username/pls7-cli.git](https://github.com/your-username/pls7-cli.git)
    ```
3.  Navigate into the newly created project directory:
    ```bash
    cd pls7-cli
    ```
4.  Download the necessary dependencies listed in the project:
    ```bash
    go mod tidy
    ```

That's it! You are now ready to run the application.

## Running the App

You can run the application using `go run main.go` with various flags to customize the game.

### Usage

```bash
go run main.go [flags]
```

### Flags

The application accepts the following flags:

| Flag, Short      | Type     | Default  | Description                                                                 |
| ---------------- | -------- | -------- | --------------------------------------------------------------------------- |
| `--rule`, `-r`   | `string` | `"pls7"` | Game rule to use. Corresponds to a file in the `/rules` directory (e.g., `pls7`, `pls`, `nlh`). |
| `--difficulty`, `-d` | `string` | `"medium"` | AI difficulty (`easy`, `medium`, `hard`).                                   |
| `--blind-up`     | `int`    | `2`      | The number of hands for blinds to increase. `0` disables blind-ups.         |
| `--dev`          | `bool`   | `false`  | Enables development mode for verbose logging.                               |
| `--outs`         | `bool`   | `false`  | Shows hand outs for the human player.                                       |
| `--load`, `-l`   | `bool`   | `false`  | Load the most recent saved game.                                            |
| `--load-file`    | `string` | `""`     | Load a specific saved game file.                                            |
| `--save-dir`     | `string` | `"saves"`| Directory to store save files.                                             |
| `--initial-chips`| `int`    | `300000` | Initial chips for each player.                                              |
| `--small-blind`  | `int`    | `500`    | Small blind amount. Big blind automatically set to twice times of it.                                                         |
| `--help`, `-h`   | `bool`   | `false`  | Shows the help message.                                                       |

### Examples

```bash
# Start a standard PLS7 game with medium AI
go run main.go

# Start a PLS (Pot-Limit Sampyeong) game
go run main.go --rule pls

# Start a No-Limit Hold'em (NLH) game with easy AI and show outs
go run main.go -r nlh -d easy --outs

# Load a saved game (most recent)
go run main.go --load

# Load a specific saved game
go run main.go --load-file my_save

# Run in development mode for detailed logs
go run main.go --dev

# Start a game with custom settings
go run main.go --initial-chips 500000 --small-blind 1000
```

### Save/Load Commands

The application also provides subcommands for managing saved games:

```bash
# List all saved games
go run main.go saves list

# Validate a save file
go run main.go saves validate my_save

# Delete a save file
go run main.go saves delete my_save
```

### Game Controls

During gameplay, you can:
- Press `ENTER` to continue to the next hand
- Type `s` to save the current game state with an auto-generated timestamp filename (available during betting rounds and between hands)
- Type `q` to quit the game

#### Betting Actions
During your turn, you can choose from:
- `f` - Fold (forfeit your hand)
- `c` - Call (match the current bet)
- `r` - Raise (increase the bet)
- `k` - Check (pass without betting, when no bet is required)
- `b` - Bet (make the first bet in a round)
- `s` - Save (save the current game state)

## Creating an Executable

```bash
go build -o pls7 main.go
```

## Testing

```bash
# Simple test
go test ./...

# To run all tests in the project with verbose output
go test -v ./...
```

## 📖 Documentation

- [Architecture (EN)](./docs/architecture.md)
- [Architecture (KO)](./docs/architecture_ko.md)
- [Directory Structure (EN)](./docs/directory_structure.md)
- [Directory Structure (KO)](./docs/directory_structure_ko.md)
- [Development Plan (KO)](./docs/development_plan.md)
- [Project Roadmap (KO)](./docs/roadmap_v20250827.md)

## Contributing

This project is being actively developed with extensive use of AI tools such as GEMINI CLI, Claude Code, Codex, etc.

For basic development rules, please refer to the guidelines in [docs/GEMINI_en.md](./docs/GEMINI_en.md). For Korean, refer to [docs/GEMINI.md](./docs/GEMINI.md).
