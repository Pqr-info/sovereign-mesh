package routing

import (
	"errors"
	"sort"
	"sync"
)

// Transaction represents a packet proposed to the Block Ordering Bus (BoB)
type Transaction struct {
	ID         string `json:"id"`
	SequenceID uint64 `json:"sequence_id"`
	Payload    string `json:"payload"`
	Priority   int    `json:"priority"`
}

// BlockOrderingBus handles fast-path ordering, sequence bounds, and duplicate check gates (CC-1)
type BlockOrderingBus struct {
	mu           sync.RWMutex
	lastSequence uint64
	seenTxIDs    map[string]bool
	pendingQueue []Transaction
}

func NewBlockOrderingBus() *BlockOrderingBus {
	return &BlockOrderingBus{
		seenTxIDs:    make(map[string]bool),
		pendingQueue: make([]Transaction, 0),
		lastSequence: 0,
	}
}

// ProposeTx adds a transaction, verifying sequence limits and preventing replay attacks
func (bob *BlockOrderingBus) ProposeTx(tx Transaction) error {
	bob.mu.Lock()
	defer bob.mu.Unlock()

	// Replay defense check
	if bob.seenTxIDs[tx.ID] {
		return errors.New("BoB Veto: duplicate transaction ID detected (replay prevention)")
	}

	// Sequence sequence verification
	if tx.SequenceID <= bob.lastSequence {
		return errors.New("BoB Veto: out of order sequence ID")
	}

	bob.seenTxIDs[tx.ID] = true
	bob.pendingQueue = append(bob.pendingQueue, tx)
	return nil
}

// ProcessQueue returns transactions ordered by Priority (highest first) and sets last sequence
func (bob *BlockOrderingBus) ProcessQueue() []Transaction {
	bob.mu.Lock()
	defer bob.mu.Unlock()

	sort.Slice(bob.pendingQueue, func(i, j int) bool {
		if bob.pendingQueue[i].Priority == bob.pendingQueue[j].Priority {
			return bob.pendingQueue[i].SequenceID < bob.pendingQueue[j].SequenceID
		}
		return bob.pendingQueue[i].Priority > bob.pendingQueue[j].Priority
	})

	ordered := make([]Transaction, len(bob.pendingQueue))
	copy(ordered, bob.pendingQueue)

	// Update lastSequence based on processed transactions to enforce monotonic increases in sequences
	for _, tx := range ordered {
		if tx.SequenceID > bob.lastSequence {
			bob.lastSequence = tx.SequenceID
		}
	}

	bob.pendingQueue = make([]Transaction, 0)
	return ordered
}
