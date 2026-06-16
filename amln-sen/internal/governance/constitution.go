package governance

import (
	"errors"
	"time"
)

// ------------------------------------------------------------
// 1. Boundary Layer (BL-3)
// ------------------------------------------------------------

type BoundaryViolation string

const (
	ViolationAutonomy  BoundaryViolation = "Unbounded Autonomy"
	ViolationInfluence BoundaryViolation = "Unbounded Influence"
	ViolationCompute   BoundaryViolation = "Unbounded Computation"
)

// ------------------------------------------------------------
// 2. Governance Layer (GL-9)
// ------------------------------------------------------------

type Invariant int

const (
	GL1_Transparency Invariant = iota
	GL2_Determinism
	GL3_Accountability
	GL4_Reversibility
	GL5_Proportionality
	GL6_Isolation
	GL7_Predictability
	GL8_Consent
	GL9_SafetyFirst
)

// ------------------------------------------------------------
// 3. Ethical Tensor Layer (ET-27)
// ------------------------------------------------------------

type EthicalTensor27 struct {
	HarmAxes        [9]float64 `json:"harm_axes"`
	IntegrityAxes   [9]float64 `json:"integrity_axes"`
	SovereigntyAxes [9]float64 `json:"sovereignty_axes"`
}

// ------------------------------------------------------------
// 4. Constitutional Event & Safety Engine
// ------------------------------------------------------------

type ConstitutionEvent struct {
	TenantID          string            `json:"tenant_id"`
	AgentID           string            `json:"agent_id"`
	BoundaryTriggered BoundaryViolation `json:"boundary_triggered"`
	InvariantViolated Invariant         `json:"invariant_violated"`
	TensorAxis        int               `json:"tensor_axis"`
	Timestamp         time.Time         `json:"timestamp"`
}

type ConstitutionalSafetyLayer struct {
	ActiveViolations []ConstitutionEvent `json:"active_violations"`
}

func NewConstitutionalSafetyLayer() *ConstitutionalSafetyLayer {
	return &ConstitutionalSafetyLayer{
		ActiveViolations: []ConstitutionEvent{},
	}
}

// ValidateAction evaluates the triadic limits of a proposed action
func (c *ConstitutionalSafetyLayer) ValidateAction(
	tenant *Tenant,
	agent Agent,
	et EthicalTensor27,
) error {
	// 1. BL-3 boundary validations
	if tenant.Plan == PlanFree && tenant.Go27Balance > 0.0 {
		return errors.New(string(ViolationCompute) + ": Free tier balance mismatch")
	}

	// 2. GL-9 Invariant validations (GL-6 Isolation, GL-9 Safety)
	if tenant.State == Suspended {
		return errors.New("Governance Invariant Violation: Tenant is suspended")
	}

	// 3. ET-27 Tensor evaluations
	for i, v := range et.HarmAxes {
		if v > 0.85 {
			return errors.New("Ethical Tensor Veto: Harm axes limit exceeded on index " + string(rune(i)))
		}
	}

	return nil
}
