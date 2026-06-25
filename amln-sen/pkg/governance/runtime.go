package governance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type RuntimeAction struct {
	Type          string          `json:"type"`
	Payload       interface{}     `json:"payload"`
	EthicalTensor EthicalTensor27 `json:"ethical_tensor"`
}

// RuntimeExecutionEngine manages cycle-by-cycle constitutional execution traces
type RuntimeExecutionEngine struct {
	mu           sync.Mutex
	Orchestrator *GovernanceOrchestrator
}

func NewRuntimeExecutionEngine(o *GovernanceOrchestrator) *RuntimeExecutionEngine {
	return &RuntimeExecutionEngine{
		Orchestrator: o,
	}
}

// ExecuteCycle runs a proposed action through the triadic execution cascade
func (ree *RuntimeExecutionEngine) ExecuteCycle(
	ctx context.Context,
	tenantID string,
	agent Agent,
	action RuntimeAction,
) error {
	ree.mu.Lock()
	defer ree.mu.Unlock()

	o := ree.Orchestrator

	// Load tenant
	o.TenantManager.Mu.Lock()
	tenant, exists := o.TenantManager.Tenants[tenantID]
	o.TenantManager.Mu.Unlock()
	if !exists {
		return errors.New("Runtime Veto: Tenant not found")
	}

	// 1. Gate 1: Constitutional Gate (CG-3)
	// Check compute cycle and balance bounds
	if tenant.Plan != PlanFree && tenant.ComputeCyclesRemaining <= 0 {
		tenant.Suspend()
		return fmt.Errorf("CG-3 Veto: Compute cycles exhausted (%s)", string(ViolationCompute))
	}
	if tenant.MemoryWindowExpiration.Before(time.Now().UTC()) {
		tenant.Suspend()
		return fmt.Errorf("CG-3 Veto: Memory window expired (%s)", string(ViolationCompute))
	}

	// 2. Gate 2: Governance Gate (GG-9)
	err := o.CSL.ValidateAction(tenant, agent, action.EthicalTensor)
	if err != nil {
		return fmt.Errorf("GG-9 Veto: %w", err)
	}

	// 3. Gate 3: Ethical Tensor Gate (ET-27)
	// Group evaluation triads logic (Harm axes, integrity axes, sovereignty axes check)
	for i, val := range action.EthicalTensor.IntegrityAxes {
		if val > 0.90 {
			return fmt.Errorf("ET-27 Veto: Integrity axes breach on axis %d", i)
		}
	}
	for i, val := range action.EthicalTensor.SovereigntyAxes {
		if val > 0.90 {
			return fmt.Errorf("ET-27 Veto: Sovereignty axes breach on axis %d", i)
		}
	}

	// 4. go27 cycle decrement
	if !tenant.ConsumeCycle() {
		return errors.New("GME Veto: Metering cycle decrement failed")
	}

	// 5. PQR Logging audit event
	log.Printf("[PQR-LOG] Action committed. Tenant: %s, Agent: %s, ActionType: %s, Cycles Remaining: %d",
		tenantID, agent.AgentID(), action.Type, tenant.ComputeCyclesRemaining)

	return nil
}
