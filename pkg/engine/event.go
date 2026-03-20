package engine

// ActionEvent represents a significant action taken by a player during a betting
// round. It is intended to be used for logging, display, or broadcasting game
// state changes to observers like a UI.
type ActionEvent struct {
	// PlayerName is the name of the player who performed the action.
	PlayerName string
	// Action is the type of action taken (e.g., Fold, Call, Raise).
	Action ActionType
	// Amount is the value associated with the action, such as the size of a
	// bet or raise. It is 0 for actions like Fold and Check.
	Amount int
}

// BlindEvent represents the posting of the small and big blinds at the beginning
// of a hand. It can be used to announce the current blind levels.
type BlindEvent struct {
	// SmallBlind is the size of the small blind.
	SmallBlind int
	// BigBlind is the size of the big blind.
	BigBlind int
}

// GameObserver receives notifications during hand execution.
// Implementations can use these to update UI, log events, etc.
// A nil GameObserver is valid and means no notifications are sent.
type GameObserver interface {
	// OnPhaseStart is called at the beginning of the hand and after each phase transition.
	OnPhaseStart(g *Game)
	// OnPlayerAction is called after a player takes an action during a betting round.
	OnPlayerAction(event *ActionEvent)
}

// HandResult contains the outcome of a completed hand.
type HandResult struct {
	// BlindEvent is non-nil if blinds increased at the start of this hand.
	BlindEvent *BlindEvent
	// IsShowdown is true if the hand went to showdown (multiple players remained).
	IsShowdown bool
	// PotResults contains the pot distribution results for all winners.
	PotResults []DistributionResult
	// CleanupMessages contains post-hand messages (e.g., player eliminations).
	CleanupMessages []string
}
