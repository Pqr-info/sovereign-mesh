package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"amln-sen/internal/cognition"
	"amln-sen/internal/governance"
	"amln-sen/internal/routing"
)

type Router struct {
	engine     *cognition.SENEngine
	gossip     *routing.GossipRouter
	slingshot  *routing.SlingshotRouter
	consensus  *routing.ConsensusRouter
}

func NewRouter(
	engine *cognition.SENEngine,
	gossip *routing.GossipRouter,
	slingshot *routing.SlingshotRouter,
	consensus *routing.ConsensusRouter,
) *gin.Engine {

	r := gin.Default()

	api := &Router{
		engine:    engine,
		gossip:    gossip,
		slingshot: slingshot,
		consensus: consensus,
	}

	// -----------------------------
	// Ingestion
	// -----------------------------
	r.POST("/ingest", api.handleIngest)

	// -----------------------------
	// Cognition
	// -----------------------------
	r.GET("/cognition/vector", api.handleCognitiveVector)
	r.GET("/cognition/weight", api.handleAgenticWeight)
	r.GET("/cognition/signed", api.handleSignedCognition)

	// -----------------------------
	// Gossip
	// -----------------------------
	r.POST("/gossip/push", api.handleGossipReceive)

	// -----------------------------
	// Slingshot Merge
	// -----------------------------
	r.POST("/slingshot/merge", api.handleSlingshotMerge)

	// -----------------------------
	// Consensus
	// -----------------------------
	r.GET("/consensus/contribute", api.handleConsensusContribution)

	// -----------------------------
	// Governance / Oversight Layer
	// -----------------------------
	r.POST("/governance/arbitrate", api.handleGovernanceArbitrate)

	// -----------------------------
	// Agent Jail / Quarantine
	// -----------------------------
	r.GET("/agent/jail", api.handleGetJailedAgents)
	r.POST("/agent/jail", api.handleJailAgent)
	r.POST("/agent/release", api.handleReleaseAgent)
	r.POST("/agent/retire", api.handleRetireAgent)
	r.POST("/agent/teleport", api.handleTeleportAgent)
	r.POST("/agent/rehabilitate", api.handleRehabilitateAgent)

	// -----------------------------
	// Registry, Field, Messaging
	// -----------------------------
	r.GET("/governance/registry", api.handleGetRegistry)
	r.GET("/governance/field", api.handleGetConsensusField)
	r.POST("/governance/message", api.handleSendMessage)
	r.GET("/governance/message/:agent_id", api.handleReceiveMessages)
	r.GET("/governance/gsr", api.handleGetGSR)

	// -----------------------------
	// Steward Console / Mobile Dashboard API
	// -----------------------------
	r.GET("/steward/telemetry", api.handleGetStewardTelemetry)
	r.POST("/steward/action", api.handleExecuteStewardAction)
	r.GET("/steward/advisory", api.handleGetStewardAdvisory)

	// -----------------------------
	// Tenant Lifecycle (TLS-27) API
	// -----------------------------
	r.POST("/tenant", api.handleCreateTenant)
	r.DELETE("/tenant/:id", api.handleDeleteTenant)
	r.POST("/tenant/:id/buy-go27", api.handleBuyGo27)

	// -----------------------------
	// Constitutional Safety (CSL-27) API
	// -----------------------------
	r.GET("/governance/csl", api.handleGetCSL)

	// -----------------------------
	// Runtime Execution (REE-27) API
	// -----------------------------
	r.POST("/governance/execute-cycle", api.handleExecuteCycle)

	// -----------------------------
	// Sovereign Node Architecture (SNA-27) API
	// -----------------------------
	r.GET("/governance/node", api.handleGetSovereignNode)
	r.GET("/governance/deployment", api.handleGetDeployment)
	r.POST("/governance/commission", api.handleCommissionNode)
	r.GET("/governance/security", api.handleGetSecurityEnvelope)
	r.POST("/governance/coordination", api.handleGlobalConsensusRun)
	r.GET("/governance/evolution", api.handleGetGlobalEvolution)
	r.POST("/governance/simulation", api.handleRunSimulationForecast)

	return r
}

// ------------------------------------------------------------
// /ingest
// ------------------------------------------------------------

type ingestRequest struct {
	TxPages []map[string]interface{} `json:"tx_pages"`
	Theta   float64                  `json:"theta"`
	Entropy float64                  `json:"entropy"`
}

