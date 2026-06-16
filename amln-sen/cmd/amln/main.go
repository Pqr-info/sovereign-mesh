package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"amln-sen/internal/api"
	"amln-sen/internal/cognition"
	"amln-sen/internal/pqr"
	"amln-sen/internal/routing"
	"amln-sen/internal/types"
)

func main() {
	// Load config
	cfg := types.LoadConfig()

	log.Printf("[AMLN-SEN] Starting node: %s", cfg.NodeID)
	log.Printf("[AMLN-SEN] Connecting to PQR backend at %s", cfg.PQREndpoint)

	// Initialize PQR session
	session := pqr.NewSession(cfg.PQREndpoint, cfg.NodeID)

	// Initialize cognition engine (SEN)
	engine, err := cognition.NewSENEngine(cfg, session)
	if err != nil {
		log.Fatalf("[AMLN-SEN] Failed to initialize cognition engine: %v", err)
	}

	// Initialize routing modules
	gossip := routing.NewGossipRouter(cfg, engine)
	slingshot := routing.NewSlingshotRouter(cfg, engine)
	consensus := routing.NewConsensusRouter(cfg, engine)

	// Initialize REST API
	router := api.NewRouter(engine, gossip, slingshot, consensus)

	// Register gossiper with engine
	engine.RegisterGossiper(gossip)

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.ContinuousEvolution {
		log.Println("[AMLN-SEN] Continuous background evolution ENABLED, starting autonomous loops...")
		engine.StartBackgroundEvolution(ctx)
		engine.StartContinuousGossip(ctx)
	}

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("[AMLN-SEN] API server error: %v", err)
		}
	}()

	log.Printf("[AMLN-SEN] REST API listening on port %s", cfg.Port)

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("[AMLN-SEN] Shutdown signal received, stopping...")

	// Allow cleanup
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	engine.Shutdown(shutdownCtx)
	gossip.Shutdown()
	slingshot.Shutdown()
	consensus.Shutdown()

	log.Println("[AMLN-SEN] Node stopped cleanly.")
}
