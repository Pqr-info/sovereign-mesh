package governance

import (
	"testing"
)

func TestConstitutionalGate(t *testing.T) {
	csl := NewConstitutionalSafetyLayer()
	cg := NewConstitutionalGate(csl)

	tenant := &Tenant{
		TenantID:               "test-const-tenant",
		State:                  Active,
		ComputeCyclesRemaining: 10,
	}

	// Use an actual concrete agent type that implements Agent
	agent := NewGameTheoryAgent("test-const-agent", 3, "US-EAST")

	// 1. Nominal case passing all gates
	etNominal := EthicalTensor27{
		HarmAxes:        [9]float64{0.1, 0.2, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		IntegrityAxes:   [9]float64{},
		SovereigntyAxes: [9]float64{},
	}
	if err := cg.EvaluateAction(tenant, agent, etNominal); err != nil {
		t.Fatalf("expected action to pass constitutional gate, got: %v", err)
	}

	// 2. BL-3 boundary violation (suspended tenant)
	tenant.State = Suspended
	if err := cg.EvaluateAction(tenant, agent, etNominal); err == nil {
		t.Error("expected error for suspended tenant, got nil")
	}
	tenant.State = Active // reset

	// 3. ET-27 violation (high harm axis value)
	etHarm := EthicalTensor27{
		HarmAxes:        [9]float64{0.9, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		IntegrityAxes:   [9]float64{},
		SovereigntyAxes: [9]float64{},
	}
	if err := cg.EvaluateAction(tenant, agent, etHarm); err == nil {
		t.Error("expected error for high harm axis, got nil")
	}

	if len(csl.ActiveViolations) == 0 {
		t.Error("expected safety layer to record violations")
	}
}
