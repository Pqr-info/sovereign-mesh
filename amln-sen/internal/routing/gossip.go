package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"amln-sen/internal/cognition"
	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

type GossipRouter struct {
	cfg     types.Config
	engine  *cognition.SENEngine
	client  *http.Client
	session *pqr.Session
}

func NewGossipRouter(cfg types.Config, engine *cognition.SENEngine) *GossipRouter {
	return &GossipRouter{
		cfg:     cfg,
		engine:  engine,
		client:  &http.Client{Timeout: 5 * time.Second},
		session: engine.Session(),
	}
}

// ------------------------------------------------------------
// GossipPayload is what we send to peers.
// ------------------------------------------------------------

type GossipPayload struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	Summary   []float64 `json:"summary"`
	Reward    float64   `json:"reward"`
}

// ------------------------------------------------------------
// Push gossip to all configured peers.
// ------------------------------------------------------------

func (g *GossipRouter) Push(ctx context.Context) {
	if len(g.cfg.GossipPeers) == 0 {
		return
	}

	summary := g.engine.GossipSummary()
	reward := g.engine.LastReward()

	payload := GossipPayload{
		NodeID:    g.cfg.NodeID,
		Timestamp: time.Now().UTC(),
		Summary:   summary,
		Reward:    reward,
	}

	body, _ := json.Marshal(payload)

	for _, peer := range g.cfg.GossipPeers {
		url := peer + "/gossip/push"

		req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		_, err := g.client.Do(req)
		if err != nil {
			log.Printf("[GOSSIP] Failed to push to %s: %v", peer, err)
		}
	}
}

// ------------------------------------------------------------
// Receive gossip from a peer.
// ------------------------------------------------------------

func (g *GossipRouter) Receive(ctx context.Context, payload GossipPayload) error {
	// Store gossip summary in PQR
	data := map[string]interface{}{
		"timestamp": payload.Timestamp.String(),
		"peer":      payload.NodeID,
		"summary":   payload.Summary,
		"reward":    payload.Reward,
	}

	ticketID, err := g.session.CreateMemory(ctx, "Gossip Summary", data)
	if err != nil {
		return err
	}

	return g.session.StoreMemory(ctx, ticketID, "custom", data)
}

// ------------------------------------------------------------
// Shutdown hook
// ------------------------------------------------------------

func (g *GossipRouter) Shutdown() {
	// No persistent resources to close
}
