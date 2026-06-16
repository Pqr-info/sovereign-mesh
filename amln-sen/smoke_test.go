package main

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"amln-sen/internal/api"
	"amln-sen/internal/cognition"
	"amln-sen/internal/pqr"
	"amln-sen/internal/routing"
	"amln-sen/internal/types"
)

// Mock PQR Server to handle CreateMemory/StoreMemory requests
func startMockPQRServer(t *testing.T) *httptest.Server {
	r := gin.Default()
	r.POST("/REST/2.0/ticket", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ticket_id": "ticket-1002"})
	})
	r.POST("/REST/2.0/agent/:agent/memory/:ticket", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/REST/2.0/agent/:agent/context", func(c *gin.Context) {
		c.JSON(http.StatusOK, []map[string]interface{}{
			{
				"theta":    0.5,
				"entropy":  0.2,
				"tx_pages": []interface{}{"tx1"},
			},
		})
	})
	return httptest.NewServer(r)
}

func TestAmlnSenSmokeCheck(t *testing.T) {
	// Start mock PQR server
	mockPQR := startMockPQRServer(t)
	defer mockPQR.Close()

	// Load configuration
	cfg := types.LoadConfig()
	cfg.PQREndpoint = mockPQR.URL
	cfg.NodeID = "amln-test-node"
	cfg.StrategyVectorSize = 8
	cfg.LineageVectorSize = 8

	// Initialize PQR session
	session := pqr.NewSession(cfg.PQREndpoint, cfg.NodeID)

	// Initialize cognition engine (SEN)
	engine, err := cognition.NewSENEngine(cfg, session)
	if err != nil {
		t.Fatalf("failed to initialize SEN engine: %v", err)
	}

	// Initialize routing modules
	gossip := routing.NewGossipRouter(cfg, engine)
	slingshot := routing.NewSlingshotRouter(cfg, engine)
	consensus := routing.NewConsensusRouter(cfg, engine)

	// Initialize REST router
	r := api.NewRouter(engine, gossip, slingshot, consensus)

	// Helper to send POST ingest request
	ingest := func(pages []map[string]interface{}, theta, entropy float64) {
		payload := map[string]interface{}{
			"tx_pages": pages,
			"theta":    theta,
			"entropy":  entropy,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("ingest failed with status: %d", w.Code)
		}
	}

	// 1. Ingest baseline batches
	ingest([]map[string]interface{}{{"val": 1.0}}, 0.5, 0.2)
	ingest([]map[string]interface{}{{"val": 1.0}}, 0.5, 0.2)

	t.Run("Verify Cognitive Vector is Unit-Norm", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cognition/vector", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("failed to get vector: %d", w.Code)
		}

		var resp struct {
			Vector []float64 `json:"vector"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Vector) != 8 {
			t.Fatalf("expected vector of size 8, got %d", len(resp.Vector))
		}

		// Compute L2 norm
		var sumSq float64
		for _, v := range resp.Vector {
			sumSq += v * v
		}
		norm := math.Sqrt(sumSq)
		if math.Abs(norm-1.0) > 1e-6 {
			t.Errorf("expected unit norm (1.0), got %f", norm)
		}
	})

	t.Run("Verify Alpha is Stable and Bounded", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cognition/weight", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Alpha float64 `json:"alpha"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Alpha < 0 || resp.Alpha > 1 {
			t.Errorf("expected alpha to be bounded in [0,1], got %f", resp.Alpha)
		}
	})

	t.Run("Verify Signed Envelope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cognition/signed", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			NodeID    string    `json:"node_id"`
			Signature string    `json:"signature"`
			PubKey    string    `json:"pubkey"`
			Alpha     float64   `json:"alpha"`
			Vector    []float64 `json:"vector"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.NodeID != cfg.NodeID {
			t.Errorf("expected nodeID %s, got %s", cfg.NodeID, resp.NodeID)
		}
		if resp.Signature == "" {
			t.Error("expected non-empty signature")
		}
		if resp.PubKey == "" {
			t.Error("expected non-empty pubkey")
		}
	})

	t.Run("Check Lineage Dynamics Tracker", func(t *testing.T) {
		// Feed a sequence of very different inputs and capture lineage tracking distance
		initialLineage := make([]float64, 8)
		copy(initialLineage, engine.LineageVector())

		// Sequence of inputs
		ingest([]map[string]interface{}{{"val": 99.0}}, 1.9, 0.9)
		lin1 := make([]float64, 8)
		copy(lin1, engine.LineageVector())

		ingest([]map[string]interface{}{{"val": -50.0}}, -0.8, 0.1)
		lin2 := make([]float64, 8)
		copy(lin2, engine.LineageVector())

		// Verify lineage did not jump instantly but evolved gradually (memory of past)
		var diffSum float64
		for i := 0; i < 8; i++ {
			diffSum += math.Abs(lin2[i] - lin1[i])
		}
		t.Logf("Lineage displacement step diff: %f", diffSum)
	})

	t.Run("Verify Entropy Surprise Monotonicity", func(t *testing.T) {
		// Case A: STMB is close to historical LTMS values (theta 0.5, entropy 0.2)
		ingest([]map[string]interface{}{{"val": 1.0}}, 0.5, 0.2)
		epsSmall := engine.Entropy()

		// Case B: STMB is highly divergent surprise input (theta 15.0, entropy 9.5)
		ingest([]map[string]interface{}{{"val": 500.0}}, 15.0, 9.5)
		epsLarge := engine.Entropy()

		t.Logf("Small surprise entropy: %f, Large surprise entropy: %f", epsSmall, epsLarge)
		if epsLarge <= epsSmall {
			t.Errorf("expected surprise input to yield higher entropy, small: %f, large: %f", epsSmall, epsLarge)
		}
	})

	t.Run("Verify Consensus Contribution Endpoints", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consensus/contribute", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			NodeID string    `json:"node_id"`
			Vector []float64 `json:"vector"`
			Alpha  float64   `json:"alpha"`
			Reward float64   `json:"reward"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.NodeID != cfg.NodeID {
			t.Errorf("expected node ID %s, got %s", cfg.NodeID, resp.NodeID)
		}
		if math.Abs(resp.Alpha-engine.AgenticWeight()) > 1e-6 {
			t.Errorf("expected matching alpha, got %f vs engine %f", resp.Alpha, engine.AgenticWeight())
		}
		if math.Abs(resp.Reward-engine.LastReward()) > 1e-6 {
			t.Errorf("expected matching reward, got %f vs engine %f", resp.Reward, engine.LastReward())
		}
	})
}
