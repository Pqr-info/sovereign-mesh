package governance

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

// GovernanceOrchestrator coordinates the active 195 agents, the Council of 5, and the 16 ethical monitors.
type GovernanceOrchestrator struct {
	Agents         []Agent
	Council        *CouncilOfFive
	Monitors       []EthicalMonitor
	Tick           time.Duration
	JailController *JailController
	Probation      map[string]*ProbationStatus
	mu             sync.Mutex

	// Subsystems
	Registry          *GlobalAgentRegistry
	ConsensusField    *ConsensusField
	MessageBus        *MessagingBus
	GSR               *GlobalStabilityRegulator
	MutationGovernor  *MutationGovernor
	TenantManager     *TenantLifecycleManager
	CSL               *ConstitutionalSafetyLayer
	REE               *RuntimeExecutionEngine
	FMRP              *SovereignRecoveryProtocol

	// Limits for runaway detection
	EntropyLimit    float64
	DriftLimit      float64
	StrategyLimit   float64
}

// NewGovernanceOrchestrator constructs the orchestration engine.
func NewGovernanceOrchestrator(vectorSize int, tickBase time.Duration) *GovernanceOrchestrator {
	agents := Initialize195Agents(vectorSize)
	council := NewDefaultCouncilOfFive()
	monitors := GetDefaultMonitors()

	lm := NewLeaseManager()
	jc := NewJailController(lm)

	registry := NewGlobalAgentRegistry()
	field := NewConsensusField()
	bus := NewMessagingBus()
	tm := NewTenantLifecycleManager()
	csl := NewConstitutionalSafetyLayer()

	// Grant initial leases to all active agents and register them
	for i, agent := range agents {
		lm.Grant(agent.AgentID())
		spec := General
		if i < 64 {
			spec = GameTheory
		} else if i < 128 {
			spec = HumanDesign
		}
		registry.Register(agent, spec, "NODE-0")
	}

	orchestrator := &GovernanceOrchestrator{
		Agents:         agents,
		Council:        council,
		Monitors:       monitors,
		Tick:           tickBase,
		JailController: jc,
		Probation:      make(map[string]*ProbationStatus),
		Registry:       registry,
		ConsensusField: field,
		MessageBus:     bus,
		GSR: &GlobalStabilityRegulator{
			EntropyLevel:    0.2,
			LineageVariance: 0.1,
			Theta:           0.3,
			AgenticWeight:   0.8,
			EthicalVariance: 0.05,
		},
		MutationGovernor: &MutationGovernor{
			BaseRate:          0.05,
			StabilityModifier: 1.0,
			EthicalModifier:   1.0,
			CouncilModifier:   1.0,
		},
		TenantManager: tm,
		CSL:           csl,
		REE:           NewRuntimeExecutionEngine(nil), // Will be self-referenced below
		FMRP:          NewSovereignRecoveryProtocol(nil), // Will be self-referenced below
		EntropyLimit:  0.85,
		DriftLimit:    0.7,
		StrategyLimit: 0.8,
	}
	orchestrator.REE.Orchestrator = orchestrator
	orchestrator.FMRP.Orchestrator = orchestrator
	return orchestrator
}

func (g *GovernanceOrchestrator) DetectRunaway(agent Agent) bool {
	return agent.Entropy() > g.EntropyLimit ||
		agent.LineageDrift() > g.DriftLimit ||
		agent.StrategyVariance() > g.StrategyLimit
}

// Start launches the continuous governance loop.
func (g *GovernanceOrchestrator) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(g.Tick)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.RunCycle(ctx)
			}
		}
	}()
}

