package governance

import (
	"context"
	"testing"
	"time"
)

func TestCycleScheduler(t *testing.T) {
	// Setup Orchestrator, Tenant & REE
	o := NewGovernanceOrchestrator(3, 1*time.Second)
	tenant, err := o.TenantManager.CreateTenant("test-sched-tenant", "addr12345678901234567890123456789012345678901234567890123456789012345678901234567", PlanStandard)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	tenant.PurchaseGo27(0.81) // grants 9 compute cycles

	ree := NewRuntimeExecutionEngine(o)
	cs := NewCycleScheduler(ree)

	agent := o.Agents[0]
	action := RuntimeAction{
		Type:    "STRATEGY_CYCLE",
		Payload: "payload",
		EthicalTensor: EthicalTensor27{
			HarmAxes:        [9]float64{},
			IntegrityAxes:   [9]float64{},
			SovereigntyAxes: [9]float64{},
		},
	}

	cs.QueueAction(ScheduledAction{
		TenantID: "test-sched-tenant",
		Agent:    agent,
		Action:   action,
	})

	// 1. Run cycle step
	ctx := context.Background()
	err = cs.RunCycleStep(ctx)
	if err != nil {
		t.Fatalf("scheduler execution rejected: %v", err)
	}

	if cs.TotalBurned != 1 {
		t.Errorf("expected TotalBurned to be 1, got %d", cs.TotalBurned)
	}

	// 2. Verify cycle was consumed
	if tenant.ComputeCyclesRemaining != 8 {
		t.Errorf("expected remaining cycles to be 8, got: %d", tenant.ComputeCyclesRemaining)
	}
}