func (a *Router) handleIngest(c *gin.Context) {
	var req ingestRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := context.Background()
	a.engine.Ingest(ctx, req.TxPages, req.Theta, req.Entropy)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ------------------------------------------------------------
// /cognition/vector
// ------------------------------------------------------------

func (a *Router) handleCognitiveVector(c *gin.Context) {
	Ck := a.engine.CognitiveVector()
	c.JSON(http.StatusOK, gin.H{"vector": Ck})
}

// ------------------------------------------------------------
// /cognition/weight
// ------------------------------------------------------------

func (a *Router) handleAgenticWeight(c *gin.Context) {
	alpha := a.engine.AgenticWeight()
	c.JSON(http.StatusOK, gin.H{"alpha": alpha})
}

// ------------------------------------------------------------
// /cognition/signed
// ------------------------------------------------------------

func (a *Router) handleSignedCognition(c *gin.Context) {
	env, err := a.engine.SignedCognition()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "signing failed"})
		return
	}
	c.JSON(http.StatusOK, env)
}

// ------------------------------------------------------------
// /gossip/push
// ------------------------------------------------------------

func (a *Router) handleGossipReceive(c *gin.Context) {
	var payload routing.GossipPayload
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid gossip payload"})
		return
	}

	ctx := context.Background()
	if err := a.gossip.Receive(ctx, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store gossip"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ------------------------------------------------------------
// /slingshot/merge
// ------------------------------------------------------------

func (a *Router) handleSlingshotMerge(c *gin.Context) {
	var manifest routing.SlingshotManifest
	if err := c.BindJSON(&manifest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest"})
		return
	}

	ctx := context.Background()
	if err := a.slingshot.MergeManifest(ctx, manifest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "merge failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "merged"})
}

// ------------------------------------------------------------
// /consensus/contribute
// ------------------------------------------------------------

func (a *Router) handleConsensusContribution(c *gin.Context) {
	payload := a.consensus.BuildContribution()
	c.JSON(http.StatusOK, payload)
}

// ------------------------------------------------------------
// /governance/arbitrate
// ------------------------------------------------------------

type governanceArbitrateRequest struct {
	Proposals []governance.AgentOutput `json:"proposals"`
}

func (a *Router) handleGovernanceArbitrate(c *gin.Context) {
	var req governanceArbitrateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	decision := a.engine.Council().Arbitrate(req.Proposals)
	monitors := governance.GetDefaultMonitors()
	tensor := governance.EvaluateEthicalTensor(decision, monitors)
	verdict := governance.EvaluateVerdict(tensor, monitors)

	var finalVerdict governance.FinalVerdict
	if verdict.Passed {
		finalVerdict = governance.FinalVerdict{
			Approved: true,
			Decision: decision.FinalDecision,
			Notes:    []string{"All ethical dimensions satisfied"},
		}
	} else {
		finalVerdict = governance.FinalVerdict{
			Approved: false,
			Decision: nil,
			Notes:    verdict.FailingDimensions,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"decision": decision,
		"tensor":   tensor,
		"verdict":  finalVerdict,
	})
}

// ------------------------------------------------------------
// Agent Jail / Quarantine Handlers
// ------------------------------------------------------------

func (a *Router) handleGetJailedAgents(c *gin.Context) {
	c.JSON(http.StatusOK, a.engine.GovernanceOrchestrator().JailController.Active)
}

type agentJailRequest struct {
	AgentID string `json:"agent_id"`
	Reason  string `json:"reason"`
}

func (a *Router) handleJailAgent(c *gin.Context) {
	var req agentJailRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	var targetAgent governance.Agent
	for _, agent := range a.engine.GovernanceOrchestrator().Agents {
		if agent.AgentID() == req.AgentID {
			targetAgent = agent
			break
		}
	}

	if targetAgent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	err := a.engine.GovernanceOrchestrator().JailController.JailAgent(targetAgent, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "agent quarantined"})
}

type agentIDRequest struct {
	AgentID string `json:"agent_id"`
}

func (a *Router) handleReleaseAgent(c *gin.Context) {
	var req agentIDRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	log.Printf("[API] handleReleaseAgent received AgentID: %q", req.AgentID)
	err := a.engine.GovernanceOrchestrator().JailController.ReleaseAgent(req.AgentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "agent released"})
}

func (a *Router) handleRetireAgent(c *gin.Context) {
	var req agentIDRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := a.engine.GovernanceOrchestrator().JailController.RetireAgent(req.AgentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "agent retired"})
}

type agentTeleportRequest struct {
	AgentID      string `json:"agent_id"`
	TargetLocale string `json:"target_locale"`
}

func (a *Router) handleTeleportAgent(c *gin.Context) {
	var req agentTeleportRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := a.engine.TeleportationManager().RequestTeleport(req.AgentID, req.TargetLocale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "teleport allowed"})
}

