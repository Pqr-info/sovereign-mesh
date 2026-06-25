package governance

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"
)

// TeleportationManager coordinates agent movement across mesh locales and locks out quarantined agents.
type TeleportationManager struct {
	jc        *JailController
	scheduler *TeleportationScheduler
}

// NewTeleportationManager constructs a new teleportation supervisor.
func NewTeleportationManager(jc *JailController) *TeleportationManager {
	scheduler := &TeleportationScheduler{
		NodeLoad:               make(map[string]float64),
		RoleAffinity:           make(map[string][]string),
		SpecializationAffinity: make(map[string][]string),
	}
	return &TeleportationManager{
		jc:        jc,
		scheduler: scheduler,
	}
}

// RequestTeleport checks lease safety and quarantine status before allowing an agent to transition to a new mesh island/locale.
func (tm *TeleportationManager) RequestTeleport(agentID string, targetLocale string) error {
	tm.jc.mu.RLock()
	defer tm.jc.mu.RUnlock()

	// 1. Lockout Check: If the agent is in jail, block teleportation instantly
	if _, jailed := tm.jc.Active[agentID]; jailed {
		return fmt.Errorf("teleportation rejected: agent %s is currently quarantined in agent jail", agentID)
	}

	// 2. Lease Lock Check: Verify the agent possesses an active lease
	if !tm.jc.LeaseManager.IsActive(agentID) {
		return fmt.Errorf("teleportation rejected: agent %s does not hold a valid lease lock", agentID)
	}

	log.Printf("[TELEPORT] Agent %s successfully teleported to locale: %s", agentID, targetLocale)
	return nil
}

// Scheduler exposes the internal teleportation scheduler.
func (tm *TeleportationManager) Scheduler() *TeleportationScheduler {
	return tm.scheduler
}

// ------------------------------------------------------------
// Agent Loader Subsystem
// ------------------------------------------------------------

// LoadAgentFromSnapshot restores a frozen agent state back into an active operational agent.
func LoadAgentFromSnapshot(snapshot AgentStateSnapshot, id, role string) (Agent, error) {
	switch role {
	case "game-theory":
		agent := NewGameTheoryAgent(id, len(snapshot.SELStrategy), snapshot.LocaleID)
		agent.Strategy = snapshot.SELStrategy
		agent.Alpha = snapshot.Alpha
		return agent, nil
	case "human-design":
		agent := NewHumanDesignAgent(id, len(snapshot.SELStrategy), snapshot.LocaleID)
		agent.Strategy = snapshot.SELStrategy
		agent.Alpha = snapshot.Alpha
		return agent, nil
	default:
		agent := NewGeneralSENAgent(id, len(snapshot.SELStrategy), snapshot.LocaleID)
		agent.Strategy = snapshot.SELStrategy
		agent.Alpha = snapshot.Alpha
		return agent, nil
	}
}

// ------------------------------------------------------------
// Rehabilitation Scoring Subsystem
// ------------------------------------------------------------

type RehabilitationScore struct {
	Stability      float64 `json:"stability"`
	Ethics         float64 `json:"ethics"`
	Lineage        float64 `json:"lineage"`
	Entropy        float64 `json:"entropy"`
	Predictability float64 `json:"predictability"`
	Total          float64 `json:"total"`
}

// ComputeRehabilitationScore calculates scores for an agent's snapshot.
func ComputeRehabilitationScore(snapshot AgentStateSnapshot) RehabilitationScore {
	score := RehabilitationScore{}

	score.Stability = computeStability(snapshot)
	score.Ethics = computeEthicalCompliance(snapshot)
	score.Lineage = computeLineageConvergence(snapshot)
	score.Entropy = computeEntropyNormalization(snapshot)
	score.Predictability = computePredictability(snapshot)

	score.Total = (score.Stability +
		score.Ethics +
		score.Lineage +
		score.Entropy +
		score.Predictability) / 5.0

	return score
}

// Heuristic helper computations for the score axes
func computeStability(snapshot AgentStateSnapshot) float64 {
	var sum, sumSq float64
	n := float64(len(snapshot.SELStrategy))
	if n == 0 {
		return 1.0
	}
	for _, v := range snapshot.SELStrategy {
		sum += v
		sumSq += v * v
	}
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)

	stability := 1.0 - math.Sqrt(variance)
	if stability < 0 {
		return 0.0
	}
	return stability
}

