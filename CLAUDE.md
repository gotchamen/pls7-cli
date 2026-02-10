# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pls7-cli is a CLI poker game engine supporting multiple poker variants: PLS7 (Pot-Limit Sampyeong 7-or-Better), PLS, PLO, PLO8, and NLH (No-Limit Hold'em). It features AI opponents with configurable difficulty, save/load functionality, and configurable game rules via YAML files.

## Build & Test Commands

```bash
go build -v ./...                  # Build all packages
go test ./...                      # Run all tests
go test -v ./...                   # Run all tests with verbose output
go test -v ./pkg/engine/...        # Run tests for a specific package
go test -v -run TestFuncName ./pkg/engine/  # Run a single test
go run main.go                     # Run the game (default: PLS7, medium difficulty)
go run main.go -r nlh -d easy      # Run NLH variant with easy AI
```

CI runs `go build -v ./...` and `go test -v ./...` on Go 1.23 (see `.github/workflows/go.yml`).

## Architecture

Three-layer architecture with strict dependency direction: CLI -> Engine -> Poker Library.

- **`pkg/poker`** - Pure poker library (zero internal dependencies). Cards, decks, hand evaluation, odds calculation, and `GameRules` struct. Supports custom hand rankings (e.g., skip straights in PLS7) and Hi-Lo evaluation.
- **`pkg/engine`** - Game state machine. Manages `Game` struct (players, pot, phases, betting rounds, turn order), AI decision-making via `AIProfile` system, save/load, and pot distribution (including side pots and Hi-Lo splits). Uses `BettingLimitCalculator` interface (Strategy Pattern) with `PotLimitCalculator` and `NoLimitCalculator`.
- **`cmd/root.go`** - Cobra CLI entry point. Orchestrates the main game loop: hand start -> betting rounds -> showdown -> cleanup. Contains `CombinedActionProvider` that routes human input through `internal/cli` and CPU actions through `engine.GetCPUAction`.
- **`internal/cli`** - Display (`display.go`) and user input (`input.go`) for the terminal UI. Formatting helpers in `format.go`.
- **`internal/config`** - Loads `rules/*.yml` files into `poker.GameRules` structs.
- **`rules/*.yml`** - YAML game variant definitions (hole card count, use constraints, betting limits, custom rankings, low hand rules).

### Key Game Flow

`Game.StartNewHand()` -> `PrepareNewBettingRound()` -> turn loop with `IsBettingRoundOver()`/`ProcessAction()`/`AdvanceTurn()` -> `Advance()` phase transitions -> `DistributePot()` at showdown -> `CleanupHand()`.

Game phases progress: PreFlop -> Flop -> Turn -> River -> Showdown -> HandOver.

### AI System

CPU players are assigned `AIProfile`s (Tight-Aggressive, Loose-Aggressive, Tight-Passive, Loose-Passive) based on difficulty. Profiles control play/raise thresholds, bluff frequency, and aggression factor. `handEvaluator` function is injectable for testing.

## Development Rules

Detailed development rules are in the `/pls7-dev-rules` skill. Key points:

- **TDD required**: Test first -> fail -> implement -> pass. No exceptions.
- **Never delete logs**: Existing `logrus` statements are untouchable.
- **English code/comments**: All code comments and documentation in English.
- **Test player naming**: `[]string{"YOU", "CPU1", "CPU2"}` format.
- **Fail-fast**: Stop on 3rd failure, report to user.

## Skills

- `/pls7-commit` — Analyze changes and execute git commit with conventional commit message (English).
- `/pls7-pr-message` — Generate PR title + description (English first, then Korean) as raw markdown.
- `/pls7-dev-rules` — Full development rules: TDD, logging, test conventions, coding standards.
- `/pls7-work-log` — Generate a work log in `docs/issue-history/` summarizing the current session.
