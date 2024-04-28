package dispatchbidder

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Bidder represents an individual participant in an auction.
type Bidder struct {
	ID            uuid.UUID
	Name          string
	StartingBid   float64
	MaxBid        float64
	CurrentBid    float64
	AutoIncrement float64
	LastBidTime   time.Time
}

// Auction holds all the details of a single auction event.
type Auction struct {
	sync.RWMutex
	ID      uuid.UUID
	Bidders []*Bidder
}

// NewAuctionConfig is used to configure a new auction.
type NewAuctionConfig struct {
	Bidders []*Bidder
}

// NewAuction creates a new auction instance from the given parameters.
func NewAuction(na NewAuctionConfig) (*Auction, error) {
	if err := validateAuctionData(na); err != nil {
		return nil, fmt.Errorf("invalid auction data: %w", err)
	}

	auction := Auction{
		ID:      uuid.New(),
		Bidders: na.Bidders,
	}

	return &auction, nil
}

// PlaceBid places a bid on the auction.
func (a *Auction) PlaceBid(bidder *Bidder, bidAmount float64) error {
	a.Lock()
	defer a.Unlock()

	// -----------------------------------------------------------------------
	// Perform validations.

	if bidAmount < bidder.StartingBid {
		return fmt.Errorf("bid amount $%.2f is less than starting bid $%.2f", bidAmount, bidder.StartingBid)
	}
	if bidAmount > bidder.MaxBid {
		return fmt.Errorf("bid amount $%.2f is greater than max bid $%.2f", bidAmount, bidder.MaxBid)
	}
	if bidAmount <= bidder.CurrentBid {
		return fmt.Errorf("bid amount $%.2f is less than or equal to current bid $%.2f", bidAmount, bidder.CurrentBid)
	}

	// -----------------------------------------------------------------------
	// Updates the bidder current bid.

	bidder.CurrentBid = bidAmount
	bidder.LastBidTime = time.Now()

	// -----------------------------------------------------------------------
	// For all other bidders, increment their current bid by their respective
	// AutoIncrement amount, provided this does not exceed their MaxBid.

	for _, otherBidder := range a.Bidders {
		if otherBidder.ID != bidder.ID {
			newBid := otherBidder.CurrentBid + otherBidder.AutoIncrement
			if newBid <= otherBidder.MaxBid {
				otherBidder.CurrentBid = newBid
				otherBidder.LastBidTime = time.Now()
			}
		}
	}

	return nil
}

// DetermineWinner determines the winner of the auction based on the highest current bid.
// In case of a tie (multiple bidders with the same highest bid), the bidder who placed
// their bid first (based on LastBidTime) is considered the winner.
func (a *Auction) DetermineWinner() *Bidder {
	var winner *Bidder

	for _, bidder := range a.Bidders {
		if isWinner(winner, bidder) {
			winner = bidder
		}
	}

	return winner
}

// isWinner checks if the provided bidder should replace the current winner.
// A bidder becomes the new winner if:
// - There is no current winner.
// - Their bid is higher than the current winner's bid.
// - Their bid is the same as the current winner's but was placed earlier.
func isWinner(currentWinner, bidder *Bidder) bool {
	return currentWinner == nil || // No current winner, so the bidder wins by default.
		bidder.CurrentBid > currentWinner.CurrentBid || // Bidder has a higher bid.
		(bidder.CurrentBid == currentWinner.CurrentBid && // Bidder has the same bid but placed it earlier.
			bidder.LastBidTime.Before(currentWinner.LastBidTime))
}

// validateAuctionData checks that the provided data for a new auction is valid.
func validateAuctionData(na NewAuctionConfig) error {
	if len(na.Bidders) <= 1 {
		return errors.New("auction must have at least two bidders")
	}

	seenIDs := make(map[uuid.UUID]bool)
	for _, bidder := range na.Bidders {
		// -----------------------------------------------------------------------
		// Check for unique IDs to prevent duplicate bidders.

		if _, exists := seenIDs[bidder.ID]; exists {
			return fmt.Errorf("duplicate bidder ID detected: %s", bidder.ID)
		}
		seenIDs[bidder.ID] = true

		// -----------------------------------------------------------------------
		// Validate individual bidder data.

		if err := validateBidder(bidder); err != nil {
			return fmt.Errorf("invalid bidder data for bidder ID %s: %w", bidder.ID, err)
		}
	}
	return nil
}

// validateBidder checks that a bidder's data is valid.
func validateBidder(b *Bidder) error {
	if b.StartingBid <= 0 {
		return fmt.Errorf("starting bid must be positive, got $%.2f", b.StartingBid)
	}
	if b.MaxBid < b.StartingBid {
		return fmt.Errorf("max bid $%.2f must be greater than or equal to starting bid $%.2f", b.MaxBid, b.StartingBid)
	}
	if b.AutoIncrement <= 0 {
		return fmt.Errorf("auto-increment must be positive, got $%.2f", b.AutoIncrement)
	}
	return nil
}
