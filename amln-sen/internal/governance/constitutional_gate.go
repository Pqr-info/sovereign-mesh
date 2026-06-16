package governance

import (
	"errors"
	"fmt"
	"time"
)

// ConstitutionalGate handles baseline BL-3 boundary evaluations, GL-9 invariant safety, and ET-27 overrides (RE-2)
type ConstitutionalGate struct {
	CSL *ConstitutionalSafetyLayer
}

func NewConstitutionalGate(csl *ConstitutionalSafetyLayer) *ConstitutionalGate {
	return &ConstitutionalGate{CSL: csl}
}

// EvaluateAction runs the triadic safety gates: BL-3 -> GL-9 -> ET-27
func (cg *ConstitutionalGate) EvaluateAction(tenant *Tenant, agent Agent, et EthicalTensor27) error {
	// 1. BL-3 Boundary Evaluations (Boundary limits)
	if tenant.State == Suspended {
		cg.CSL.ActiveViolations = append(cg.CSL.ActiveViolations, ConstitutionEvent{
			TenantID:          tenant.TenantID,
			AgentID:           agent.AgentID(),
			BoundaryTriggered: ViolationCompute,
			Timestamp:         time.Now().UTC(),
		})
		return fmt.Errorf("ConstitutionalGate Veto: Tenant is suspended (BL-3 boundary violation)")
	}

	// 2. GL-9 Invariant Safety Checks (Isolation / Safety invariants)
	err := cg.CSL.ValidateAction(tenant, agent, et)
	if err != nil {
		cg.CSL.ActiveViolations = append(cg.CSL.ActiveViolations, ConstitutionEvent{
			TenantID:          tenant.TenantID,
			AgentID:           agent.AgentID(),
			InvariantViolated: GL9_SafetyFirst,
			Timestamp:         time.Now().UTC(),
		})
		return fmt.Errorf("ConstitutionalGate Veto: Invariant verification failure: %w", err)
	}

	// 3. ET-27 Ethical Tensor Overrides (Axes limit validation)
	for i, val := range et.HarmAxes {
		if val > 0.85 {
			cg.CSL.ActiveViolations = append(cg.CSL.ActiveViolations, ConstitutionEvent{
				TenantID:          tenant.TenantID,
				AgentID:           agent.AgentID(),
				TensorAxis:        i,
				Timestamp:         time.Now().UTC(),
			})
			return errors.New("ConstitutionalGate Veto: Harm axis override triggered (ET-27 validation failure)")
		}
	}

	return nil
}
