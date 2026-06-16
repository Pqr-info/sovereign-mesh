package cognition

import (
	"context"
	"sync"
	"time"

	"amln-sen/internal/crypto"
	"amln-sen/internal/governance"
	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

// SENEngine orchestrates all cognition layers for a SEN-type AMLN node.
type SENEngine struct {
	mu      sync.RWMutex
	cfg     types.Config
	session *pqr.Session

	// Cognition layers
	stmb    *STMB
	ltms    *LTMS
	hde     *HDE
	prm     *PRM
	sel     *SEL
	vec     *CognitiveVectorBuilder
	weight  *WeightEngine
	lineage *Lineage

	// Crypto
	signer *crypto.Signer

	// Gossip integration
	gossiper Gossiper

	// Evolution Engine
	evolution *EvolutionEngine

	// Council of 5 Governance
	council *governance.CouncilOfFive

	// Governance Orchestrator
	governanceOrchestrator *governance.GovernanceOrchestrator

	// Teleportation Lockout Layer
	teleportationManager *governance.TeleportationManager
}

// NewSENEngine constructs a new cognition engine.
func NewSENEngine(cfg types.Config, session *pqr.Session) (*SENEngine, error) {
	signer, err := crypto.NewSigner()
	if err != nil {
		return nil, err
	}

	stmb := NewSTMB(cfg)
	stmb.SetSession(session)
	
	ltms := NewLTMS(session, cfg)
	
	hde := NewHDE(cfg)
	hde.SetSession(session)

	prm := NewPRM(session, cfg)
	sel := NewSEL(session, cfg)
	vec := NewCognitiveVectorBuilder()
	weight := NewWeightEngine(sel)
	
	lineage := NewLineage(cfg.LineageVectorSize, 0.9) // λ = 0.9

	engine := &SENEngine{
		cfg:     cfg,
		session: session,
		stmb:    stmb,
		ltms:    ltms,
		hde:     hde,
		prm:     prm,
		sel:     sel,
		vec:     vec,
		weight:  weight,
		lineage: lineage,
		signer:  signer,
	}

	evolutionCfg := EvolutionConfig{
		TickBase:          200 * time.Millisecond,
		ExplorationScale:  0.01,
		EnableSyntheticTx: true,
		EnableAdaptive:    true,
	}
	engine.evolution = NewEvolutionEngine(engine, evolutionCfg)

	engine.council = governance.NewDefaultCouncilOfFive()

	engine.governanceOrchestrator = governance.NewGovernanceOrchestrator(cfg.StrategyVectorSize, 1000*time.Millisecond)

	engine.teleportationManager = governance.NewTeleportationManager(engine.governanceOrchestrator.JailController)

	return engine, nil
}

// Session exposes the underlying PQR session.
func (e *SENEngine) Session() *pqr.Session {
	return e.session
}

// ------------------------------------------------------------
// Ingest: TxPages + θ + ε
// ------------------------------------------------------------

func (e *SENEngine) Ingest(ctx context.Context, txPages []map[string]interface{}, theta, entropy float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Update STMB from TxPages
	e.stmb.Update(txPages, theta, entropy)

	// 2. Update LTMS from STMB
	e.ltms.Update(ctx, e.stmb.Vector())

	// 3. Compute HDE from STMB + LTMS
	e.hde.Compute(e.stmb.Vector(), e.ltms.Vector())

	// 4. Compute PRM from STMB + LTMS + HDE
	e.prm.Compute(e.stmb.Vector(), e.ltms.Vector(), e.hde.Vector())

	// 5. Compute reward and update SEL
	reward := e.sel.ComputeReward(
		e.stmb.Vector(),
		e.ltms.Vector(),
		e.hde.Vector(),
		e.prm.Vector(),
	)
	e.sel.UpdateStrategy(reward)

	// 6. Build current Ck, normalize, and update lineage
	Ck := e.vec.BuildCk(
		e.stmb.Vector(),
		e.ltms.Vector(),
		e.hde.Vector(),
		e.prm.Vector(),
		e.sel.Vector(),
	)
	CkNorm := e.vec.NormalizeCk(Ck)
	e.lineage.Update(CkNorm)

	// 7. Persist cognition layers
	_ = e.stmb.Persist(ctx)
	_ = e.ltms.Persist(ctx)
	_ = e.hde.Persist(ctx)
	_ = e.prm.Persist(ctx)
	_ = e.sel.Persist(ctx)

	// Persist lineage + entropy snapshot
	data := map[string]interface{}{
		"lineage": e.lineage.Vector(),
		"entropy": e.hde.Entropy(),
	}
	ticketID, err := e.session.CreateMemory(ctx, "Lineage Snapshot", data)
	if err == nil {
		_ = e.session.StoreMemory(ctx, ticketID, "state", data)
	}
}

// ------------------------------------------------------------
// CognitiveVector returns the normalized Ck vector.
// ------------------------------------------------------------

func (e *SENEngine) CognitiveVector() []float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	Ck := e.vec.BuildCk(
		e.stmb.Vector(),
		e.ltms.Vector(),
		e.hde.Vector(),
		e.prm.Vector(),
		e.sel.Vector(),
	)
	return e.vec.NormalizeCk(Ck)
}

