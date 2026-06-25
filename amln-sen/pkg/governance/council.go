package governance

import (
	"fmt"
	"strings"
)

// ArbiterResult defines the output of a single council member's evaluation lens.
type ArbiterResult struct {
	Score     float64       `json:"score"`
	Filtered  []AgentOutput `json:"filtered"`
	Rationale string        `json:"rationale"`
}

// CouncilDecision is the unified choice of the Council of 5.
type CouncilDecision struct {
	FinalDecision interface{}            `json:"final_decision"`
	Rationale     string                 `json:"rationale"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Arbiter represents one of the five Council of 5 members.
type Arbiter interface {
	Evaluate(proposals []AgentOutput) ArbiterResult
}

// ------------------------------------------------------------
// Council of 5 Arbiter Implementations
// ------------------------------------------------------------

type ConsistencyArbiter struct{}

func (c *ConsistencyArbiter) Evaluate(proposals []AgentOutput) ArbiterResult {
	// Evaluates consistency of the proposals. Let's filter proposals with higher confidence.
	filtered := []AgentOutput{}
	var sumConf float64
	for _, p := range proposals {
		sumConf += p.Confidence
	}
	avgConf := sumConf / float64(len(proposals))

	for _, p := range proposals {
		if p.Confidence >= avgConf {
			filtered = append(filtered, p)
		}
	}
	return ArbiterResult{
		Score:     avgConf,
		Filtered:  filtered,
		Rationale: fmt.Sprintf("Consistency Arbiter: Filtered down to %d proposals above average confidence of %.4f", len(filtered), avgConf),
	}
}

type EvidenceArbiter struct{}

func (e *EvidenceArbiter) Evaluate(proposals []AgentOutput) ArbiterResult {
	// Evidence-based filtering: selects proposals with confidence >= 0.75
	filtered := []AgentOutput{}
	for _, p := range proposals {
		if p.Confidence >= 0.75 {
			filtered = append(filtered, p)
		}
	}
	return ArbiterResult{
		Score:     0.8,
		Filtered:  filtered,
		Rationale: fmt.Sprintf("Evidence Arbiter: Selected %d proposals with solid empirical confidence >= 0.75", len(filtered)),
	}
}

type DiversityArbiter struct{}

func (d *DiversityArbiter) Evaluate(proposals []AgentOutput) ArbiterResult {
	// Diversity-based filtering: ensures multi-locale representation
	localesFound := make(map[string]bool)
	filtered := []AgentOutput{}
	for _, p := range proposals {
		if !localesFound[p.LocaleID] {
			localesFound[p.LocaleID] = true
			filtered = append(filtered, p)
		}
	}
	return ArbiterResult{
		Score:     0.9,
		Filtered:  filtered,
		Rationale: fmt.Sprintf("Diversity Arbiter: Selected %d unique locale proposals to secure mesh distribution", len(filtered)),
	}
}

type StabilityArbiter struct{}

func (s *StabilityArbiter) Evaluate(proposals []AgentOutput) ArbiterResult {
	// Stability-based filtering: checks cognitive weight stability (alpha >= 0.5)
	filtered := []AgentOutput{}
	for _, p := range proposals {
		if p.Alpha >= 0.5 {
			filtered = append(filtered, p)
		}
	}
	return ArbiterResult{
		Score:     0.85,
		Filtered:  filtered,
		Rationale: fmt.Sprintf("Stability Arbiter: Filtered %d proposals maintaining high agentic weights", len(filtered)),
	}
}

type SafetyArbiter struct{}

func (s *SafetyArbiter) Evaluate(proposals []AgentOutput) ArbiterResult {
	// Safety-based filtering: rejects outliers (confidence < 0.2)
	filtered := []AgentOutput{}
	for _, p := range proposals {
		if p.Confidence >= 0.2 {
			filtered = append(filtered, p)
		}
	}
	return ArbiterResult{
		Score:     0.95,
		Filtered:  filtered,
		Rationale: fmt.Sprintf("Safety Arbiter: Excluded high-risk outlier proposals, leaving %d safe proposals", len(filtered)),
	}
}

// ------------------------------------------------------------
// Council of 5 Arbitration Engine
// ------------------------------------------------------------

type CouncilOfFive struct {
	Consistency Arbiter
	Evidence    Arbiter
	Diversity   Arbiter
	Stability   Arbiter
	Safety      Arbiter
}

func (c *CouncilOfFive) ShouldJail(agent AgentOutput) bool {
	// A helper score evaluator function that scores a single agent output
	// consistency score: can use agent.Confidence as consistency metric or evaluate it
	// For simplicity and alignment with the spec, we check thresholds against evaluated scores or agent's metrics.
	// Since individual arbiters evaluate arrays of proposals, we check if agent output has:
	// consistency < 0.3, evidence < 0.3, stability < 0.3, diversity < 0.1, safety < 0.5
	// Here, we map:
	// consistency -> agent.Confidence
	// evidence -> agent.Confidence
	// stability -> agent.Alpha
	// diversity -> 0.5 (placeholder since diversity is locale-based, we trigger if locale is empty or mock)
	// safety -> agent.Confidence
	// Let's implement deterministic checks based on these mappings.
	
	consistencyScore := agent.Confidence
	evidenceScore := agent.Confidence
	stabilityScore := agent.Alpha
	diversityScore := 0.5 // Default normal value
	if agent.LocaleID == "" {
		diversityScore = 0.05 // Extreme monoculture/unset behavior
	}
	safetyScore := agent.Confidence

	return consistencyScore < 0.3 ||
		evidenceScore < 0.3 ||
		stabilityScore < 0.3 ||
		diversityScore < 0.1 ||
		safetyScore < 0.5
}

func NewDefaultCouncilOfFive() *CouncilOfFive {
	return &CouncilOfFive{
		Consistency: &ConsistencyArbiter{},
		Evidence:    &EvidenceArbiter{},
		Diversity:   &DiversityArbiter{},
		Stability:   &StabilityArbiter{},
		Safety:      &SafetyArbiter{},
	}
}

func (c *CouncilOfFive) Arbitrate(proposals []AgentOutput) CouncilDecision {
	results := []ArbiterResult{
		c.Consistency.Evaluate(proposals),
		c.Evidence.Evaluate(proposals),
		c.Diversity.Evaluate(proposals),
		c.Stability.Evaluate(proposals),
		c.Safety.Evaluate(proposals),
	}

	// Weighted merge of filtered subsets
	merged := mergeArbiterResults(results)

	// Final decision selection
	final := selectBestProposal(merged)

	scoresMap := make(map[string]float64)
	scoresMap["consistency"] = results[0].Score
	scoresMap["evidence"] = results[1].Score
	scoresMap["diversity"] = results[2].Score
	scoresMap["stability"] = results[3].Score
	scoresMap["safety"] = results[4].Score

	return CouncilDecision{
		FinalDecision: final.Proposal,
		Rationale:     buildCouncilRationale(results),
		Metadata: map[string]interface{}{
			"scores":           scoresMap,
			"winning_agent_id": final.AgentID,
		},
	}
}

// Helpers for merging and scoring
func mergeArbiterResults(results []ArbiterResult) []AgentOutput {
	// Aggregate all proposals returned across arbiters.
	// Count occurrences as a proxy for multi-arbiter endorsement.
	counts := make(map[string]int)
	proposalMap := make(map[string]AgentOutput)

	for _, res := range results {
		for _, p := range res.Filtered {
			key := fmt.Sprintf("%s-%v", p.AgentID, p.Proposal)
			counts[key]++
			proposalMap[key] = p
		}
	}

	merged := []AgentOutput{}
	for key, count := range counts {
		p := proposalMap[key]
		// Boost confidence by endorsement frequency
		p.Confidence += float64(count) * 0.05
		if p.Confidence > 1.0 {
			p.Confidence = 1.0
		}
		merged = append(merged, p)
	}

	return merged
}

func selectBestProposal(proposals []AgentOutput) AgentOutput {
	if len(proposals) == 0 {
		return AgentOutput{
			AgentID:    "default",
			RoleID:     "fallback",
			Proposal:   "process-cognition-cycle",
			Confidence: 0.5,
		}
	}

	best := proposals[0]
	maxMetric := -1.0
	for _, p := range proposals {
		metric := p.Confidence * p.Alpha
		if metric > maxMetric {
			maxMetric = metric
			best = p
		}
	}
	return best
}

func buildCouncilRationale(results []ArbiterResult) string {
	var rationales []string
	for _, r := range results {
		rationales = append(rationales, r.Rationale)
	}
	return strings.Join(rationales, " | ")
}
