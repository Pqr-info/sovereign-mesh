package routing

import (
	"testing"
)

func TestSlingshotMerge(t *testing.T) {
	smm := NewSlingshotMergeManager()
	bob := NewBlockOrderingBus()

	smm.StartOfflineEpoch("epoch-101", "token-xyz")

	tx := Transaction{
		ID:         "tx-offline-1",
		SequenceID: 5,
		Payload:    "offline-payload",
		Priority:   3,
	}

	bundle := SlingshotBundle{
		BundleID:     "bundle-99",
		EpochID:      "epoch-101",
		Transactions: []Transaction{tx},
		LineageHash:  "PQR-OFFLINE-TEST-HASH",
	}

	// 1. Success case: merge valid bundle during active epoch
	err := smm.MergeBundle(bundle, bob)
	if err != nil {
		t.Fatalf("failed to merge offline bundle: %v", err)
	}

	// Verify transaction was queued in BoB
	processed := bob.ProcessQueue()
	if len(processed) != 1 || processed[0].ID != "tx-offline-1" {
		t.Errorf("expected transaction to be proposed to BoB, got: %v", processed)
	}

	// 2. Reject case: merge against closed epoch
	errClosed := smm.MergeBundle(bundle, bob)
	if errClosed == nil {
		t.Error("expected error when merging against closed epoch, got nil")
	}

	// 3. Reject case: merge unknown epoch ID
	bundleUnknown := SlingshotBundle{
		BundleID:     "bundle-100",
		EpochID:      "unknown-epoch",
		Transactions: []Transaction{tx},
		LineageHash:  "hash",
	}
	errUnknown := smm.MergeBundle(bundleUnknown, bob)
	if errUnknown == nil {
		t.Error("expected error for unknown epoch ID, got nil")
	}
}
