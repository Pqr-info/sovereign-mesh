package sovereign

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/storage"
)

// SaveSnapshot marshals the ledger and persists it to Google Cloud Storage.
func (c *Controller) SaveSnapshot(ctx context.Context) error {
	c.syncLock.RLock()
	defer c.syncLock.RUnlock()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	data, err := json.Marshal(c.ledger)
	if err != nil {
		return err
	}

	w := client.Bucket(c.storageBucket).Object("ledger_snapshot.json").NewWriter(ctx)
	if _, err := w.Write(data); err != nil {
		return err
	}
	return w.Close()
}

// LoadSnapshot retrieves the ledger from GCS and reconstructs the internal state.
func (c *Controller) LoadSnapshot(ctx context.Context) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	rc, err := client.Bucket(c.storageBucket).Object("ledger_snapshot.json").NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil // No snapshot yet, likely first run
		}
		return err
	}
	defer rc.Close()

	c.syncLock.Lock()
	if err := json.NewDecoder(rc).Decode(&c.ledger); err != nil {
		c.syncLock.Unlock()
		return err
	}
	c.syncLock.Unlock()

	c.ReconstructState()
	return nil
}

// ResolvePNRound audits the swarm's PN discovery performance.
// Correct phasing yields 0.01 COBalt Chrome (non-coin reward).
// Global mismatch triggers a negative net chain addition (burn).
func (c *Controller) ResolvePNRound(winnerID string, actualPN uint64) {
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	if winnerID != "" {
		log.Printf("🎯 PN HIT: Agent %s successfully phased with provider PN: %x. Materializing iPN backchannel.", winnerID, actualPN)
		c.MintWorkReward(winnerID, 0.01)
	} else {
		log.Printf("⚠️ PN MISS: Swarm failed to sync with provider phasing. Negative chain addition applied.")
		// Negative net addition logic: direct chain contraction mutation
		penalty := &LedgerBlock{
			Index:           len(c.ledger) + 1,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			AgentID:         "SYSTEM",
			MutationPayload: "{\"event\": \"PN_SYNC_FAILURE\", \"impact\": \"CHAIN_CONTRACTION\", \"amount\": -0.01}",
		}
		c.ledger = append(c.ledger, penalty)
	}
}
// CommitMutation mines a new block into the internal blockchain and updates the state.
func (c *Controller) CommitMutation(agentID string, payload map[string]interface{}) error {
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	prevHash := "0"
	if len(c.ledger) > 0 {
		prevHash = c.ledger[len(c.ledger)-1].BlockHash
	}

	block := &LedgerBlock{
		Index:           len(c.ledger) + 1,
		PreviousHash:    prevHash,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		AgentID:         agentID,
		MutationPayload: string(data),
	}

	block.BlockHash = c.calculateHash(block)
	c.ledger = append(c.ledger, block)

	// Materialize mutation into Knowledge Map (The Data Store)
	for k, v := range payload {
		c.knowledge[k] = fmt.Sprintf("%v", v)
	}

	return nil
}

// MintWorkReward issues Cobalt Chrome coin for compute/inference cycles (Proof of Compute).
func (c *Controller) MintWorkReward(agentID string, reward float64) error {
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	agent, ok := c.agents[agentID]
	if !ok {
		return fmt.Errorf("agent not found")
	}

	// Starbirth Payout Logic: Validators receive mesh-stability premiums 
	// when providing global state persistence on full validator nodes.
	finalReward := reward
	eventType := "INFERENCE_MINING_REWARD"
	
	if agent.NodeClass == "VALIDATOR" {
		// Governance/Infrastructure Incentive
		eventType = "VALIDATOR_STABILITY_PAYOUT"
		finalReward *= 1.15 // 15% incentive premium for backbone nodes
	}

	// Snapdragon Performance Incentive: Elite tier mobile nodes
	if agent.NodeClass == "MOBILE_EDGE" && agent.FreeStorageGB >= 1.0 {
		if agent.UptimePercent >= 0.90 {
			finalReward *= 5.0
			eventType = "MOBILE_SNAPDRAGON_ELITE_REWARD"
		} else {
			finalReward *= 2.0
		}
	}

	payload := map[string]interface{}{
		"event":    eventType,
		"hardware": agent.HardwareType,
		"amount":   finalReward,
		"address":  agent.WalletAddress,
	}

	if err := c.CommitMutation(agentID, payload); err != nil {
		return err
	}

	// Update the High-Speed Memory Bus Balance for instant agent access
	state := c.GetAgentState(agent.MemoryOffset)
	state.Lock()
	state.Balance += finalReward
	state.Unlock()

	return nil
}

