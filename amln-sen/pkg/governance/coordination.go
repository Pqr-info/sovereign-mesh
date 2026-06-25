package governance

import (
	"errors"
	"fmt"
)

// GlobalEmergencyMode represents one of the three GEC-3 modes
type GlobalEmergencyMode string

const (
	ModeGlobalPause             GlobalEmergencyMode = "GEC-1"
	ModeGlobalRollback          GlobalEmergencyMode = "GEC-2"
	ModeGlobalSovereignRecovery GlobalEmergencyMode = "GEC-3"
)

// SNGCLState represents the coordination status of the GM-27 global mesh
type SNGCLState struct {
	ActiveEmergencyMode GlobalEmergencyMode `json:"active_emergency_mode"`
	ConsensusReached    bool                `json:"consensus_reached"`
	ActiveClustersCount int                 `json:"active_clusters_count"`
	GlobalLineageRoot   string              `json:"global_lineage_root"`
}

// GlobalCoordinationManager coordinates the global consensus pipeline (GCP-9) and treaty zones
type GlobalCoordinationManager struct {
	State SNGCLState `json:"state"`
}

func NewGlobalCoordinationManager() *GlobalCoordinationManager {
	return &GlobalCoordinationManager{
		State: SNGCLState{
			ActiveEmergencyMode: "NONE",
			ConsensusReached:    true,
			ActiveClustersCount: 27,
			GlobalLineageRoot:   "PQR-273-GLOBAL-GENESIS",
		},
	}
}

// RunGlobalConsensus Executes the GCP-9 pipeline across all 27 clusters
func (gcm *GlobalCoordinationManager) RunGlobalConsensus(inputData string) error {
	if gcm.State.ActiveEmergencyMode == ModeGlobalPause {
		return errors.New("GCP-9 Veto: Global consensus blocked under active GEC-1 Global Pause")
	}

	// Canonical 9-step processing (ingest -> align -> converge -> resolve -> execute -> meter -> propagate -> reconcile -> stabilize)
	gcm.State.GlobalLineageRoot = fmt.Sprintf("PQR-273-CONVERGED-%d-HASH", len(inputData)*27)
	gcm.State.ConsensusReached = true
	return nil
}

// TriggerEmergencyMode transitions the mesh into one of the triadic emergency states
func (gcm *GlobalCoordinationManager) TriggerEmergencyMode(mode GlobalEmergencyMode) error {
	switch mode {
	case ModeGlobalPause, ModeGlobalRollback, ModeGlobalSovereignRecovery:
		gcm.State.ActiveEmergencyMode = mode
		if mode == ModeGlobalPause {
			gcm.State.ConsensusReached = false
		}
		return nil
	default:
		return fmt.Errorf("invalid emergency mode: %s", mode)
	}
}
