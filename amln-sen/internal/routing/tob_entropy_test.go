package routing

import (
	"testing"
)

func TestTransactionOrderingBus(t *testing.T) {
	tob := NewTransactionOrderingBus()

	// Propose multiple auction bids
	tob.SubmitBid(EntropyAuction{
		BidderNodeID:    "node-alpha",
		BidAmount:       1.2,
		ProposedEntropy: 0.5,
	})
	tob.SubmitBid(EntropyAuction{
		BidderNodeID:    "node-beta",
		BidAmount:       4.5, // High bidder
		ProposedEntropy: 1.0,
	})
	tob.SubmitBid(EntropyAuction{
		BidderNodeID:    "node-gamma",
		BidAmount:       0.8,
		ProposedEntropy: 0.2,
	})

	winner, epsilon, theta, err := tob.ResolveAuction()
	if err != nil {
		t.Fatalf("failed to resolve Lighthouse ToB auction: %v", err)
	}

	if winner.BidderNodeID != "node-beta" {
		t.Errorf("expected node-beta to win auction, got: %s", winner.BidderNodeID)
	}

	if epsilon <= 0.0 || epsilon >= 1.0 {
		t.Errorf("expected normalized epsilon in (0, 1) range, got: %f", epsilon)
	}

	if theta == 0.0 {
		t.Error("expected non-zero target conformation theta")
	}

	// Verify flush
	if len(tob.ActiveBids) != 0 {
		t.Errorf("expected bids to be flushed post-auction, got: %d", len(tob.ActiveBids))
	}
}
