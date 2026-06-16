package types

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	NodeID             string   `json:"node_id"`
	MeshIslandID       string   `json:"mesh_island_id"`
	GossipPeers        []string `json:"gossip_peers"`
	Port               string   `json:"port"`

	// Cognition sizes
	StrategyVectorSize int `json:"strategy_vector_size"`
	LineageVectorSize  int `json:"lineage_vector_size"`

	// RL params
	LearningRate float64 `json:"learning_rate"`
	RewardDecay  float64 `json:"reward_decay"`

	// Slingshot
	SlingshotEnabled bool   `json:"slingshot_enabled"`
	PQREndpoint      string `json:"pqr_endpoint"`

	// Autonomous Organism flags
	ContinuousEvolution bool          `json:"continuous_evolution"`
	GossipTick          time.Duration `json:"gossip_tick"`
}

func LoadConfig() Config {
	cfg := Config{}

	nodeID := os.Getenv("AMLN_NODE_ID")
	if nodeID == "" {
		nodeID = uuid.New().String()
		log.Printf("[AMLN-SEN] No AMLN_NODE_ID provided, generated: %s", nodeID)
	}
	cfg.NodeID = nodeID

	cfg.PQREndpoint = getEnvOrDefault("PQR_ENDPOINT", "http://localhost:8080")
	cfg.Port = getEnvOrDefault("AMLN_PORT", "9090")

	rawPeers := os.Getenv("GOSSIP_PEERS")
	if rawPeers != "" {
		cfg.GossipPeers = strings.Split(rawPeers, ",")
	} else {
		cfg.GossipPeers = []string{}
	}

	cfg.StrategyVectorSize = getIntOrDefault("STRATEGY_VECTOR_SIZE", 16)
	cfg.LineageVectorSize = getIntOrDefault("LINEAGE_VECTOR_SIZE", 16)
	cfg.LearningRate = getFloatOrDefault("LEARNING_RATE", 0.05)
	cfg.RewardDecay = getFloatOrDefault("REWARD_DECAY", 0.90)

	cfg.SlingshotEnabled = getBoolOrDefault("SLINGSHOT_ENABLED", true)
	cfg.MeshIslandID = getEnvOrDefault("MESH_ISLAND_ID", "default-island")

	cfg.ContinuousEvolution = getBoolOrDefault("CONTINUOUS_EVOLUTION", true)

	gossipTickMs := getIntOrDefault("GOSSIP_TICK_MS", 1000)
	cfg.GossipTick = time.Duration(gossipTickMs) * time.Millisecond

	return cfg
}

func getEnvOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getIntOrDefault(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return i
}

func getFloatOrDefault(key string, def float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return f
}

func getBoolOrDefault(key string, def bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	switch strings.ToLower(val) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return def
	}
}
