package governance

import (
	"math"
)

// ------------------------------------------------------------
// 1. Agentic Memory Node (AMLN)
// ------------------------------------------------------------

type AMLN struct {
	MemoryPointer  string  `json:"memory_pointer"`
	AgenticWeight  float64 `json:"agentic_weight"`
	StabilityScore float64 `json:"stability_score"`
	DriftScore     float64 `json:"drift_score"`
}

func (m *AMLN) ComputeAgenticWeight() float64 {
	// Formula: 0.30*Stability + 0.25*(1 - Drift) + 0.20*Ethics + 0.15*Council + 0.10*Memory
	// Let's model ethics, council alignment, memory continuity as standard parameters derived from stability/drift scores
	ethicsCompliance := 1.0 - m.DriftScore
	if ethicsCompliance < 0 {
		ethicsCompliance = 0
	}
	councilAlignment := m.StabilityScore
	memoryContinuity := 1.0 - (0.5 * m.DriftScore)

	m.AgenticWeight = 0.30*m.StabilityScore +
		0.25*(1.0-m.DriftScore) +
		0.20*ethicsCompliance +
		0.15*councilAlignment +
		0.10*memoryContinuity

	return m.AgenticWeight
}

// ------------------------------------------------------------
// 2. Hypothesis-Delta Engine
// ------------------------------------------------------------

type Hypothesis struct {
	Model      []float64 `json:"model"`
	Confidence float64   `json:"confidence"`
	Lineage    []float64 `json:"lineage"`
	Entropy    float64   `json:"entropy"`
}

type HypothesisDelta struct {
	DeltaH          float64 `json:"delta_h"`
	DeltaConfidence float64 `json:"delta_confidence"`
	DeltaLineage    float64 `json:"delta_lineage"`
	DeltaEntropy    float64 `json:"delta_entropy"`
}

func ComputeHypothesisDelta(oldHyp, newHyp Hypothesis) HypothesisDelta {
	var sum float64
	n := len(oldHyp.Model)
	if n > len(newHyp.Model) {
		n = len(newHyp.Model)
	}
	for i := 0; i < n; i++ {
		diff := oldHyp.Model[i] - newHyp.Model[i]
		sum += diff * diff
	}
	dist := math.Sqrt(sum)

	var linSum float64
	nl := len(oldHyp.Lineage)
	if nl > len(newHyp.Lineage) {
		nl = len(newHyp.Lineage)
	}
	for i := 0; i < nl; i++ {
		linSum += math.Abs(oldHyp.Lineage[i] - newHyp.Lineage[i])
	}

	return HypothesisDelta{
		DeltaH:          dist,
		DeltaConfidence: newHyp.Confidence - oldHyp.Confidence,
		DeltaLineage:    linSum,
		DeltaEntropy:    newHyp.Entropy - oldHyp.Entropy,
	}
}

// ------------------------------------------------------------
// 3. Conformation-Tolerance Engine (θ-based)
// ------------------------------------------------------------

type ConformationTolerance struct {
	Theta                float64 `json:"theta"`
	NegotiationTolerance float64 `json:"negotiation_tolerance"`
	MutationTolerance    float64 `json:"mutation_tolerance"`
	EthicalTolerance     float64 `json:"ethical_tolerance"`
}

func ComputeTolerance(theta float64) ConformationTolerance {
	return ConformationTolerance{
		Theta:                theta,
		NegotiationTolerance: 1.0 - theta,
		MutationTolerance:    0.5 * theta,
		EthicalTolerance:     0.2 * (1.0 - theta),
	}
}

// ------------------------------------------------------------
// 4. Global Stability Regulator (GSR)
// ------------------------------------------------------------

type GlobalStabilityRegulator struct {
	EntropyLevel    float64 `json:"entropy_level"`
	LineageVariance float64 `json:"lineage_variance"`
	Theta           float64 `json:"theta"`
	AgenticWeight   float64 `json:"agentic_weight"`
	EthicalVariance float64 `json:"ethical_variance"`
	StabilityScore  float64 `json:"stability_score"`
}

func (g *GlobalStabilityRegulator) ComputeStability() float64 {
	g.StabilityScore = 0.25*(1.0-g.EntropyLevel) +
		0.25*(1.0-g.LineageVariance) +
		0.20*(1.0-g.EthicalVariance) +
		0.15*(1.0-g.Theta) +
		0.15*g.AgenticWeight
	return g.StabilityScore
}

// ------------------------------------------------------------
// 5. Council-Aware Mutation Governor
// ------------------------------------------------------------

type MutationGovernor struct {
	BaseRate          float64 `json:"base_rate"`
	StabilityModifier float64 `json:"stability_modifier"`
	EthicalModifier   float64 `json:"ethical_modifier"`
	CouncilModifier   float64 `json:"council_modifier"`
}

func (mg *MutationGovernor) ComputeRate(gsr GlobalStabilityRegulator) float64 {
	return mg.BaseRate *
		mg.StabilityModifier *
		mg.EthicalModifier *
		mg.CouncilModifier *
		(1.0 - gsr.Theta)
}

// ------------------------------------------------------------
// 6. Ethical Tensor Feedback Loop & Multi-Agent Predictive Model
// ------------------------------------------------------------

type EthicalFeedback struct {
	Pressure     []float64 `json:"pressure"`
	Gradient     []float64 `json:"gradient"`
	DriftPenalty float64   `json:"drift_penalty"`
}

type PredictiveModel struct {
	CouncilPrediction   float64            `json:"council_prediction"`
	EthicalPrediction   float64            `json:"ethical_prediction"`
	PeerPrediction      map[string]float64 `json:"peer_prediction"`
	StabilityPrediction float64            `json:"stability_prediction"`
}

func NewPredictiveModel() *PredictiveModel {
	return &PredictiveModel{
		PeerPrediction: make(map[string]float64),
	}
}

func (pm *PredictiveModel) Update(cf *ConsensusField, gsr GlobalStabilityRegulator) {
	cf.Mu.RLock()
	defer cf.Mu.RUnlock()

	// Simple non-anthropomorphic predictive derivations
	pm.StabilityPrediction = gsr.StabilityScore
	pm.CouncilPrediction = 0.8
	pm.EthicalPrediction = 0.9
}
