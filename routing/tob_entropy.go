package routing

import (
	"errors"
	"math"
)

// EntropyAuction represents a Lighthouse Transaction Ordering Bus (ToB) auction round (CC-2)
type EntropyAuction struct {
	BidderNodeID   string  `json:"bidder_node_id"`
	BidAmount      float64 `json:"bid_amount"`
	ProposedEntropy float64 `json:"proposed_entropy"`
}

// TransactionOrderingBus manages ToB Lighthouse auctions and conformation angles (CC-2)
type TransactionOrderingBus struct {
	ActiveBids []EntropyAuction
}

func NewTransactionOrderingBus() *TransactionOrderingBus {
	return &TransactionOrderingBus{
		ActiveBids: make([]EntropyAuction, 0),
	}
}

// SubmitBid logs an active auction proposal
func (tob *TransactionOrderingBus) SubmitBid(bid EntropyAuction) {
	tob.ActiveBids = append(tob.ActiveBids, bid)
}

// ResolveAuction calculates the winner, resolution epsilon, and target conformation theta
func (tob *TransactionOrderingBus) ResolveAuction() (winner EntropyAuction, epsilon float64, theta float64, err error) {
	if len(tob.ActiveBids) == 0 {
		return EntropyAuction{}, 0, 0, errors.New("ToB Veto: no active bids in entropy auction")
	}

	// Winner has the highest bid amount
	highestBidIdx := 0
	totalBids := 0.0
	for i, bid := range tob.ActiveBids {
		totalBids += bid.BidAmount
		if bid.BidAmount > tob.ActiveBids[highestBidIdx].BidAmount {
			highestBidIdx = i
		}
	}

	winner = tob.ActiveBids[highestBidIdx]

	// ToB Logic: Epsilon (resolution) scales with total bids energy
	epsilon = math.Tanh(totalBids)

	// Conformation Theta resolves as a normalized function of the winning proposed entropy
	theta = math.Sin(winner.ProposedEntropy)

	// Flush active bids for next loop
	tob.ActiveBids = make([]EntropyAuction, 0)

	return winner, epsilon, theta, nil
}
