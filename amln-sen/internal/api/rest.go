package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"amln-sen/internal/cognition"
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
