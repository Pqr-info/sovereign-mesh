package cognition

import (
	"context"
	"math"
	"time"

	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type PRM struct {
	session *pqr.Session
	cfg     types.Config

	// Last computed pattern activation vector
	Activation []float64
}

func NewPRM(session *pqr.Session, cfg types.Config) *PRM {
	return &PRM{
		session:    session,
		cfg:        cfg,
		Activation: []float64{0, 0, 0},
	}
}

// ------------------------------------------------------------
// Compute pattern activation using sliding-window correlation
// ------------------------------------------------------------

func (p *PRM) Compute(stmbVec, ltmsVec, hdeVec []float64) {
	// All vectors must be same length
	n := len(stmbVec)
	if len(ltmsVec) != n || len(hdeVec) != n {
		p.Activation = []float64{0, 0, 0}
		return
	}

	activation := make([]float64, n)

	for i := 0; i < n; i++ {
		// Simple pattern activation:
		// activation[i] = correlation(STMB[i], LTMS[i], HDE[i])
		activation[i] = p.patternCorrelation(stmbVec[i], ltmsVec[i], hdeVec[i])
	}

	p.Activation = activation
}

// ------------------------------------------------------------
// Pattern correlation function
// ------------------------------------------------------------

func (p *PRM) patternCorrelation(a, b, c float64) float64 {
	// Normalize values
	norm := func(x float64) float64 {
		return x / (1 + math.Abs(x))
	}

	na := norm(a)
	nb := norm(b)
	nc := norm(c)

	// Correlation = average of pairwise products
	return (na*nb + nb*nc + na*nc) / 3
}

// ------------------------------------------------------------
// Persist pattern activation into PQR as custom memory
// ------------------------------------------------------------

func (p *PRM) Persist(ctx context.Context) error {
	data := map[string]interface{}{
		"timestamp":  time.Now().UTC().String(),
		"activation": p.Activation,
	}

	ticketID, err := p.session.CreateMemory(ctx, "PRM Pattern Activation", data)
	if err != nil {
		return err
	}

	return p.session.StoreMemory(ctx, ticketID, "custom", data)
}

// ------------------------------------------------------------
// Export PRM vector for cognitive processing
// ------------------------------------------------------------

func (p *PRM) Vector() []float64 {
	return p.Activation
}

// ------------------------------------------------------------
// Pattern strength (used by SEL reward shaping)
// ------------------------------------------------------------

func (p *PRM) Strength() float64 {
	var sum float64
	for _, v := range p.Activation {
		sum += math.Abs(v)
	}
	return sum / float64(len(p.Activation))
}
