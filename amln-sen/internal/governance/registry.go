package governance

import (
	"sync"
	"time"
)

// ------------------------------------------------------------
// 1. Role-Affinity Matrix
// ------------------------------------------------------------

type RoleAffinityMatrix struct {
	ServerAffinity map[string]float64 `json:"server_affinity"`
	TaskAffinity   map[string]float64 `json:"task_affinity"`
	PeerAffinity   map[string]float64 `json:"peer_affinity"`
}

type RoleProfile struct {
	RoleID   string             `json:"role_id"`
	Affinity RoleAffinityMatrix `json:"affinity"`
}

// ------------------------------------------------------------
// 2. Specialization Behavior Engine
// ------------------------------------------------------------

type Specialization int

const (
	GameTheory Specialization = iota
	HumanDesign
	General
)

type SpecializationBehavior struct {
	ExplorationWeight  float64 `json:"exploration_weight"`
	ExploitationWeight float64 `json:"exploitation_weight"`
	NarrativeWeight    float64 `json:"narrative_weight"`
	StrategyWeight     float64 `json:"strategy_weight"`
	ConsensusWeight    float64 `json:"consensus_weight"`
	EthicalSensitivity float64 `json:"ethical_sensitivity"`
}

func GetSpecializationBehavior(spec Specialization) SpecializationBehavior {
	switch spec {
	case GameTheory:
		return SpecializationBehavior{
			ExplorationWeight:  0.4,
			ExploitationWeight: 0.9,
			NarrativeWeight:    0.1,
			StrategyWeight:     0.95,
			ConsensusWeight:    0.8,
			EthicalSensitivity: 0.5,
		}
	case HumanDesign:
		return SpecializationBehavior{
			ExplorationWeight:  0.8,
			ExploitationWeight: 0.5,
			NarrativeWeight:    0.9,
			StrategyWeight:     0.3,
			ConsensusWeight:    0.6,
			EthicalSensitivity: 0.7,
		}
	default:
		return SpecializationBehavior{
			ExplorationWeight:  0.6,
			ExploitationWeight: 0.6,
			NarrativeWeight:    0.5,
			StrategyWeight:     0.5,
			ConsensusWeight:    0.6,
			EthicalSensitivity: 0.6,
		}
	}
}

// ------------------------------------------------------------
// 3. Global Agent Registry
// ------------------------------------------------------------

type AgentRecord struct {
	AgentID        string         `json:"agent_id"`
	Specialization Specialization `json:"specialization"`
	RoleID         string         `json:"role_id"`
	LocaleID       string         `json:"locale_id"`
	CurrentNode    string         `json:"current_node"`
	LeaseHolder    string         `json:"lease_holder"`
	InJail         bool           `json:"in_jail"`
	OnProbation    bool           `json:"on_probation"`
}

type GlobalAgentRegistry struct {
	mu     sync.RWMutex
	Agents map[string]*AgentRecord `json:"agents"`
}

func NewGlobalAgentRegistry() *GlobalAgentRegistry {
	return &GlobalAgentRegistry{
		Agents: make(map[string]*AgentRecord),
	}
}

func (r *GlobalAgentRegistry) Register(agent Agent, spec Specialization, currentNode string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Agents[agent.AgentID()] = &AgentRecord{
		AgentID:        agent.AgentID(),
		Specialization: spec,
		RoleID:         agent.RoleID(),
		LocaleID:       "US-EAST", // default mapping
		CurrentNode:    currentNode,
		InJail:         false,
		OnProbation:    false,
	}
}

// ------------------------------------------------------------
// 4. Global Consensus Field & Message Bus & Negotiation
// ------------------------------------------------------------

type ConsensusField struct {
	Mu           sync.RWMutex
	LineageField map[string][]float64 `json:"lineage_field"`
	EntropyField map[string]float64   `json:"entropy_field"`
	EthicalField map[string]float64   `json:"ethical_field"`
	CouncilField CouncilDecision      `json:"council_field"`
	JailField    map[string]bool      `json:"jail_field"`
}

func NewConsensusField() *ConsensusField {
	return &ConsensusField{
		LineageField: make(map[string][]float64),
		EntropyField: make(map[string]float64),
		EthicalField: make(map[string]float64),
		JailField:    make(map[string]bool),
	}
}

type NegotiationPacket struct {
	FromAgent  string      `json:"from_agent"`
	Proposal   interface{} `json:"proposal"`
	Confidence float64     `json:"confidence"`
	Lineage    []float64   `json:"lineage"`
	Entropy    float64     `json:"entropy"`
}

type AgentMessage struct {
	From      string      `json:"from"`
	To        string      `json:"to"`
	Payload   interface{} `json:"payload"`
	Priority  float64     `json:"priority"`
	Timestamp time.Time   `json:"timestamp"`
}

type MessagingBus struct {
	mu     sync.RWMutex
	Queues map[string][]AgentMessage `json:"queues"`
}

func NewMessagingBus() *MessagingBus {
	return &MessagingBus{
		Queues: make(map[string][]AgentMessage),
	}
}

func (b *MessagingBus) Send(msg AgentMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Queues[msg.To] = append(b.Queues[msg.To], msg)
}

func (b *MessagingBus) Receive(agentID string) []AgentMessage {
	b.mu.Lock()
	defer b.mu.Unlock()
	msgs := b.Queues[agentID]
	delete(b.Queues, agentID)
	return msgs
}