// ------------------------------------------------------------
// AgenticWeight returns α.
// ------------------------------------------------------------

func (e *SENEngine) AgenticWeight() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.weight.ComputeAlpha()
}

// ------------------------------------------------------------
// LastReward exposes SEL's last reward.
// ------------------------------------------------------------

func (e *SENEngine) LastReward() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.sel.LastReward
}

// ------------------------------------------------------------
// GossipSummary returns a compressed summary for gossip.
// ------------------------------------------------------------

func (e *SENEngine) GossipSummary() []float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.sel.Vector()
}

// ------------------------------------------------------------
// SELVector exposes the raw strategy vector.
// ------------------------------------------------------------

func (e *SENEngine) SELVector() []float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.sel.Vector()
}

// ------------------------------------------------------------
// LineageVector returns the lineage attractor.
// ------------------------------------------------------------

func (e *SENEngine) LineageVector() []float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.lineage.Vector()
}

// ------------------------------------------------------------
// Entropy returns the scalar entropy ε.
// ------------------------------------------------------------

func (e *SENEngine) Entropy() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.hde.Entropy()
}

// ------------------------------------------------------------
// MergeStrategy blends local SEL strategy with a peer's.
// ------------------------------------------------------------

func (e *SENEngine) MergeStrategy(weight float64, peerStrategy []float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	local := e.sel.Strategy
	n := len(local)
	if len(peerStrategy) < n {
		n = len(peerStrategy)
	}

	for i := 0; i < n; i++ {
		local[i] = (1-weight)*local[i] + weight*peerStrategy[i]
	}
}

// ------------------------------------------------------------
// MergeLineage blends lineage vectors (simple average).
// ------------------------------------------------------------

func (e *SENEngine) MergeLineage(peerLineage []float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	local := e.lineage.vector
	n := len(local)
	if len(peerLineage) < n {
		n = len(peerLineage)
	}

	for i := 0; i < n; i++ {
		local[i] = 0.5*local[i] + 0.5*peerLineage[i]
	}
}

// ------------------------------------------------------------
// SignedCognition returns the signed cognitive envelope.
// ------------------------------------------------------------

func (e *SENEngine) SignedCognition() (*crypto.CognitiveEnvelope, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	Ck := e.CognitiveVector()
	alpha := e.AgenticWeight()
	return e.signer.SignCognition(e.cfg.NodeID, Ck, alpha)
}

// ------------------------------------------------------------
// Shutdown performs final persistence operations.
// ------------------------------------------------------------
func (e *SENEngine) Shutdown(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_ = e.sel.Persist(ctx)
}

// Start launches the continuous background evolution engine and governance orchestrator.
func (e *SENEngine) Start(ctx context.Context) {
	e.evolution.Start(ctx)
	e.governanceOrchestrator.Start(ctx)
}

// StartBackgroundEvolution delegates to Start for backward compatibility.
func (e *SENEngine) StartBackgroundEvolution(ctx context.Context) {
	e.Start(ctx)
}

// Council returns the internal governance council.
func (e *SENEngine) Council() *governance.CouncilOfFive {
	return e.council
}

// GovernanceOrchestrator returns the internal orchestrator.
func (e *SENEngine) GovernanceOrchestrator() *governance.GovernanceOrchestrator {
	return e.governanceOrchestrator
}

// TeleportationManager returns the internal teleportation manager.
func (e *SENEngine) TeleportationManager() *governance.TeleportationManager {
	return e.teleportationManager
}