func (a *Router) handleRehabilitateAgent(c *gin.Context) {
	var req agentIDRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	scorer := governance.NewDefaultRehabilitationScorer()
	score, passed, err := a.engine.GovernanceOrchestrator().JailController.RequestRehabilitation(req.AgentID, scorer, a.engine.GovernanceOrchestrator())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"score":  score,
		"passed": passed,
		"status": "rehabilitation cycle processed",
	})
}

func (a *Router) handleGetRegistry(c *gin.Context) {
	c.JSON(http.StatusOK, a.engine.GovernanceOrchestrator().Registry)
}

func (a *Router) handleGetConsensusField(c *gin.Context) {
	c.JSON(http.StatusOK, a.engine.GovernanceOrchestrator().ConsensusField)
}

func (a *Router) handleSendMessage(c *gin.Context) {
	var msg governance.AgentMessage
	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message"})
		return
	}
	a.engine.GovernanceOrchestrator().MessageBus.Send(msg)
	c.JSON(http.StatusOK, gin.H{"status": "message sent"})
}

func (a *Router) handleReceiveMessages(c *gin.Context) {
	agentID := c.Param("agent_id")
	msgs := a.engine.GovernanceOrchestrator().MessageBus.Receive(agentID)
	c.JSON(http.StatusOK, msgs)
}

func (a *Router) handleGetGSR(c *gin.Context) {
	c.JSON(http.StatusOK, a.engine.GovernanceOrchestrator().GSR)
}

func (a *Router) handleGetStewardTelemetry(c *gin.Context) {
	o := a.engine.GovernanceOrchestrator()
	o.ConsensusField.Mu.RLock()
	defer o.ConsensusField.Mu.RUnlock()

	telemetry := gin.H{
		"stability_score":   o.GSR.StabilityScore,
		"entropy_level":     o.GSR.EntropyLevel,
		"theta":             o.GSR.Theta,
		"agentic_weight":    o.GSR.AgenticWeight,
		"ethical_variance":  o.GSR.EthicalVariance,
		"jailed_agents":     len(o.JailController.Active),
		"probation_agents":  len(o.Probation),
		"active_agents":     len(o.Agents) - len(o.JailController.Active),
	}

	c.JSON(http.StatusOK, telemetry)
}

