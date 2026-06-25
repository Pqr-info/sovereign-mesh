package routing

import (
	"errors"
	"time"
)

// SlingshotEpoch represents an offline epoch boundary (MS-1)
type SlingshotEpoch struct {
	EpochID   string    `json:"epoch_id"`
	Token     string    `json:"token"`
	StartTime time.Time `json:"start_time"`
	Closed    bool      `json:"closed"`
}

// SlingshotBundle encapsulates offline transactions and metadata accumulated during an epoch (MS-2)
type SlingshotBundle struct {
	BundleID     string        `json:"bundle_id"`
	EpochID      string        `json:"epoch_id"`
	Transactions []Transaction `json:"transactions"`
	LineageHash  string        `json:"lineage_hash"`
}

// SlingshotMergeManager executes conflict-free deterministic merges for offline epochs
type SlingshotMergeManager struct {
	ActiveEpochs map[string]SlingshotEpoch
}

func NewSlingshotMergeManager() *SlingshotMergeManager {
	return &SlingshotMergeManager{
		ActiveEpochs: make(map[string]SlingshotEpoch),
	}
}

// StartOfflineEpoch registers a new offline epoch token
func (smm *SlingshotMergeManager) StartOfflineEpoch(epochID string, token string) {
	smm.ActiveEpochs[epochID] = SlingshotEpoch{
		EpochID:   epochID,
		Token:     token,
		StartTime: time.Now().UTC(),
		Closed:    false,
	}
}

// MergeBundle merges an offline bundle if the associated epoch is valid and open
func (smm *SlingshotMergeManager) MergeBundle(bundle SlingshotBundle, bob *BlockOrderingBus) error {
	epoch, exists := smm.ActiveEpochs[bundle.EpochID]
	if !exists {
		return errors.New("Slingshot Veto: unknown epoch ID")
	}
	if epoch.Closed {
		return errors.New("Slingshot Veto: epoch is already closed")
	}

	// Stream and propose transactions into BlockOrderingBus
	for _, tx := range bundle.Transactions {
		err := bob.ProposeTx(tx)
		if err != nil {
			return err
		}
	}

	// Close the epoch post-merge
	epoch.Closed = true
	smm.ActiveEpochs[bundle.EpochID] = epoch

	return nil
}
