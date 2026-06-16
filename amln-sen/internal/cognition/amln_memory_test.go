package cognition

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"amln-sen/internal/pqr"
	"amln-sen/internal/types"
)

func startMockPQRMemoryServer(t *testing.T) *httptest.Server {
	r := gin.Default()
	r.POST("/REST/2.0/ticket", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ticket_id": "mem-ticket-10"})
	})
	r.POST("/REST/2.0/agent/:agent/memory/:ticket", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/REST/2.0/agent/:agent/context", func(c *gin.Context) {
		c.JSON(http.StatusOK, []map[string]interface{}{
			{
				"theta":    0.8,
				"entropy":  0.4,
				"tx_pages": []interface{}{"tx1", "tx2"},
			},
		})
	})
	return httptest.NewServer(r)
}

func TestAMLNMemoryManager(t *testing.T) {
	mockPQR := startMockPQRMemoryServer(t)
	defer mockPQR.Close()

	cfg := types.Config{
		NodeID:             "amln-memory-test-node",
		PQREndpoint:        mockPQR.URL,
		LineageVectorSize:  3,
		StrategyVectorSize: 3,
	}

	session := pqr.NewSession(cfg.PQREndpoint, cfg.NodeID)
	stmb := NewSTMB(cfg)
	stmb.SetSession(session)
	ltms := NewLTMS(session, cfg)

	amm := NewAMLNMemoryManager(stmb, ltms, 27*24*time.Hour)

	// 1. Run memory update
	ctx := context.Background()
	txPages := []map[string]interface{}{{"tx_id": "tx-1"}}
	err := amm.UpdateMemory(ctx, txPages, 0.9, 0.3)
	if err != nil {
		t.Fatalf("failed to update AMLN memory: %v", err)
	}

	// 2. Fetch combined vector
	vec, err := amm.GetMemoryVector()
	if err != nil {
		t.Fatalf("failed to fetch memory vector: %v", err)
	}

	// Short term vector (3 values) + Long term vector (3 values) = 6 values
	if len(vec) != 6 {
		t.Errorf("expected combined vector of length 6, got %d", len(vec))
	}
}
