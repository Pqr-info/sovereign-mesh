package routing

import (
	"context"
	"time"

	"amln-sen/internal/cognition"
	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type ConsensusRouter struct {
	cfg     types.Config
	engine  *cognition.SENEngine
	session *pqr.Session
}

func NewConsensusRouter(cfg types.Config, engine *cognition.SENEngine) *ConsensusRouter {
	return &ConsensusRouter{
		cfg:     cfg,
		engine:  engine,
		session: engine.Session(),
	}
}

// ------------------------------------------------------------
// ConsensusContribution is what this node contributes to the field.
// ------------------------------------------------------------

type ConsensusContribution struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	Vector    []float64 `json:"vector"`
	Alpha     float64   `json:"alpha"`
	Reward    float64   `json:"reward"`
}

// ------------------------------------------------------------
// BuildContribution constructs the node's consensus payload.
// ------------------------------------------------------------

func (c *ConsensusRouter) BuildContribution() ConsensusContribution {
	return ConsensusContribution{
		NodeID:    c.cfg.NodeID,
		Timestamp: time.Now().UTC(),
		Vector:    c.engine.CognitiveVector(),
		Alpha:     c.engine.AgenticWeight(),
		Reward:    c.engine.LastReward(),
	}
}

// ------------------------------------------------------------
// PersistContribution stores the consensus contribution in PQR.
// ------------------------------------------------------------

func (c *ConsensusRouter) PersistContribution(ctx context.Context) error {
	payload := c.BuildContribution()

	data := map[string]interface{}{
		"timestamp": payload.Timestamp.String(),
		"vector":    payload.Vector,
		"alpha":     payload.Alpha,
		"reward":    payload.Reward,
	}

	ticketID, err := c.session.CreateMemory(ctx, "Consensus Contribution", data)
	if err != nil {
		return err
	}

	return c.session.StoreMemory(ctx, ticketID, "state", data)
}

// ------------------------------------------------------------
// Shutdown hook
// ------------------------------------------------------------

func (c *ConsensusRouter) Shutdown() {
	// No persistent resources to close
}
