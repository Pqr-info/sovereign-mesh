package governance

import (
	"errors"
	"log"
)

type FailureMode string

const (
	FailureConstitutional ViolationType = "Constitutional Violation"
	FailureLineage        FailureMode   = "Lineage Corruption"
	FailureConformation   FailureMode   = "Conformation Collapse"
	FailureMeshPartition  FailureMode   = "Mesh Partition"
	FailureExhaustion     FailureMode   = "Economic Exhaustion"
)

type ViolationType string

// SovereignRecoveryProtocol manages failure recoveries
type SovereignRecoveryProtocol struct {
	Orchestrator *GovernanceOrchestrator
}

func NewSovereignRecoveryProtocol(o *GovernanceOrchestrator) *SovereignRecoveryProtocol {
	return &SovereignRecoveryProtocol{
		Orchestrator: o,
	}
}

// TriggerEmergencyShutdown halts execution loop to preserve node state
func (p *SovereignRecoveryProtocol) TriggerEmergencyShutdown(reason string) {
	log.Printf("[FMRP-27] EMERGENCY SHUTDOWN TRIGGERED: %s", reason)
	p.Orchestrator.MutationGovernor.StabilityModifier = 0.0
}

// ReconcileLineage attempts recovery from conformation/lineage errors
func (p *SovereignRecoveryProtocol) ReconcileLineage(agentID string) error {
	o := p.Orchestrator
	o.mu.Lock()
	defer o.mu.Unlock()

	// Locate agent state and reset parameters
	for _, ag := range o.Agents {
		if ag.AgentID() == agentID {
			ag.Freeze()
			_ = o.JailController.JailAgent(ag, "FMRP Reconcile: Resetting lineage vector due to divergence bounds")
			return nil
		}
	}

	return errors.New("FMRP: Agent not found for lineage reconciliation")
}
