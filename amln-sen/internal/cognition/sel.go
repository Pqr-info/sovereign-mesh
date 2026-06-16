package cognition

import (
	"context"
	"math"
	"math/rand"
	"time"

	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type SEL struct {
	session *pqr.Session
	cfg     types.Config

	// Strategy vector S(t)
	Strategy []float64

	// Last computed reward
	LastReward float64
}

func NewSEL(session *pqr.Session, cfg types.Config) *SEL {
	// Initialize strategy vector with zeros
	strategy := make([]float64, cfg.StrategyVectorSize)

	return &SEL{
		session:  session,
		cfg:      cfg,
		Strategy: strategy,
	}
}

// ------------------------------------------------------------
// Compute reward signal for reinforcement learning
// ------------------------------------------------------------

func (s *SEL) ComputeReward(stmbVec, ltmsVec, hdeVec, prmVec []float64) float64 {
	// Reward components:
	// 1. Similarity between STMB and LTMS (stability)
	// 2. Negative magnitude of HDE (penalize uncertainty)
	// 3. PRM pattern strength (pattern coherence)

	// 1. Stability reward
	var stability float64
	for i := range stmbVec {
		diff := stmbVec[i] - ltmsVec[i]
		stability += 1 / (1 + math.Abs(diff))
	}
	stability /= float64(len(stmbVec))

	// 2. Predictive penalty
	var hdeMag float64
	for _, v := range hdeVec {
		hdeMag += v * v
	}
	hdeMag = math.Sqrt(hdeMag)

	// 3. Pattern strength
	var patternStrength float64
	for _, v := range prmVec {
		patternStrength += math.Abs(v)
	}
	patternStrength /= float64(len(prmVec))

	// Final reward
	reward := stability + patternStrength - hdeMag
	s.LastReward = reward

	return reward
}

// ------------------------------------------------------------
// Update strategy vector using reinforcement learning
// ------------------------------------------------------------

func (s *SEL) UpdateStrategy(reward float64) {
	for i := range s.Strategy {
		// Reinforcement update:
		// S(t+1) = S(t) + learning_rate * reward
		s.Strategy[i] += s.cfg.LearningRate * reward

		// Apply decay to prevent runaway growth
		s.Strategy[i] *= s.cfg.RewardDecay
	}
}

// ------------------------------------------------------------
// Persist strategy vector into PQR as state memory
// ------------------------------------------------------------

func (s *SEL) Persist(ctx context.Context) error {
	data := map[string]interface{}{
		"timestamp": time.Now().UTC().String(),
		"strategy":  s.Strategy,
		"reward":    s.LastReward,
	}

	ticketID, err := s.session.CreateMemory(ctx, "SEL Strategy Vector", data)
	if err != nil {
		return err
	}

	return s.session.StoreMemory(ctx, ticketID, "state", data)
}

// ------------------------------------------------------------
// Export SEL vector for cognitive processing
// ------------------------------------------------------------

func (s *SEL) Vector() []float64 {
	return s.Strategy
}

// ------------------------------------------------------------
// Agentic weight α = normalized magnitude of strategy vector
// ------------------------------------------------------------

func (s *SEL) AgenticWeight() float64 {
	var sum float64
	for _, v := range s.Strategy {
		sum += v * v
	}
	mag := math.Sqrt(sum)

	// Normalize to [0,1]
	return mag / (1 + mag)
}

// InjectExplorationNoise injects noise to perturb the strategy vector.
func (s *SEL) InjectExplorationNoise(scale float64) {
	for i := range s.Strategy {
		noise := (rand.Float64() - 0.5) * scale
		s.Strategy[i] += noise
	}
}
