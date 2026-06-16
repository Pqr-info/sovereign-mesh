package cognition

import (
	"context"
	"math/rand"
	"runtime"
	"time"
)

// EvolutionEngine drives continuous background evolution.
type EvolutionEngine struct {
	engine *SENEngine
	cfg    EvolutionConfig
}

// EvolutionConfig controls evolution behavior.
type EvolutionConfig struct {
	TickBase          time.Duration // base tick (e.g., 200ms)
	ExplorationScale  float64       // noise amplitude
	EnableSyntheticTx bool          // allow synthetic research
	EnableAdaptive    bool          // adapt tick to CPU load
}

// NewEvolutionEngine constructs the evolution subsystem.
func NewEvolutionEngine(engine *SENEngine, cfg EvolutionConfig) *EvolutionEngine {
	return &EvolutionEngine{
		engine: engine,
		cfg:    cfg,
	}
}

// Start launches the continuous evolution loop.
func (ev *EvolutionEngine) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(ev.cfg.TickBase)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ev.step(ctx)

				// Adjust tick if adaptive mode is enabled
				if ev.cfg.EnableAdaptive {
					load := ev.cpuLoad()
					newTick := ev.adaptTick(load)
					ticker.Reset(newTick)
				}
			}
		}
	}()
}

// ------------------------------------------------------------
// Evolution Step
// ------------------------------------------------------------

func (ev *EvolutionEngine) step(ctx context.Context) {
	ev.engine.mu.Lock()
	defer ev.engine.mu.Unlock()

	// 1. Exploration noise
	ev.engine.sel.InjectExplorationNoise(ev.cfg.ExplorationScale)

	// 2. Optional synthetic research
	if ev.cfg.EnableSyntheticTx {
		synth := ev.engine.GenerateSyntheticTx()
		theta := ev.engine.stmb.RecentTheta
		entropy := ev.engine.hde.Entropy()
		ev.engine.stmb.Update(synth, theta, entropy)
	}

	// Refresh memory attractor
	_ = ev.engine.ltms.Refresh(ctx)

	stmbVec := ev.engine.stmb.Vector()
	ltmsVec := ev.engine.ltms.Vector()
	ev.engine.hde.Compute(stmbVec, ltmsVec)
	hdeVec := ev.engine.hde.Vector()
	
	ev.engine.prm.Compute(stmbVec, ltmsVec, hdeVec)
	prmVec := ev.engine.prm.Vector()

	// 3. Internal reward update
	reward := ev.engine.sel.ComputeReward(
		stmbVec,
		ltmsVec,
		hdeVec,
		prmVec,
	)
	ev.engine.sel.UpdateStrategy(reward)

	// 4. Update lineage
	Ck := ev.engine.vec.BuildCk(
		stmbVec,
		ltmsVec,
		hdeVec,
		prmVec,
		ev.engine.sel.Vector(),
	)
	CkNorm := ev.engine.vec.NormalizeCk(Ck)
	ev.engine.lineage.Update(CkNorm)

	// 5. Persist lineage + entropy
	go func() {
		bgCtx := context.Background()
		_ = ev.engine.sel.Persist(bgCtx)
		data := map[string]interface{}{
			"lineage": ev.engine.lineage.Vector(),
			"entropy": ev.engine.hde.Entropy(),
			"reward":  reward,
			"mode":    "BACKGROUND_EVOLUTION",
		}
		ticketID, err := ev.engine.session.CreateMemory(bgCtx, "Background Evolution", data)
		if err == nil {
			_ = ev.engine.session.StoreMemory(bgCtx, ticketID, "state", data)
		}
	}()
}

// ------------------------------------------------------------
// Synthetic Research
// ------------------------------------------------------------

func (e *SENEngine) GenerateSyntheticTx() []map[string]interface{} {
	base := e.stmb.Vector()
	perturbed := make([]float64, len(base))

	for i := range base {
		perturbed[i] = base[i] + (rand.Float64()-0.5)*0.1
	}

	return []map[string]interface{}{
		{
			"synthetic": true,
			"vector":    perturbed,
		},
	}
}

// ------------------------------------------------------------
// CPU Load & Adaptive Tick
// ------------------------------------------------------------

func (ev *EvolutionEngine) cpuLoad() float64 {
	// Simple heuristic: number of goroutines vs. CPU cores
	g := float64(runtime.NumGoroutine())
	c := float64(runtime.NumCPU())
	load := g / (10 * c)
	if load > 1 {
		load = 1
	}
	return load
}

func (ev *EvolutionEngine) adaptTick(load float64) time.Duration {
	scale := 1.0 - load
	if scale < 0.1 {
		scale = 0.1
	}
	return time.Duration(float64(ev.cfg.TickBase) * scale)
}

// ------------------------------------------------------------
// Gossip system integration
// ------------------------------------------------------------

// Gossiper defines the contract for pushing gossip messages.
type Gossiper interface {
	Push(ctx context.Context)
}

// RegisterGossiper registers a gossip routing component with the engine.
func (e *SENEngine) RegisterGossiper(g Gossiper) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.gossiper = g
}

// StartContinuousGossip runs a background loop to continuously broadcast gossip summaries to peers.
func (e *SENEngine) StartContinuousGossip(ctx context.Context) {
	go func() {
		tick := e.cfg.GossipTick
		if tick <= 0 {
			tick = 1000 * time.Millisecond
		}
		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				e.mu.RLock()
				g := e.gossiper
				e.mu.RUnlock()
				if g != nil {
					g.Push(ctx)
				}
			}
		}
	}()
}
