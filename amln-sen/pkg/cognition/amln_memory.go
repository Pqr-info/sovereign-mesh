package cognition

import (
	"context"
	"errors"
	"sync"
	"time"
)

// AMLNMemoryManager manages the dual-partition memory architecture (short-term vs. long-term) (CC-3)
type AMLNMemoryManager struct {
	mu            sync.RWMutex
	stmb          *STMB
	ltms          *LTMS
	memoryWindow  time.Duration
	lastUpdated   time.Time
}

func NewAMLNMemoryManager(stmb *STMB, ltms *LTMS, window time.Duration) *AMLNMemoryManager {
	return &AMLNMemoryManager{
		stmb:         stmb,
		ltms:         ltms,
		memoryWindow: window,
		lastUpdated:  time.Now(),
	}
}

// UpdateMemory performs memory migration: STMB -> HDE -> LTMS Attractor partition mapping
func (amm *AMLNMemoryManager) UpdateMemory(ctx context.Context, txPages []map[string]interface{}, theta, entropy float64) error {
	amm.mu.Lock()
	defer amm.mu.Unlock()

	// 1. Update Short Term Memory Buffer
	amm.stmb.Update(txPages, theta, entropy)

	// 2. Persist STMB into PQR context partition
	err := amm.stmb.Persist(ctx)
	if err != nil {
		return err
	}

	// 3. Update Long Term Memory Store attractor
	amm.ltms.Update(ctx, amm.stmb.Vector())

	// 4. Persist LTMS attractor
	err = amm.ltms.Persist(ctx)
	if err != nil {
		return err
	}

	amm.lastUpdated = time.Now()
	return nil
}

// GetMemoryVector returns combined cognitive components
func (amm *AMLNMemoryManager) GetMemoryVector() ([]float64, error) {
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	stmbVec := amm.stmb.Vector()
	ltmsVec := amm.ltms.Vector()

	if len(stmbVec) == 0 || len(ltmsVec) == 0 {
		return nil, errors.New("AMLN memory not initialized")
	}

	// Combined short + long term representation
	return append(stmbVec, ltmsVec...), nil
}
