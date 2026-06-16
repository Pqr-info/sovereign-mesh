package routing

import (
	"context"
	"time"

	"amln-sen/internal/cognition"
	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type SlingshotRouter struct {
	cfg     types.Config
	engine  *cognition.SENEngine
	session *pqr.Session
}

func NewSlingshotRouter(cfg types.Config, engine *cognition.SENEngine) *SlingshotRouter {
	return &SlingshotRouter{
		cfg:     cfg,
		engine:  engine,
		session: engine.Session(),
	}
}

// ------------------------------------------------------------
// SlingshotManifest is the structure exchanged during merges.
// ------------------------------------------------------------

type SlingshotManifest struct {
	NodeID     string    `json:"node_id"`
	Timestamp  time.Time `json:"timestamp"`
	Lineage    []float64 `json:"lineage"`
	Entropy    float64   `json:"entropy"`
	Alpha      float64   `json:"alpha"`
	Strategy   []float64 `json:"strategy"`
	MeshIsland string    `json:"mesh_island"`
}

// ------------------------------------------------------------
// BuildManifest constructs the node's merge payload.
// ------------------------------------------------------------

func (s *SlingshotRouter) BuildManifest() SlingshotManifest {
	return SlingshotManifest{
		NodeID:     s.cfg.NodeID,
		Timestamp:  time.Now().UTC(),
		Lineage:    s.engine.LineageVector(),
		Entropy:    s.engine.Entropy(),
		Alpha:      s.engine.AgenticWeight(),
		Strategy:   s.engine.SELVector(),
		MeshIsland: s.cfg.MeshIslandID,
	}
}

// ------------------------------------------------------------
// MergeManifest blends a peer's manifest into local state.
// ------------------------------------------------------------

func (s *SlingshotRouter) MergeManifest(ctx context.Context, m SlingshotManifest) error {
	// Only merge if slingshot is enabled
	if !s.cfg.SlingshotEnabled {
		return nil
	}

	// Mesh-island check: only merge if same island
	if m.MeshIsland != s.cfg.MeshIslandID {
		return nil
	}

	// Merge strategy vectors using weighted average:
	// weight = peer_alpha / (local_alpha + peer_alpha)
	localAlpha := s.engine.AgenticWeight()
	peerAlpha := m.Alpha

	weight := peerAlpha / (localAlpha + peerAlpha + 1e-9)

	s.engine.MergeStrategy(weight, m.Strategy)

	// Merge lineage (simple average)
	s.engine.MergeLineage(m.Lineage)

	// Persist merge event in PQR
	data := map[string]interface{}{
		"timestamp": m.Timestamp.String(),
		"peer":      m.NodeID,
		"alpha":     m.Alpha,
		"entropy":   m.Entropy,
		"strategy":  m.Strategy,
		"lineage":   m.Lineage,
	}

	ticketID, err := s.session.CreateMemory(ctx, "Slingshot Merge", data)
	if err != nil {
		return err
	}

	return s.session.StoreMemory(ctx, ticketID, "state", data)
}

// ------------------------------------------------------------
// Shutdown hook
// ------------------------------------------------------------

func (s *SlingshotRouter) Shutdown() {
	// No persistent resources to close
}
