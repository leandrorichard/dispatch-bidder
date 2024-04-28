package dispatchbidder

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// createBidder helps to create a Bidder with less verbosity.
func createBidder(name string, startingBid, maxBid, increment float64) *Bidder {
	return &Bidder{
		ID:            uuid.New(),
		Name:          name,
		StartingBid:   startingBid,
		MaxBid:        maxBid,
		CurrentBid:    startingBid,
		AutoIncrement: increment,
		LastBidTime:   time.Now(),
	}
}

// TestAuctionScenarios tests multiple auction scenarios.
func TestAuctionScenarios(t *testing.T) {
	tests := []struct {
		name         string
		bidders      []*Bidder
		expectedName string
	}{
		{
			name: "Auction #1",
			bidders: []*Bidder{
				createBidder("Sasha", 50.00, 80.00, 3.00),
				createBidder("John", 60.00, 82.00, 2.00),
				createBidder("Pat", 55.00, 85.00, 5.00),
			},
			expectedName: "Pat",
		},
		{
			name: "Auction #2",
			bidders: []*Bidder{
				createBidder("Riley", 700.00, 725.00, 2.00),
				createBidder("Morgan", 599.00, 725.00, 15.00),
				createBidder("Charlie", 625.00, 725.00, 8.00),
			},
			expectedName: "Riley",
		},
		{
			name: "Auction #3",
			bidders: []*Bidder{
				createBidder("Alex", 2500.00, 3000.00, 500.00),
				createBidder("Jesse", 2800.00, 3100.00, 201.00),
				createBidder("Drew", 2501.00, 3200.00, 247.00),
			},
			expectedName: "Jesse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auction, err := NewAuction(NewAuctionConfig{Bidders: tt.bidders})
			assert.NoError(t, err)

			// Simulate multiple rounds of bidding until no more bids can be placed.
			active := true
			for active {
				active = false
				for _, bidder := range tt.bidders {
					// Determine the next possible bid for the current bidder.
					nextBid := bidder.CurrentBid + bidder.AutoIncrement
					if nextBid <= bidder.MaxBid {
						err := auction.PlaceBid(bidder, nextBid)
						if assert.NoError(t, err) {
							active = true // Continue another round if at least one bid was successfully placed.
						}
					}
				}
			}

			winner := auction.DetermineWinner()
			assert.NotNil(t, winner)
			assert.Equal(t, tt.expectedName, winner.Name, "the expected winner does not match.")
		})
	}
}