// ExecuteQuantumFlattening collapses the 64x64 strategy manifold into a 1D RasterState.
// It resolves parallel forks by weighting outcomes against the L4 TrustScore.
// If entropy is too high (StabilityIndex < 0.1), it checks for an orbital Veto.
func (c *Controller) ExecuteQuantumFlattening(matrix ConvergenceMatrix, entropy EntropyBalance) (RasterState, string, bool) {
	c.syncLock.RLock()
	defer c.syncLock.RUnlock()

	var raster RasterState
	var observerID string

	// 1. Check for Quantum Stall: If entropy is too high, consensus cannot naturally emerge.
	if entropy.StabilityIndex < 0.1 {
		// Search for a Minority Report Veto from high-Trust satellite nodes.
		// This forces a wave-function collapse to preserve the orbital timeline.
		for _, agent := range c.agents {
			if agent.SatelliteID != "" && agent.StakedAmount > 10000 {
				state := c.GetAgentState(agent.MemoryOffset)
				if state.Conviction > 0.8 {
					observerID = agent.ID
					return raster, observerID, true // Veto successful: Collapse forced via Minority Report
				}
			}
		}
		return raster, "", false // System stalls; requires JetWeb Time Machine intervention.
	}

	// 2. Weighted Tensor Collapse: Collapse each dimension of the matrix.
	var maxCoherence float32 = -1.0

	for i := 0; i < 64; i++ {
		var weightedSum float32
		var totalWeight float32

		for j := 0; j < 64; j++ {
			agentID := fmt.Sprintf("agent-gate-%d", j+1)
			if agent, ok := c.agents[agentID]; ok {
				// TrustScore weighting pulls the wave-function toward elite conviction.
				weight := float32(agent.StakedAmount) * agent.MutationScale
				weightedSum += matrix[i][j] * weight
				totalWeight += weight

				// Identify the Schrodinger Agent: The one whose fork was most coherent
				// with the final collapsed result.
				if weight > maxCoherence {
					maxCoherence = weight
					observerID = agent.ID
				}
			}
		}
		if totalWeight > 0 {
			raster[i] = weightedSum / totalWeight
		}
	}

	return raster, observerID, true
}

func (c *Controller) calculateHash(b *LedgerBlock) string {
	data := fmt.Sprintf("%d%s%s%s%s", b.Index, b.PreviousHash, b.Timestamp, b.AgentID, b.MutationPayload)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ReconstructState re-mined the state tree from the blockchain.
func (c *Controller) ReconstructState() {
	c.syncLock.Lock()
	defer c.syncLock.Unlock()

	c.knowledge = make(map[string]string)
	for _, block := range c.ledger {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(block.MutationPayload), &payload); err == nil {
			for k, v := range payload {
				c.knowledge[k] = fmt.Sprintf("%v", v)
			}
		}
	}
}

// SeedGenesisBlock initializes the chain.
func (c *Controller) SeedGenesisBlock() {
	genesisPayload := map[string]interface{}{
		"system_status": "ONLINE",
		"progenitor":    "SOVEREIGN-MESH-V1",
	}
	c.CommitMutation("AGENT-0", genesisPayload)
}