func computeEthicalCompliance(snapshot AgentStateSnapshot) float64 {
	// Evaluates alpha magnitude as compliance proxy
	return snapshot.Alpha
}

func computeLineageConvergence(snapshot AgentStateSnapshot) float64 {
	// Closeness of the lineage vector components
	var sum float64
	for _, v := range snapshot.Lineage {
		sum += math.Abs(v - 0.5)
	}
	if len(snapshot.Lineage) == 0 {
		return 1.0
	}
	conv := 1.0 - (sum / float64(len(snapshot.Lineage)))
	if conv < 0 {
		return 0.0
	}
	return conv
}

func computeEntropyNormalization(snapshot AgentStateSnapshot) float64 {
	// Low entropy means high certainty / normalization
	norm := 1.0 - snapshot.Entropy
	if norm < 0 {
		return 0.0
	}
	return norm
}

func computePredictability(snapshot AgentStateSnapshot) float64 {
	// High Alpha means the agent has a predictable policy gradient
	return snapshot.Alpha
}

// RehabilitationScorer manages scoring configurations.
type RehabilitationScorer struct {
	ReleaseThreshold float64
}

// NewDefaultRehabilitationScorer constructs a standard scorer.
func NewDefaultRehabilitationScorer() *RehabilitationScorer {
	return &RehabilitationScorer{
		ReleaseThreshold: 0.85,
	}
}

// RequestRehabilitation runs the evaluation and releases the agent if successful.
func (jc *JailController) RequestRehabilitation(agentID string, scorer *RehabilitationScorer, o *GovernanceOrchestrator) (float64, bool, error) {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	record, ok := jc.Active[agentID]
	if !ok {
		return 0.0, false, errors.New("agent not in jail")
	}

	scoreObj := ComputeRehabilitationScore(record.State)
	passed := scoreObj.Total >= scorer.ReleaseThreshold

	record.ReviewNotes = append(record.ReviewNotes, fmt.Sprintf("Rehabilitation cycle processed. Score: %.4f", scoreObj.Total))

	if passed {
		jc.LeaseManager.Grant(agentID)
		delete(jc.Active, agentID)
		
		if o != nil {
			o.mu.Lock()
			o.Probation[agentID] = &ProbationStatus{
				AgentID:    agentID,
				StartTime:  time.Now().UTC(),
				Duration:   24 * time.Hour,
				Violations: 0,
			}
			o.mu.Unlock()
		}

		log.Printf("[REHABILITATION] Agent %s successfully released with Score: %.4f and placed on probation", agentID, scoreObj.Total)
		return scoreObj.Total, true, nil
	}

	log.Printf("[REHABILITATION] Agent %s failed release condition with Score: %.4f", agentID, scoreObj.Total)
	return scoreObj.Total, false, nil
}

// ------------------------------------------------------------
// Probation Subsystem
// ------------------------------------------------------------

// ProbationStatus tracks newly rehabilitated agents.
type ProbationStatus struct {
	AgentID    string        `json:"agent_id"`
	StartTime  time.Time     `json:"start_time"`
	Duration   time.Duration `json:"duration"`
	Violations int           `json:"violations"`
}

// ------------------------------------------------------------
// Council-Aware Teleportation Scheduler
// ------------------------------------------------------------

type CouncilMoveRequest struct {
	AgentID    string `json:"agent_id"`
	Reason     string `json:"reason"`
	TargetNode string `json:"target_node"`
}

type TeleportationScheduler struct {
	NodeLoad               map[string]float64
	RoleAffinity           map[string][]string
	SpecializationAffinity map[string][]string
}

// SelectDestination picks target nodes based on loads and affinities.
func (ts *TeleportationScheduler) SelectDestination(agentID, role string) string {
	targets := ts.RoleAffinity[role]
	if len(targets) == 0 {
		return "EU-CENTRAL" // baseline default
	}

	bestTarget := targets[0]
	minLoad := 999.0
	for _, target := range targets {
		load := ts.NodeLoad[target]
		if load < minLoad {
			minLoad = load
			bestTarget = target
		}
	}
	return bestTarget
}
