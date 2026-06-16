package cognition

import (
	"context"
	"math"
	"time"

	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type HDE struct {
	session *pqr.Session
	cfg     types.Config

	vector  []float64
	mag     float64
	entropy float64
}

func NewHDE(cfg types.Config) *HDE {
	return &HDE{
		cfg:    cfg,
		vector: make([]float64, cfg.LineageVectorSize),
	}
}

// SetSession attaches the session dynamically when available
func (h *HDE) SetSession(session *pqr.Session) {
	h.session = session
}

// ------------------------------------------------------------
// Compute hypothesis deltas and scalar entropy ε
// ------------------------------------------------------------

func (h *HDE) Compute(stmbVec, ltmsVec []float64) {
	n := len(h.vector)
	if len(stmbVec) < n {
		n = len(stmbVec)
	}
	if len(ltmsVec) < n {
		n = len(ltmsVec)
	}

	var sumSq float64
	for i := 0; i < n; i++ {
		d := stmbVec[i] - ltmsVec[i]
		h.vector[i] = d
		sumSq += d * d
	}
	h.mag = math.Sqrt(sumSq)

	// Entropy ε: normalized magnitude in [0,1)
	// ε = ||Δ|| / (1 + ||Δ||)
	h.entropy = h.mag / (1 + h.mag)
}

// ------------------------------------------------------------
// Persist hypothesis delta into PQR as custom memory
// ------------------------------------------------------------

func (h *HDE) Persist(ctx context.Context) error {
	if h.session == nil {
		return nil
	}
	data := map[string]interface{}{
		"timestamp": time.Now().UTC().String(),
		"delta":     h.vector,
		"entropy":   h.entropy,
	}

	ticketID, err := h.session.CreateMemory(ctx, "HDE Hypothesis Delta", data)
	if err != nil {
		return err
	}

	return h.session.StoreMemory(ctx, ticketID, "custom", data)
}

// ------------------------------------------------------------
// Export HDE vector for cognitive processing
// ------------------------------------------------------------

func (h *HDE) Vector() []float64 {
	return h.vector
}

// ------------------------------------------------------------
// Predictive magnitude
// ------------------------------------------------------------

func (h *HDE) Magnitude() float64 {
	return h.mag
}

// ------------------------------------------------------------
// Entropy ε
// ------------------------------------------------------------

func (h *HDE) Entropy() float64 {
	return h.entropy
}
