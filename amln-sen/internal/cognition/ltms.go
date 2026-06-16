package cognition

import (
	"context"
	"math"
	"time"

	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type LTMS struct {
	session *pqr.Session
	cfg     types.Config

	// Cached attractor vector
	Attractor []float64
}

func NewLTMS(session *pqr.Session, cfg types.Config) *LTMS {
	return &LTMS{
		session:   session,
		cfg:       cfg,
		Attractor: []float64{0, 0, 0},
	}
}

// ------------------------------------------------------------
// Update triggers the refresh of historical memory from PQR
// ------------------------------------------------------------

func (l *LTMS) Update(ctx context.Context, stmbVec []float64) {
	_ = l.Refresh(ctx)
}

// ------------------------------------------------------------
// Pull historical memory from PQR and compute attractor
// ------------------------------------------------------------

func (l *LTMS) Refresh(ctx context.Context) error {
	memories, err := l.session.GetAllMemories(ctx)
	if err != nil {
		return err
	}

	if len(memories) == 0 {
		l.Attractor = []float64{0, 0, 0}
		return nil
	}

	// Compute simple attractor:
	// [avg_theta, avg_entropy, avg_txpage_count]
	var sumTheta float64
	var sumEntropy float64
	var sumCount float64

	for _, m := range memories {
		if theta, ok := m["theta"].(float64); ok {
			sumTheta += theta
		}
		if entropy, ok := m["entropy"].(float64); ok {
			sumEntropy += entropy
		}
		if tx, ok := m["tx_pages"].([]interface{}); ok {
			sumCount += float64(len(tx))
		}
	}

	n := float64(len(memories))
	l.Attractor = []float64{
		sumTheta / n,
		sumEntropy / n,
		sumCount / n,
	}

	return nil
}

// ------------------------------------------------------------
// Persist attractor snapshot into PQR as knowledge memory
// ------------------------------------------------------------

func (l *LTMS) Persist(ctx context.Context) error {
	data := map[string]interface{}{
		"timestamp": time.Now().UTC().String(),
		"attractor": l.Attractor,
	}

	ticketID, err := l.session.CreateMemory(ctx, "LTMS Attractor", data)
	if err != nil {
		return err
	}

	return l.session.StoreMemory(ctx, ticketID, "knowledge", data)
}

// ------------------------------------------------------------
// Export LTMS vector for cognitive processing
// ------------------------------------------------------------

func (l *LTMS) Vector() []float64 {
	return l.Attractor
}

// ------------------------------------------------------------
// Compute similarity between STMB and LTMS attractor
// ------------------------------------------------------------

func (l *LTMS) Similarity(stmbVec []float64) float64 {
	if len(stmbVec) != len(l.Attractor) {
		return 0
	}

	var sum float64
	for i := range stmbVec {
		diff := stmbVec[i] - l.Attractor[i]
		sum += diff * diff
	}

	// Inverse distance similarity
	return 1 / (1 + math.Sqrt(sum))
}
