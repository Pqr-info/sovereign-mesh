package cognition

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"time"
)

// StartBackgroundEvolution triggers the self-evolving loop in the background.
func (e *SENEngine) StartBackgroundEvolution(ctx context.Context) {
	go func() {
		// Base interval is 200ms
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Adjust evolution interval dynamically based on compute capacity
				cpuFactor := e.GetComputeCapacity()
				dynamicInterval := time.Duration(200.0 / cpuFactor) * time.Millisecond
				
				// Perform evolution step
				e.EvolveStep(ctx)

				// Adjust ticker dynamically
				ticker.Reset(dynamicInterval)
			}
		}
	}()
}

// GetComputeCapacity returns a scalar scaling factor representing available compute.
// Energy surplus / idle periods -> evolution rates accelerate (factor > 1.0).
func (e *SENEngine) GetComputeCapacity() float64 {
	// In a real system, this queries OS CPU load or GPU metrics.
	// We simulate a stable capacity factor of 1.25 (acceleration during idle surplus).
	return 1.25
}

// EvolveStep runs one tick of autonomous strategy evolution and noise perturbation.
func (e *SENEngine) EvolveStep(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Generate synthetic scenario inputs if no active ingestion occurred
	syntheticPages := e.GenerateSyntheticTxPages()
	
	// Use last known signals as baseline
	theta := e.stmb.RecentTheta
	entropy := e.hde.Entropy()

	// 2. STMB registers simulated inputs
	e.stmb.Update(syntheticPages, theta, entropy)

	// 3. Inject exploration noise into SEL strategy vector
	e.InjectExplorationNoise(entropy)

	// 4. Update memory attractor and compute deltas
	_ = e.ltms.Refresh(ctx)
	
	stmbVec := e.stmb.Vector()
	ltmsVec := e.ltms.Vector()
	e.hde.Compute(stmbVec, ltmsVec)
	hdeVec := e.hde.Vector()
	
	e.prm.Compute(stmbVec, ltmsVec, hdeVec)
	prmVec := e.prm.Vector()

	// 5. Compute adaptive reward and reinforce strategy vector
	reward := e.sel.ComputeReward(stmbVec, ltmsVec, hdeVec, prmVec)
	e.sel.UpdateStrategy(reward)

	// 6. Build new contribution vector C_k and update lineage trajectory
	Ck := e.vec.BuildCk(
		stmbVec,
		ltmsVec,
		hdeVec,
		prmVec,
		e.sel.Vector(),
	)
	CkNorm := e.vec.NormalizeCk(Ck)
	e.lineage.Update(CkNorm)

	// 7. Persist evolution state to the PQR Ticketing memory backend
	go func() {
		bgCtx := context.Background()
		_ = e.sel.Persist(bgCtx)
		
		// Persist lineage snapshot
		data := map[string]interface{}{
			"lineage": e.lineage.Vector(),
			"entropy": e.hde.Entropy(),
			"mode":    "BACKGROUND_EVOLUTION",
		}
		ticketID, err := e.session.CreateMemory(bgCtx, "Autonomous Lineage Update", data)
		if err == nil {
			_ = e.session.StoreMemory(bgCtx, ticketID, "state", data)
		}
	}()
}

// InjectExplorationNoise injects exploration curiosity drift into the strategy vector.
// Low Entropy -> high curiosity (needs to explore alternative pathways).
// High Entropy -> low curiosity (needs to stabilize strategy).
func (e *SENEngine) InjectExplorationNoise(entropy float64) {
	strategy := e.sel.Strategy
	n := len(strategy)
	if n == 0 {
		return
	}

	// Curiosity is inversely proportional to entropy
	curiosity := 1.0 - entropy
	if curiosity < 0.05 {
		curiosity = 0.05 // baseline search pressure
	}

	// Inject noise scaled by curiosity
	noiseScale := 0.02 * curiosity
	for i := range strategy {
		strategy[i] += getRandomFloat() * noiseScale
	}
}

// GenerateSyntheticTxPages simulates hypothetical transaction pages for scenario testing.
func (e *SENEngine) GenerateSyntheticTxPages() []map[string]interface{} {
	// Synthesize transactions with different entropy geometries
	txs := make([]map[string]interface{}, 3)
	for i := 0; i < 3; i++ {
		txs[i] = map[string]interface{}{
			"tx_id":          getRandomInt(),
			"simulated_flow": 0.01 * float64(getRandomInt()%100),
			"attractor_ref":  "synthetic-hyperplane-27",
		}
	}
	return txs
}

// Helpers
func getRandomFloat() float64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return 0.0
	}
	return (float64(nBig.Int64()) / 1000.0) - 0.5
}

func getRandomInt() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int64(binary.BigEndian.Uint64(b[:]))
}
