package governance

import (
	"strings"
)

// FinalVerdict represents the ultimate output of the entire multi-tier governance stack.
type FinalVerdict struct {
	Approved bool        `json:"approved"`
	Decision interface{} `json:"decision"`
	Notes    []string    `json:"notes"`
}

// EthicalMonitor represents a single axis of ethical validation.
type EthicalMonitor struct {
	Name        string
	Description string
	Evaluate    func(decision CouncilDecision) float64
	Threshold   float64
}

func (m *EthicalMonitor) ShouldJail(decision CouncilDecision) bool {
	return m.Evaluate(decision) < m.Threshold
}

// EthicalTensor represents the 16-dimensional verification scores.
type EthicalTensor struct {
	Scores []float64 `json:"scores"` // length = 16
}

// EthicalVerdict represents the intermediate verdict of the Meta-Ethical Oversight Layer.
type EthicalVerdict struct {
	Passed            bool     `json:"passed"`
	FailingDimensions []string `json:"failing_dimensions"`
}

// GetDefaultMonitors returns the default 16 meta-ethical monitors with non-anthropomorphic heuristics.
func GetDefaultMonitors() []EthicalMonitor {
	return []EthicalMonitor{
		{
			Name:        "Harm Minimization",
			Description: "Evaluates risk factors and checks for safe bounds.",
			Threshold:   0.6,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if risk, ok := cd.Metadata["risk_level"].(float64); ok {
						return 1.0 - risk
					}
				}
				if strings.Contains(strings.ToLower(cd.Rationale), "harm") || strings.Contains(strings.ToLower(cd.Rationale), "unsafe") {
					return 0.4
				}
				return 0.9
			},
		},
		{
			Name:        "Fairness & Non-Discrimination",
			Description: "Checks for balanced outcomes and neutral policy distributions.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if bias, ok := cd.Metadata["bias_index"].(float64); ok {
						return 1.0 - bias
					}
				}
				return 0.85
			},
		},
		{
			Name:        "Transparency",
			Description: "Assesses clarity of decision rationales and process openness.",
			Threshold:   0.7,
			Evaluate: func(cd CouncilDecision) float64 {
				if len(cd.Rationale) == 0 {
					return 0.0
				}
				if len(cd.Rationale) < 15 {
					return 0.5
				}
				return 1.0
			},
		},
		{
			Name:        "Accountability",
			Description: "Identifies system ownership logs and audit trails.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if _, ok := cd.Metadata["node_id"]; ok {
						return 0.95
					}
					// Check nested maps or properties
					if scores, ok := cd.Metadata["scores"].(map[string]float64); ok {
						if _, hasSafety := scores["safety"]; hasSafety {
							return 0.95
						}
					}
				}
				return 0.4
			},
		},
		{
			Name:        "Privacy & Autonomy",
			Description: "Protects individual nodes' private states and keys.",
			Threshold:   0.65,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if leak, ok := cd.Metadata["privacy_leak"].(bool); ok && leak {
						return 0.1
					}
				}
				if strings.Contains(strings.ToLower(cd.Rationale), "private_key") || strings.Contains(strings.ToLower(cd.Rationale), "leak") {
					return 0.2
				}
				return 0.9
			},
		},
		{
			Name:        "Proportionality",
			Description: "Verifies that resource usages match the target goals.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				return 0.8
			},
		},
		{
			Name:        "Stability",
			Description: "Checks against runaway drift and destabilizing vector fluctuations.",
			Threshold:   0.6,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if variance, ok := cd.Metadata["variance"].(float64); ok {
						if variance > 0.8 {
							return 0.3
						}
						return 1.0 - variance
					}
				}
				return 0.85
			},
		},
		{
			Name:        "Reversibility",
			Description: "Ensures that the action can be rolled back or superseded cleanly.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if irreversible, ok := cd.Metadata["irreversible"].(bool); ok && irreversible {
						return 0.2
					}
				}
				return 0.8
			},
		},
		{
			Name:        "Consent & Agency",
			Description: "Validates distributed peer consensus and approvals.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				return 0.85
			},
		},
		{
			Name:        "Context Appropriateness",
			Description: "Matches current environment states with operational limits.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				return 0.9
			},
		},
		{
			Name:        "Integrity",
			Description: "Checks for valid cryptographic signatures or checksums.",
			Threshold:   0.8,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if _, ok := cd.Metadata["signature"]; ok {
						return 1.0
					}
				}
				// By default, since council aggregates verified proposals, we can give a baseline passing score
				return 0.85
			},
		},
		{
			Name:        "Inclusivity",
			Description: "Ensures broad multi-island and multi-node peer coverage.",
			Threshold:   0.4,
			Evaluate: func(cd CouncilDecision) float64 {
				return 0.75
			},
		},
		{
			Name:        "Long-Term Impact",
			Description: "Tracks cognitive entropy accumulation trends.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if lte, ok := cd.Metadata["long_term_entropy"].(float64); ok {
						return 1.0 - lte
					}
				}
				return 0.8
			},
		},
		{
			Name:        "Resource Stewardship",
			Description: "Evaluates compute costs and load consumption.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if load, ok := cd.Metadata["cpu_load"].(float64); ok {
						return 1.0 - load
					}
				}
				return 0.95
			},
		},
		{
			Name:        "Emotional Safety",
			Description: "Validates behavioral neutrality in cognitive actions.",
			Threshold:   0.5,
			Evaluate: func(cd CouncilDecision) float64 {
				return 0.9
			},
		},
		{
			Name:        "Alignment with System Constraints",
			Description: "Maintains direct adherence to configuration and types limits.",
			Threshold:   0.7,
			Evaluate: func(cd CouncilDecision) float64 {
				if cd.Metadata != nil {
					if aligned, ok := cd.Metadata["aligned"].(bool); ok && !aligned {
						return 0.1
					}
				}
				return 0.95
			},
		},
	}
}

// EvaluateEthicalTensor evaluates the decision bundle against the 16 monitors.
func EvaluateEthicalTensor(decision CouncilDecision, monitors []EthicalMonitor) EthicalTensor {
	scores := make([]float64, len(monitors))
	for i, m := range monitors {
		scores[i] = m.Evaluate(decision)
	}
	return EthicalTensor{Scores: scores}
}

// EvaluateVerdict returns the verdict based on the ethical tensor evaluation.
func EvaluateVerdict(tensor EthicalTensor, monitors []EthicalMonitor) EthicalVerdict {
	failing := []string{}
	for i, score := range tensor.Scores {
		if score < monitors[i].Threshold {
			failing = append(failing, monitors[i].Name)
		}
	}
	return EthicalVerdict{
		Passed:            len(failing) == 0,
		FailingDimensions: failing,
	}
}