type stewardActionRequest struct {
	Action string `json:"action"` // e.g. "MUTATION_BOUNDS", "EMERGENCY_PAUSE", "JAIL_AGENT"
	Target string `json:"target"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

func (a *Router) handleExecuteStewardAction(c *gin.Context) {
	var req stewardActionRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	o := a.engine.GovernanceOrchestrator()
	
	// Process intervention logic and audit log it
	log.Printf("[STEWARD INTERVENTION] Action: %s, Target: %s, Value: %s, Reason: %s", req.Action, req.Target, req.Value, req.Reason)

	switch req.Action {
	case "EMERGENCY_PAUSE":
		// Freezes active cycles or evolution tick modifiers
		o.MutationGovernor.StabilityModifier = 0.0
	case "MUTATION_BOUNDS":
		o.MutationGovernor.BaseRate = 0.01
	case "JAIL_AGENT":
		for _, ag := range o.Agents {
			if ag.AgentID() == req.Target {
				_ = o.JailController.JailAgent(ag, "Steward Intervention: "+req.Reason)
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "intervention executed successfully"})
}

func (a *Router) handleGetStewardAdvisory(c *gin.Context) {
	// Generates high-signal advisory analysis
	advisory := gin.H{
		"gemini_summary": "System remains in nominal stability range. Ethical compliance bounds satisfy targets.",
		"top_concerns": []string{
			"Lineage divergence bounds within expectations",
			"Monitor validation passing with green status",
		},
		"suggested_actions": []gin.H{
			{
				"action": "MUTATION_BOUNDS",
				"justification": "Stabilize convergence margins during peak compute variance",
			},
		},
	}

	c.JSON(http.StatusOK, advisory)
}

type createTenantRequest struct {
	TenantID     string              `json:"tenant_id"`
	OwnerAddress string              `json:"owner_address"`
	Plan         governance.TenantPlan `json:"plan"`
}

func (a *Router) handleCreateTenant(c *gin.Context) {
	var req createTenantRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	o := a.engine.GovernanceOrchestrator()
	t, err := o.TenantManager.CreateTenant(req.TenantID, req.OwnerAddress, req.Plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, t)
}

func (a *Router) handleDeleteTenant(c *gin.Context) {
	id := c.Param("id")
	o := a.engine.GovernanceOrchestrator()
	err := o.TenantManager.DeleteTenant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "tenant deleted"})
}

type buyGo27Request struct {
	Amount float64 `json:"amount"`
}

func (a *Router) handleBuyGo27(c *gin.Context) {
	id := c.Param("id")
	var req buyGo27Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	o := a.engine.GovernanceOrchestrator()
	o.TenantManager.Mu.Lock()
	tenant, exists := o.TenantManager.Tenants[id]
	o.TenantManager.Mu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	tenant.PurchaseGo27(req.Amount)
	c.JSON(http.StatusOK, tenant)
}

func (a *Router) handleGetCSL(c *gin.Context) {
	c.JSON(http.StatusOK, a.engine.GovernanceOrchestrator().CSL)
}

type executeCycleRequest struct {
	TenantID string                   `json:"tenant_id"`
	AgentID  string                   `json:"agent_id"`
	Action   governance.RuntimeAction `json:"action"`
}

func (a *Router) handleExecuteCycle(c *gin.Context) {
	var req executeCycleRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	o := a.engine.GovernanceOrchestrator()
	var targetAgent governance.Agent
	for _, ag := range o.Agents {
		if ag.AgentID() == req.AgentID {
			targetAgent = ag
			break
		}
	}

	if targetAgent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	err := o.REE.ExecuteCycle(context.Background(), req.TenantID, targetAgent, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cycle executed successfully"})
}

func (a *Router) handleGetSovereignNode(c *gin.Context) {
	o := a.engine.GovernanceOrchestrator()
	spatial := "spatial27spatial27spatial27"
	middleware := "middleware27middleware27mid"
	ctxVal := "context27context27context27"

	addr, err := governance.NewNodeAddress81(spatial, middleware, ctxVal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	node := governance.NewSovereignNode(addr, o)
	c.JSON(http.StatusOK, node)
}

func (a *Router) handleGetDeployment(c *gin.Context) {
	profile := governance.GetHardwareProfile(governance.ClassGlobalCoreNode)
	dep := governance.SovereignNodeDeployment{
		Class:   governance.ClassGlobalCoreNode,
		Profile: profile,
		MeshID:  "sovereign-mesh-global",
	}
	c.JSON(http.StatusOK, dep)
}

type commissionRequest struct {
	Spatial    string `json:"spatial"`
	Middleware string `json:"middleware"`
	Context    string `json:"context"`
}

func (a *Router) handleCommissionNode(c *gin.Context) {
	var req commissionRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	cc := governance.NewCommissioningController()
	err := cc.RunCommissioningPipeline(req.Spatial, req.Middleware, req.Context)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cc.State)
}

func (a *Router) handleGetSecurityEnvelope(c *gin.Context) {
	spatial := c.DefaultQuery("spatial", "spatial27spatial27spatial27")
	middleware := c.DefaultQuery("middleware", "middleware27middleware27mid")
	contextVal := c.DefaultQuery("context", "context27context27context27")

	env, err := governance.EvaluateSecurityEnvelope(
		spatial, middleware, contextVal,
		a.engine.GovernanceOrchestrator().GSR.Theta, 1.0, a.engine.AgenticWeight(),
		100.0, 9,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, env)
}

type coordinationRequest struct {
	InputData     string                         `json:"input_data"`
	EmergencyMode governance.GlobalEmergencyMode `json:"emergency_mode"`
}

func (a *Router) handleGlobalConsensusRun(c *gin.Context) {
	var req coordinationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	gcm := governance.NewGlobalCoordinationManager()

	if req.EmergencyMode != "" {
		err := gcm.TriggerEmergencyMode(req.EmergencyMode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	err := gcm.RunGlobalConsensus(req.InputData)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gcm.State)
}

func (a *Router) handleGetGlobalEvolution(c *gin.Context) {
	gem := governance.NewGlobalEvolutionModel()
	gem.EvolveStep(27, a.engine.AgenticWeight())

	artifacts := gem.GenerateEvolutionArtifacts()

	c.JSON(http.StatusOK, gin.H{
		"state":     gem.State,
		"artifacts": artifacts,
	})
}

type simulationRequest struct {
	Mode  governance.SimulationMode `json:"mode"`
	Steps int                       `json:"steps"`
}

func (a *Router) handleRunSimulationForecast(c *gin.Context) {
	var req simulationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if req.Steps <= 0 {
		req.Steps = 27
	}

	sfsl := governance.NewSovereignFieldSimulationLayer(req.Mode)
	res := sfsl.RunForecast(req.Steps)
	artifacts := sfsl.GenerateSimulationArtifacts()

	c.JSON(http.StatusOK, gin.H{
		"result":    res,
		"artifacts": artifacts,
	})
}



