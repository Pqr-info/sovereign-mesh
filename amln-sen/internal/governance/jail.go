package governance

import (
	"errors"
	"sync"
	"time"
)

// AgentStateSnapshot represents the frozen memory and strategy of a quarantined agent.
type AgentStateSnapshot struct {
	Lineage       []float64 `json:"lineage"`
	Entropy       float64   `json:"entropy"`
	SELStrategy   []float64 `json:"sel_strategy"`
	RoleVector    []float64 `json:"role_vector"`
	LocaleID      string    `json:"locale_id"`
	CognitiveK    []float64 `json:"cognitive_k"`
	RewardHistory []float64 `json:"reward_history"`
	MemoryPointer string    `json:"memory_pointer"`
	Alpha         float64   `json:"alpha"`
	RoleID        string    `json:"role_id"`
}

// JailRecord represents the containment log for a quarantined agent.
type JailRecord struct {
	AgentID     string             `json:"agent_id"`
	Reason      string             `json:"reason"`
	Timestamp   time.Time          `json:"timestamp"`
	State       AgentStateSnapshot `json:"state"`
	ReviewNotes []string           `json:"review_notes"`
}

// LeaseManager coordinates active leases to prevent teleportation duplication races.
type LeaseManager struct {
	mu     sync.Mutex
	leases map[string]bool
}

// NewLeaseManager constructs a new lease manager.
func NewLeaseManager() *LeaseManager {
	return &LeaseManager{
		leases: make(map[string]bool),
	}
}

// Grant issues a lease lock for an agent ID.
func (lm *LeaseManager) Grant(agentID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.leases[agentID] = true
}

// Revoke terminates the lease lock for an agent ID.
func (lm *LeaseManager) Revoke(agentID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	delete(lm.leases, agentID)
}

// IsActive checks if the agent lease is currently active.
func (lm *LeaseManager) IsActive(agentID string) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.leases[agentID]
}

// JailController manages the lifecycle of quarantined agents.
type JailController struct {
	mu           sync.RWMutex
	Active       map[string]*JailRecord
	LeaseManager *LeaseManager
}

// NewJailController constructs a new jail controller.
func NewJailController(lm *LeaseManager) *JailController {
	return &JailController{
		Active:       make(map[string]*JailRecord),
		LeaseManager: lm,
	}
}

// JailAgent isolates the agent, revokes its lease, and snapshots its state.
func (jc *JailController) JailAgent(agent Agent, reason string) error {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	agent.Freeze()
	agentID := agent.AgentID()
	jc.LeaseManager.Revoke(agentID)

	snapshot := agent.Snapshot()

	jc.Active[agentID] = &JailRecord{
		AgentID:     agentID,
		Reason:      reason,
		Timestamp:   time.Now().UTC(),
		State:       snapshot,
		ReviewNotes: []string{"Agent isolated due to: " + reason},
	}

	return nil
}

// ReleaseAgent restores agent lease and removes containment.
func (jc *JailController) ReleaseAgent(agentID string) error {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	record, ok := jc.Active[agentID]
	if !ok {
		return errors.New("agent not in jail")
	}

	agent, err := LoadAgentFromSnapshot(record.State, agentID, record.State.RoleID)
	if err == nil {
		agent.Resume()
	}

	jc.LeaseManager.Grant(agentID)
	delete(jc.Active, agentID)

	return nil
}

// RetireAgent permanently archives the agent.
func (jc *JailController) RetireAgent(agentID string) error {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	record, ok := jc.Active[agentID]
	if !ok {
		return errors.New("agent not in jail")
	}

	record.ReviewNotes = append(record.ReviewNotes, "Agent permanently retired.")
	delete(jc.Active, agentID)

	return nil
}