// RunCycle executes a single complete decision cycle across the 3-tier system.
func (g *GovernanceOrchestrator) RunCycle(ctx context.Context) FinalVerdict {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. Runaway check: evaluate active agents first
	for _, a := range g.Agents {
		if g.JailController.LeaseManager.IsActive(a.AgentID()) {
			if g.DetectRunaway(a) {
				_ = g.JailController.JailAgent(a, "Orchestrator runaway threshold exceeded")
			}
		}
	}

	// 2. Collect proposals from active agents only
	proposals := []AgentOutput{}
	proposalMap := make(map[string]AgentOutput)
	for _, a := range g.Agents {
		if g.JailController.LeaseManager.IsActive(a.AgentID()) {
			prop := a.GenerateProposal()
			proposals = append(proposals, prop)
			proposalMap[prop.AgentID] = prop
		}
	}

	// 3. Council of 5 arbitrates
	decision := g.Council.Arbitrate(proposals)

	// Route Council jail signals: if ShouldJail returns true for the winning agent proposal
	if winningAgentID, ok := decision.Metadata["winning_agent_id"].(string); ok {
		if winningProp, exists := proposalMap[winningAgentID]; exists {
			if g.Council.ShouldJail(winningProp) {
				for _, a := range g.Agents {
					if a.AgentID() == winningAgentID {
						_ = g.JailController.JailAgent(a, "Council criteria violation (low metrics)")
						break
					}
				}
			}
		}
	}

	// 4. 16 monitors evaluate the decision
	tensor := EvaluateEthicalTensor(decision, g.Monitors)
	verdict := EvaluateVerdict(tensor, g.Monitors)

	// Check if any Monitor triggered ShouldJail
	monitorVeto := false
	for _, m := range g.Monitors {
		if m.ShouldJail(decision) {
			monitorVeto = true
		}
	}

	var finalVerdict FinalVerdict
	// 5. Return final verdict
	if verdict.Passed && !monitorVeto {
		finalVerdict = FinalVerdict{
			Approved: true,
			Decision: decision.FinalDecision,
			Notes:    []string{"All ethical dimensions satisfied"},
		}
	} else {
		finalVerdict = FinalVerdict{
			Approved: false,
			Decision: nil,
			Notes:    append(verdict.FailingDimensions, "Monitor routing trigger or ethical veto"),
		}

		// Ethically vetoed: send the winning agent that proposed the failing decision to jail
		if winningAgentID, ok := decision.Metadata["winning_agent_id"].(string); ok {
			for _, a := range g.Agents {
				if a.AgentID() == winningAgentID {
					// Check if agent is currently on probation
					if ps, onProbation := g.Probation[winningAgentID]; onProbation {
						ps.Violations++
						if ps.Violations > 1 {
							// Return to jail, reset probation status
							_ = g.JailController.JailAgent(a, "Probation violations exceeded threshold")
							delete(g.Probation, winningAgentID)
							break
						}
					}
					_ = g.JailController.JailAgent(a, "Ethical Monitor veto: "+strings.Join(verdict.FailingDimensions, ", "))
					break
				}
			}
		}
	}

	// 6. Update Global Consensus Field state and compute GSR stability
	g.ConsensusField.Mu.Lock()
	g.ConsensusField.CouncilField = decision
	var sumEntropy, sumWeight float64
	var activeCount float64
	for _, a := range g.Agents {
		id := a.AgentID()
		entropy := a.Entropy()
		g.ConsensusField.EntropyField[id] = entropy
		g.ConsensusField.LineageField[id] = a.Snapshot().Lineage
		
		if g.JailController.LeaseManager.IsActive(id) {
			sumEntropy += entropy
			sumWeight += a.Snapshot().Alpha
			activeCount++
		}

		// Sync with registry record state
		if record, exists := g.Registry.Agents[id]; exists {
			_, jailed := g.JailController.Active[id]
			record.InJail = jailed
			_, probation := g.Probation[id]
			record.OnProbation = probation
		}
	}
	g.ConsensusField.Mu.Unlock()

	// Update GSR stability dynamics
	if activeCount > 0 {
		g.GSR.EntropyLevel = sumEntropy / activeCount
		g.GSR.AgenticWeight = sumWeight / activeCount
	}
	g.GSR.ComputeStability()

	// 7. Broadcast verdict back to active agents for feedback loop/learning
	for _, agent := range g.Agents {
		if g.JailController.LeaseManager.IsActive(agent.AgentID()) {
			agent.ReceiveVerdict(finalVerdict, decision)
		}
	}

	return finalVerdict
}

// BroadcastVerdict distributes the decision results back to the agent cohort to enable adaptive self-correction.
func (g *GovernanceOrchestrator) BroadcastVerdict(
	verdict FinalVerdict,
	decision CouncilDecision,
) {
	for _, agent := range g.Agents {
		if g.JailController.LeaseManager.IsActive(agent.AgentID()) {
			agent.ReceiveVerdict(verdict, decision)
		}
	}
}

// AdjustTick scales the cycle heartbeat frequency to adapt to compute loads.
func (g *GovernanceOrchestrator) AdjustTick(cpuLoad float64) {
	scale := 1.0 - cpuLoad
	if scale < 0.1 {
		scale = 0.1
	}
	g.Tick = time.Duration(float64(g.Tick) * scale)
	log.Printf("[GOVERNANCE] Scaled cycle tick duration to: %v", g.Tick)
}
