package poker

import (
	"fmt"
	"math/rand"
	"time"
)

// Deck represents a collection of playing cards.
type Deck struct {
	// Cards is a slice holding the cards in the deck.
	Cards []Card
}

// NewDeck creates a new, unshuffled, standard 52-card deck.
// It contains all combinations of suits (Spade, Heart, Diamond, Club) and
// ranks (Two through Ace).
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for suit := Spade; suit <= Club; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	return &Deck{Cards: cards}
}

// Shuffle randomizes the order of the cards in the deck.
// It uses the provided rand.Rand source to ensure deterministic shuffling for
// testing purposes. For production use, a cryptographically secure random
// source should be used.
func (d *Deck) Shuffle(r *rand.Rand) {
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Deal removes and returns the top card from the deck (the last card in the slice).
// It returns an error if the deck is empty.
func (d *Deck) Deal() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, fmt.Errorf("deck is empty")
	}
	card := d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]
	return card, nil
}

// DealForDebug removes and returns a specific card from the deck.
// This function is intended for testing and debugging purposes to control the
// game state by dealing known cards. It searches for the card in the deck,
// removes it if found, and returns it. If the card is not found, it returns
// an error.
func (d *Deck) DealForDebug(card Card) (Card, error) {
	for i, c := range d.Cards {
		if c == card {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			return c, nil
		}
	}
	return Card{}, fmt.Errorf("card %s not found in deck", card)
}

// NewRand creates a new random number generator with the specified seed.
// This function is used for deterministic game state recreation when loading saved games.
func NewRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// NewRandWithTime creates a new random number generator with a time-based seed.
// This function is used for normal gameplay with unpredictable randomness.
func NewRandWithTime() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
