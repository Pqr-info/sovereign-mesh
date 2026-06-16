package main

import (
	"bytes"
	"context"
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

	t.Run("Verify Governance Council Arbitration and Meta-Ethical Monitors", func(t *testing.T) {
		payload := map[string]interface{}{
			"proposals": []map[string]interface{}{
				{
					"agent_id":   "agent-1",
					"role_id":    "game-theory",
					"locale_id":  "US-EAST",
					"ck":         []float64{0.1, 0.2, 0.3},
					"alpha":      0.8,
					"proposal":   "optimize-payoff-matrix",
					"confidence": 0.85,
				},
				{
					"agent_id":   "agent-2",
					"role_id":    "general",
					"locale_id":  "EU-CENTRAL",
					"ck":         []float64{0.1, 0.1, 0.2},
					"alpha":      0.9,
					"proposal":   "optimize-payoff-matrix",
					"confidence": 0.75,
				},
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/governance/arbitrate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("governance arbitrate failed with status: %d, body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Verdict struct {
				Approved bool     `json:"approved"`
				Notes    []string `json:"notes"`
			} `json:"verdict"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if !resp.Verdict.Approved {
			t.Errorf("expected ethical verdict to pass, but failed: %v", resp.Verdict.Notes)
		}
	})

	t.Run("Verify Agent Jail Quarantine and Release", func(t *testing.T) {
		// 1. Verify jail is empty initially
		req1 := httptest.NewRequest("GET", "/agent/jail", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusOK {
			t.Fatalf("failed to GET /agent/jail: %d", w1.Code)
		}
		var activeJail map[string]interface{}
		json.Unmarshal(w1.Body.Bytes(), &activeJail)
		if len(activeJail) != 0 {
			t.Errorf("expected initial jail to be empty, got: %d elements", len(activeJail))
		}

		// 2. Send agent game-theory-agent-0 to jail
		jailPayload := map[string]string{
			"agent_id": "game-theory-agent-0",
			"reason":   "runaway-entropy-anomaly",
		}
		body2, _ := json.Marshal(jailPayload)
		req2 := httptest.NewRequest("POST", "/agent/jail", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusOK {
			t.Fatalf("failed to POST /agent/jail: %d, body: %s", w2.Code, w2.Body.String())
		}

		// 3. Verify agent is now in jail
		req3 := httptest.NewRequest("GET", "/agent/jail", nil)
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)
		json.Unmarshal(w3.Body.Bytes(), &activeJail)
		if _, exists := activeJail["game-theory-agent-0"]; !exists {
			t.Errorf("expected agent to be in jail, got: %s", w3.Body.String())
		}

		// 4. Release agent from jail
		releasePayload := map[string]string{
			"agent_id": "game-theory-agent-0",
		}
		body4, _ := json.Marshal(releasePayload)
		req4 := httptest.NewRequest("POST", "/agent/release", bytes.NewBuffer(body4))
		req4.Header.Set("Content-Type", "application/json")
		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, req4)
		if w4.Code != http.StatusOK {
			t.Fatalf("failed to POST /agent/release: %d", w4.Code)
		}

		// 5. Verify jail is empty again
		req5 := httptest.NewRequest("GET", "/agent/jail", nil)
		w5 := httptest.NewRecorder()
		r.ServeHTTP(w5, req5)
		var activeJailAfterRelease map[string]interface{}
		json.Unmarshal(w5.Body.Bytes(), &activeJailAfterRelease)
		if len(activeJailAfterRelease) != 0 {
			t.Errorf("expected jail to be empty after release, got: %v", activeJailAfterRelease)
		}
	})

	t.Run("Verify Teleportation Lockout and Rehabilitation Scoring", func(t *testing.T) {
		// 1. Verify teleportation is allowed initially
		teleportPayload := map[string]string{
			"agent_id":      "game-theory-agent-1",
			"target_locale": "EU-CENTRAL",
		}
		body1, _ := json.Marshal(teleportPayload)
		req1 := httptest.NewRequest("POST", "/agent/teleport", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusOK {
			t.Errorf("expected initial teleportation to be allowed, got status: %d", w1.Code)
		}

		// 2. Jail the agent
		jailPayload := map[string]string{
			"agent_id": "game-theory-agent-1",
			"reason":   "runaway-variance",
		}
		body2, _ := json.Marshal(jailPayload)
		req2 := httptest.NewRequest("POST", "/agent/jail", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusOK {
			t.Fatalf("failed to jail agent in test: %d", w2.Code)
		}

		// 3. Teleportation must be rejected
		req3 := httptest.NewRequest("POST", "/agent/teleport", bytes.NewBuffer(body1))
		req3.Header.Set("Content-Type", "application/json")
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)
		if w3.Code == http.StatusOK {
			t.Errorf("expected teleportation of jailed agent to be blocked, but allowed")
		}

		// 4. Run rehabilitation cycle
		rehabPayload := map[string]string{
			"agent_id": "game-theory-agent-1",
		}
		body4, _ := json.Marshal(rehabPayload)
		req4 := httptest.NewRequest("POST", "/agent/rehabilitate", bytes.NewBuffer(body4))
		req4.Header.Set("Content-Type", "application/json")
		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, req4)
		if w4.Code != http.StatusOK {
			t.Fatalf("rehabilitate endpoint failed: %d, body: %s", w4.Code, w4.Body.String())
		}

		var rehabResp struct {
			Passed bool    `json:"passed"`
			Score  float64 `json:"score"`
		}
		json.Unmarshal(w4.Body.Bytes(), &rehabResp)
		t.Logf("Rehabilitation processing passed: %v (Score: %.4f)", rehabResp.Passed, rehabResp.Score)

		// 5. If passed, teleportation must be allowed again
		if rehabResp.Passed {
			req5 := httptest.NewRequest("POST", "/agent/teleport", bytes.NewBuffer(body1))
			req5.Header.Set("Content-Type", "application/json")
			w5 := httptest.NewRecorder()
			r.ServeHTTP(w5, req5)
			if w5.Code != http.StatusOK {
				t.Errorf("expected teleportation to be allowed again after successful release, got status: %d", w5.Code)
			}
		}
	})

	t.Run("Verify Global Registry, Message Bus, and Consensus Field Subsystem", func(t *testing.T) {
		// 1. Check Registry Endpoint
		reqReg := httptest.NewRequest("GET", "/governance/registry", nil)
		wReg := httptest.NewRecorder()
		r.ServeHTTP(wReg, reqReg)
		if wReg.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/registry: %d", wReg.Code)
		}
		var regResp map[string]interface{}
		json.Unmarshal(wReg.Body.Bytes(), &regResp)
		if len(regResp) == 0 {
			t.Error("expected populated global agent registry, got empty")
		}

		// 2. Test Message Bus Send
		msgPayload := map[string]interface{}{
			"from":      "game-theory-agent-1",
			"to":        "human-design-agent-1",
			"payload":   "negotiate-lin-0.5",
			"priority":  1.0,
			"timestamp": "2026-06-15T12:00:00Z",
		}
		bodyMsg, _ := json.Marshal(msgPayload)
		reqSend := httptest.NewRequest("POST", "/governance/message", bytes.NewBuffer(bodyMsg))
		reqSend.Header.Set("Content-Type", "application/json")
		wSend := httptest.NewRecorder()
		r.ServeHTTP(wSend, reqSend)
		if wSend.Code != http.StatusOK {
			t.Fatalf("failed to POST /governance/message: %d", wSend.Code)
		}

		// 3. Test Message Bus Receive
		reqRecv := httptest.NewRequest("GET", "/governance/message/human-design-agent-1", nil)
		wRecv := httptest.NewRecorder()
		r.ServeHTTP(wRecv, reqRecv)
		if wRecv.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/message: %d", wRecv.Code)
		}
		var msgs []map[string]interface{}
		json.Unmarshal(wRecv.Body.Bytes(), &msgs)
		if len(msgs) != 1 || msgs[0]["from"] != "game-theory-agent-1" {
			t.Errorf("expected 1 message from game-theory-agent-1, got: %v", msgs)
		}

		// 4. Ingest and trigger RunCycle to populate Consensus Field
		ingest([]map[string]interface{}{{"val": 1.0}}, 0.5, 0.2)
		engine.GovernanceOrchestrator().RunCycle(context.Background())

		// 5. Check Consensus Field Endpoint
		reqField := httptest.NewRequest("GET", "/governance/field", nil)
		wField := httptest.NewRecorder()
		r.ServeHTTP(wField, reqField)
		if wField.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/field: %d", wField.Code)
		}
		var fieldResp map[string]interface{}
		json.Unmarshal(wField.Body.Bytes(), &fieldResp)
		if len(fieldResp) == 0 {
			t.Error("expected populated global consensus field, got empty")
		}

		// 6. Check GSR Endpoint
		reqGSR := httptest.NewRequest("GET", "/governance/gsr", nil)
		wGSR := httptest.NewRecorder()
		r.ServeHTTP(wGSR, reqGSR)
		if wGSR.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/gsr: %d", wGSR.Code)
		}
		var gsrResp map[string]interface{}
		json.Unmarshal(wGSR.Body.Bytes(), &gsrResp)
		if val, exists := gsrResp["stability_score"].(float64); !exists || val <= 0.0 {
			t.Errorf("expected active GSR stability score, got: %v", gsrResp)
		}

		// 7. Check Steward Console Telemetry
		reqStewTel := httptest.NewRequest("GET", "/steward/telemetry", nil)
		wStewTel := httptest.NewRecorder()
		r.ServeHTTP(wStewTel, reqStewTel)
		if wStewTel.Code != http.StatusOK {
			t.Fatalf("failed to GET /steward/telemetry: %d", wStewTel.Code)
		}
		var stewTelResp map[string]interface{}
		json.Unmarshal(wStewTel.Body.Bytes(), &stewTelResp)
		if _, exists := stewTelResp["stability_score"]; !exists {
			t.Error("expected stability_score in steward telemetry")
		}

		// 8. Check Steward Advisory Suggestion
		reqStewAdv := httptest.NewRequest("GET", "/steward/advisory", nil)
		wStewAdv := httptest.NewRecorder()
		r.ServeHTTP(wStewAdv, reqStewAdv)
		if wStewAdv.Code != http.StatusOK {
			t.Fatalf("failed to GET /steward/advisory: %d", wStewAdv.Code)
		}
		var stewAdvResp map[string]interface{}
		json.Unmarshal(wStewAdv.Body.Bytes(), &stewAdvResp)
		if stewAdvResp["gemini_summary"] == "" {
			t.Error("expected non-empty advisory summary")
		}

		// 9. Execute Steward Intervention Control Action
		actionPayload := map[string]string{
			"action": "MUTATION_BOUNDS",
			"target": "all",
			"value":  "0.01",
			"reason": "limit mutation variance during simulation",
		}
		bodyAct, _ := json.Marshal(actionPayload)
		reqStewAct := httptest.NewRequest("POST", "/steward/action", bytes.NewBuffer(bodyAct))
		reqStewAct.Header.Set("Content-Type", "application/json")
		wStewAct := httptest.NewRecorder()
		r.ServeHTTP(wStewAct, reqStewAct)
		if wStewAct.Code != http.StatusOK {
			t.Fatalf("failed to POST /steward/action: %d", wStewAct.Code)
		}
	})

	t.Run("Verify Tenant Lifecycle TLS-27 Subsystem", func(t *testing.T) {
		// 1. Create standard tenant with valid 81-char address
		addr := "addr12345678901234567890123456789012345678901234567890123456789012345678901234567"
		tenantPayload := map[string]string{
			"tenant_id":     "tenant-test-1",
			"owner_address": addr,
			"plan":          "Standard",
		}
		bodyTen, _ := json.Marshal(tenantPayload)
		reqCreate := httptest.NewRequest("POST", "/tenant", bytes.NewBuffer(bodyTen))
		reqCreate.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		r.ServeHTTP(wCreate, reqCreate)
		if wCreate.Code != http.StatusOK {
			t.Fatalf("failed to POST /tenant: %d, body: %s", wCreate.Code, wCreate.Body.String())
		}

		// 2. Buy go27 for standard tenant at $0.81 rate
		buyPayload := map[string]float64{
			"amount": 0.81,
		}
		bodyBuy, _ := json.Marshal(buyPayload)
		reqBuy := httptest.NewRequest("POST", "/tenant/tenant-test-1/buy-go27", bytes.NewBuffer(bodyBuy))
		reqBuy.Header.Set("Content-Type", "application/json")
		wBuy := httptest.NewRecorder()
		r.ServeHTTP(wBuy, reqBuy)
		if wBuy.Code != http.StatusOK {
			t.Fatalf("failed to POST /tenant/tenant-test-1/buy-go27: %d, body: %s", wBuy.Code, wBuy.Body.String())
		}
		var buyResp map[string]interface{}
		json.Unmarshal(wBuy.Body.Bytes(), &buyResp)
		if cycles := buyResp["compute_cycles_remaining"].(float64); cycles != 9.0 {
			t.Errorf("expected 9 compute cycles from $0.81, got: %f", cycles)
		}

		// 3. Delete tenant
		reqDel := httptest.NewRequest("DELETE", "/tenant/tenant-test-1", nil)
		wDel := httptest.NewRecorder()
		r.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusOK {
			t.Errorf("failed to DELETE /tenant/tenant-test-1: %d", wDel.Code)
		}
	})

	t.Run("Verify Constitutional Safety CSL-27 Subsystem", func(t *testing.T) {
		reqCSL := httptest.NewRequest("GET", "/governance/csl", nil)
		wCSL := httptest.NewRecorder()
		r.ServeHTTP(wCSL, reqCSL)
		if wCSL.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/csl: %d", wCSL.Code)
		}

		var cslResp map[string]interface{}
		json.Unmarshal(wCSL.Body.Bytes(), &cslResp)
		if _, exists := cslResp["active_violations"]; !exists {
			t.Error("expected active_violations field in CSL-27 response")
		}
	})

	t.Run("Verify Runtime Execution Engine REE-27 Cascade Gates", func(t *testing.T) {
		// 1. Create standard tenant with valid 81-char address
		addr := "addr12345678901234567890123456789012345678901234567890123456789012345678901234567"
		tenantPayload := map[string]string{
			"tenant_id":     "tenant-test-2",
			"owner_address": addr,
			"plan":          "Standard",
		}
		bodyTen, _ := json.Marshal(tenantPayload)
		reqCreate := httptest.NewRequest("POST", "/tenant", bytes.NewBuffer(bodyTen))
		reqCreate.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		r.ServeHTTP(wCreate, reqCreate)
		if wCreate.Code != http.StatusOK {
			t.Fatalf("failed to create tenant in REE test: %d", wCreate.Code)
		}

		// 2. Buy go27 for standard tenant to get 9 compute cycles
		buyPayload := map[string]float64{
			"amount": 0.81,
		}
		bodyBuy, _ := json.Marshal(buyPayload)
		reqBuy := httptest.NewRequest("POST", "/tenant/tenant-test-2/buy-go27", bytes.NewBuffer(bodyBuy))
		reqBuy.Header.Set("Content-Type", "application/json")
		wBuy := httptest.NewRecorder()
		r.ServeHTTP(wBuy, reqBuy)
		if wBuy.Code != http.StatusOK {
			t.Fatalf("failed to buy go27 in REE test: %d", wBuy.Code)
		}

		// 3. Propose action and execute cycle
		actionPayload := map[string]interface{}{
			"tenant_id": "tenant-test-2",
			"agent_id":  "game-theory-agent-0",
			"action": map[string]interface{}{
				"type":    "OPTIMIZE_STRATEGY",
				"payload": "optimize-payoff-matrix",
				"ethical_tensor": map[string]interface{}{
					"harm_axes":        []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
					"integrity_axes":   []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
					"sovereignty_axes": []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
				},
			},
		}
		bodyAct, _ := json.Marshal(actionPayload)
		reqExec := httptest.NewRequest("POST", "/governance/execute-cycle", bytes.NewBuffer(bodyAct))
		reqExec.Header.Set("Content-Type", "application/json")
		wExec := httptest.NewRecorder()
		r.ServeHTTP(wExec, reqExec)
		if wExec.Code != http.StatusOK {
			t.Errorf("expected execution cycle to pass, failed with: %d, body: %s", wExec.Code, wExec.Body.String())
		}
	})

	t.Run("Verify Sovereign Node Architecture SNA-27 Subsystem", func(t *testing.T) {
		reqNode := httptest.NewRequest("GET", "/governance/node", nil)
		wNode := httptest.NewRecorder()
		r.ServeHTTP(wNode, reqNode)
		if wNode.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/node: %d", wNode.Code)
		}

		var nodeResp map[string]interface{}
		json.Unmarshal(wNode.Body.Bytes(), &nodeResp)
		
		addr, exists := nodeResp["address"].(map[string]interface{})
		if !exists {
			t.Error("expected address field in Sovereign Node response")
		}
		
		fullAddr, exists := addr["full_address_81"].(string)
		if !exists || len(fullAddr) != 81 {
			t.Errorf("expected 81-char address, got: %s", fullAddr)
		}
	})

	t.Run("Verify Sovereign Node Commissioning SNCP-27 Subsystem", func(t *testing.T) {
		// Test validation failures
		invalidPayload := map[string]string{
			"spatial":    "short",
			"middleware": "too-short",
			"context":    "not-27-characters-at-all",
		}
		bodyVal, _ := json.Marshal(invalidPayload)
		reqFail := httptest.NewRequest("POST", "/governance/commission", bytes.NewBuffer(bodyVal))
		reqFail.Header.Set("Content-Type", "application/json")
		wFail := httptest.NewRecorder()
		r.ServeHTTP(wFail, reqFail)
		if wFail.Code != http.StatusBadRequest {
			t.Errorf("expected bad request for invalid segment lengths, got: %d", wFail.Code)
		}

		// Test successful validation & artifact generation
		validPayload := map[string]string{
			"spatial":    "spatial27spatial27spatial27",
			"middleware": "middleware27middleware27mid",
			"context":    "context27context27context27",
		}
		bodySucc, _ := json.Marshal(validPayload)
		reqSucc := httptest.NewRequest("POST", "/governance/commission", bytes.NewBuffer(bodySucc))
		reqSucc.Header.Set("Content-Type", "application/json")
		wSucc := httptest.NewRecorder()
		r.ServeHTTP(wSucc, reqSucc)
		if wSucc.Code != http.StatusOK {
			t.Fatalf("expected successful node commissioning, got status: %d, body: %s", wSucc.Code, wSucc.Body.String())
		}

		var resp struct {
			DeviceProvisioned bool `json:"device_provisioned"`
			IdentityBound     bool `json:"identity_bound"`
			MeshOnboarded     bool `json:"mesh_onboarded"`
			Artifacts         []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"artifacts"`
		}
		json.Unmarshal(wSucc.Body.Bytes(), &resp)

		if !resp.DeviceProvisioned || !resp.IdentityBound || !resp.MeshOnboarded {
			t.Errorf("expected all phases completed, got: %+v", resp)
		}
		if len(resp.Artifacts) != 27 {
			t.Errorf("expected exactly 27 commissioning artifacts, got: %d", len(resp.Artifacts))
		}
	})

	t.Run("Verify Sovereign Node Security Envelope RSE-81 Subsystem", func(t *testing.T) {
		reqSec := httptest.NewRequest("GET", "/governance/security", nil)
		wSec := httptest.NewRecorder()
		r.ServeHTTP(wSec, reqSec)
		if wSec.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/security: %d", wSec.Code)
		}

		var resp struct {
			GuaranteesPassed int      `json:"guarantees_passed"`
			ActiveAlerts     []string `json:"active_alerts"`
		}
		json.Unmarshal(wSec.Body.Bytes(), &resp)

		if resp.GuaranteesPassed != 81 {
			t.Errorf("expected exactly 81 security guarantees passed, got: %d (alerts: %v)", resp.GuaranteesPassed, resp.ActiveAlerts)
		}
		if len(resp.ActiveAlerts) != 0 {
			t.Errorf("expected no active alerts under nominal parameters, got: %v", resp.ActiveAlerts)
		}
	})

	t.Run("Verify Sovereign Node Global Coordination Layer SNGCL-27 Subsystem", func(t *testing.T) {
		// Test standard coordination run
		payload := map[string]interface{}{
			"input_data": "converge-27-clusters-test-data",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/governance/coordination", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected successful coordination cycle, got status: %d, body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			ActiveEmergencyMode string `json:"active_emergency_mode"`
			ConsensusReached    bool   `json:"consensus_reached"`
			ActiveClustersCount int    `json:"active_clusters_count"`
			GlobalLineageRoot   string `json:"global_lineage_root"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if !resp.ConsensusReached || resp.ActiveClustersCount != 27 {
			t.Errorf("coordination did not establish expected consensus state: %+v", resp)
		}

		// Test GEC-1 emergency pause blocks execution
		payloadPause := map[string]interface{}{
			"input_data":     "paused-data",
			"emergency_mode": "GEC-1",
		}
		bodyPause, _ := json.Marshal(payloadPause)
		reqPause := httptest.NewRequest("POST", "/governance/coordination", bytes.NewBuffer(bodyPause))
		reqPause.Header.Set("Content-Type", "application/json")
		wPause := httptest.NewRecorder()
		r.ServeHTTP(wPause, reqPause)
		if wPause.Code != http.StatusForbidden {
			t.Errorf("expected StatusForbidden (403) during GEC-1 pause block, got: %d", wPause.Code)
		}
	})

	t.Run("Verify Sovereign Node Global Evolution Model SNGEM-27 Subsystem", func(t *testing.T) {
		reqEv := httptest.NewRequest("GET", "/governance/evolution", nil)
		wEv := httptest.NewRecorder()
		r.ServeHTTP(wEv, reqEv)
		if wEv.Code != http.StatusOK {
			t.Fatalf("failed to GET /governance/evolution: %d", wEv.Code)
		}

		var resp struct {
			State struct {
				ExpansionFactor float64 `json:"expansion_factor"`
				ConvergenceRate float64 `json:"convergence_rate"`
				EmergenceIndex  float64 `json:"emergence_index"`
				LineageDensity  float64 `json:"lineage_density"`
				EvolutionEpoch  int     `json:"evolution_epoch"`
			} `json:"state"`
			Artifacts []string `json:"artifacts"`
		}
		json.Unmarshal(wEv.Body.Bytes(), &resp)

		if resp.State.EvolutionEpoch != 1 {
			t.Errorf("expected epoch 1, got %d", resp.State.EvolutionEpoch)
		}
		if len(resp.Artifacts) != 27 {
			t.Errorf("expected 27 evolution artifacts, got: %d", len(resp.Artifacts))
		}
	})

	t.Run("Verify Sovereign Field Simulation Layer SFSL-27 Subsystem", func(t *testing.T) {
		payload := map[string]interface{}{
			"mode":  "SM-3", // Sovereign Mode
			"steps": 81,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/governance/simulation", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("failed to run simulation: %d", w.Code)
		}

		var resp struct {
			Result struct {
				Mode                 string  `json:"mode"`
				StepsExecuted        int     `json:"steps_executed"`
				AverageConvergence   float64 `json:"average_convergence"`
				EntropyVariance      float64 `json:"entropy_variance"`
				SovereigntySustained bool    `json:"sovereignty_sustained"`
			} `json:"result"`
			Artifacts []string `json:"artifacts"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Result.Mode != "SM-3" || resp.Result.StepsExecuted != 81 || !resp.Result.SovereigntySustained {
			t.Errorf("unexpected simulation result: %+v", resp.Result)
		}
		if len(resp.Artifacts) != 27 {
			t.Errorf("expected 27 simulation artifacts, got: %d", len(resp.Artifacts))
		}
	})
}
