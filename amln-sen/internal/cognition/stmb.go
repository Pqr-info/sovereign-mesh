package cognition

import (
	"context"
	"time"

	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type STMB struct {
	session *pqr.Session
	cfg     types.Config

	// Local cache of recent signals
	RecentTxPages []map[string]interface{}
	RecentTheta   float64
	RecentEntropy float64
}

func NewSTMB(cfg types.Config) *STMB {
	return &STMB{
		cfg:           cfg,
		RecentTxPages: []map[string]interface{}{},
	}
}

// ------------------------------------------------------------
// Update: Ingest new short-term signals
// ------------------------------------------------------------

func (s *STMB) Update(txPages []map[string]interface{}, theta float64, entropy float64) {
	s.RecentTxPages = txPages
	s.RecentTheta = theta
	s.RecentEntropy = entropy
}

// SetSession attaches the session dynamically when available
func (s *sTMB) SetSession(session *pqr.Session) {
	s.session = session
}

type sTMB = STMB

// ------------------------------------------------------------
// Persist STMB snapshot into PQR as context memory
// ------------------------------------------------------------

func (s *STMB) Persist(ctx context.Context) error {
	if s.session == nil {
		return nil
	}
	data := map[string]interface{}{
		"timestamp": time.Now().UTC().String(),
		"tx_pages":  s.RecentTxPages,
		"theta":     s.RecentTheta,
		"entropy":   s.RecentEntropy,
	}

	// Create a new memory ticket
	ticketID, err := s.session.CreateMemory(ctx, "STMB Snapshot", data)
	if err != nil {
		return err
	}

	// Store as context memory
	return s.session.StoreMemory(ctx, ticketID, "context", data)
}

// ------------------------------------------------------------
// Export STMB vector for cognitive processing
// ------------------------------------------------------------

func (s *STMB) Vector() []float64 {
	// Minimal vectorization:
	// [theta, entropy, len(txPages)]
	return []float64{
		s.RecentTheta,
		s.RecentEntropy,
		float64(len(s.RecentTxPages)),
	}
}
