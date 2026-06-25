package routing

import (
	"testing"
)

func TestBlockOrderingBus(t *testing.T) {
	bob := NewBlockOrderingBus()

	tx1 := Transaction{
		ID:         "tx-1",
		SequenceID: 10,
		Payload:    "payload-1",
		Priority:   5,
	}

	tx2 := Transaction{
		ID:         "tx-2",
		SequenceID: 12,
		Payload:    "payload-2",
		Priority:   10,
	}

	// 1. Propose nominal transactions
	if err := bob.ProposeTx(tx1); err != nil {
		t.Fatalf("failed to propose tx1: %v", err)
	}
	if err := bob.ProposeTx(tx2); err != nil {
		t.Fatalf("failed to propose tx2: %v", err)
	}

	// 2. Assert double-spend / replay prevention
	if err := bob.ProposeTx(tx1); err == nil {
		t.Error("expected error for duplicate transaction propose, got nil")
	}

	// 3. Process first batch to progress lastSequence
	processed := bob.ProcessQueue()
	if len(processed) != 2 {
		t.Fatalf("expected 2 processed transactions, got %d", len(processed))
	}

	if processed[0].ID != "tx-2" {
		t.Errorf("expected tx-2 (higher priority) to sort first, got: %s", processed[0].ID)
	}

	// 4. Assert sequence bounds enforcement on subsequent proposals
	txInvalidSeq := Transaction{
		ID:         "tx-3",
		SequenceID: 8, // sequence ID 8 is <= bob.lastSequence (which is now 12)
		Payload:    "invalid-seq",
		Priority:   1,
	}
	if err := bob.ProposeTx(txInvalidSeq); err == nil {
		t.Error("expected error for out of order sequence ID, got nil")
	}
}
