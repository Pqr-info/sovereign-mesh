package routing

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type SRRPNeighbor struct {
	ID        string
	Addr      string
	Transport string
	Metric    float64
	LastSeen  time.Time
}

type MeshRoutingEngine struct {
	mu          sync.RWMutex
	neighbors   map[string]SRRPNeighbor
	ClusterID   string
	GlobalTable map[string]string // Target address prefix -> next-hop cluster ID
}

func NewMeshRoutingEngine(clusterID string) *MeshRoutingEngine {
	return &MeshRoutingEngine{
		neighbors:   make(map[string]SRRPNeighbor),
		ClusterID:   clusterID,
		GlobalTable: make(map[string]string),
	}
}

func (m *MeshRoutingEngine) Start(ctx context.Context) error {
	return nil
}

func (m *MeshRoutingEngine) AddOrUpdateNeighbor(n SRRPNeighbor) {
	m.mu.Lock()
	m.neighbors[n.ID] = n
	m.mu.Unlock()

	m.recomputeRouting()
}

func (m *MeshRoutingEngine) RemoveNeighbor(id string) {
	m.mu.Lock()
	delete(m.neighbors, id)
	m.mu.Unlock()

	m.recomputeRouting()
}

func (m *MeshRoutingEngine) LocalNeighborCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.neighbors)
}

func (m *MeshRoutingEngine) recomputeRouting() {
	m.mu.RLock()
	neighborsCopy := make([]SRRPNeighbor, 0, len(m.neighbors))
	for _, n := range m.neighbors {
		neighborsCopy = append(neighborsCopy, n)
	}
	m.mu.RUnlock()

	log.Printf("[SRRP] Recomputing routing table with %d neighbors", len(neighborsCopy))
}

func (m *MeshRoutingEngine) ResolveNextHop(targetAddr string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(targetAddr) != 81 {
		return "", errors.New("invalid target address length (must be 81 chars)")
	}

	// 1. Check local neighbors
	for _, node := range m.neighbors {
		if node.ID == targetAddr {
			return node.Addr, nil
		}
	}

	// 2. Check prefix mapping
	targetPrefix := targetAddr[:27]
	if nextHopCluster, exists := m.GlobalTable[targetPrefix]; exists {
		return nextHopCluster, nil
	}

	// 3. Fallback to closest neighbor by metric weight
	var bestHop string
	bestMetric := 999999.0
	for _, node := range m.neighbors {
		if node.Metric < bestMetric {
			bestMetric = node.Metric
			bestHop = node.Addr
		}
	}

	if bestHop != "" {
		return bestHop, nil
	}

	return "", errors.New("no routing path found")
}
