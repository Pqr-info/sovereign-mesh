package routing

import (
	"errors"
	"math"
)

// MeshNode represents a neighbor reference in the topology
type MeshNode struct {
	Address81 string
	Class     string // "Tier-1", "Tier-2", "Tier-3"
	Metric    float64
}

// MeshRoutingEngine optimizes routing vectors across local, regional, and global scales (MS-2)
type MeshRoutingEngine struct {
	Neighbors   map[string]MeshNode
	ClusterID   string
	GlobalTable map[string]string // Target address prefix -> next-hop cluster ID
}

func NewMeshRoutingEngine(clusterID string) *MeshRoutingEngine {
	return &MeshRoutingEngine{
		Neighbors:   make(map[string]MeshNode),
		ClusterID:   clusterID,
		GlobalTable: make(map[string]string),
	}
}

// AddNeighbor registers a local neighbor in the LM-3 boundary table
func (mre *MeshRoutingEngine) AddNeighbor(node MeshNode) error {
	if len(mre.Neighbors) >= 3 {
		return errors.New("LM-3 Limit: node can have at most 3 local neighbors")
	}
	mre.Neighbors[node.Address81] = node
	return nil
}

// ResolveNextHop determines next-hop routing target matching LM-3 -> RM-9 -> GM-27 domains
func (mre *MeshRoutingEngine) ResolveNextHop(targetAddr string) (string, error) {
	if len(targetAddr) != 81 {
		return "", errors.New("invalid target address length (must be 81 chars)")
	}

	// 1. Check local neighbors (LM-3)
	if node, exists := mre.Neighbors[targetAddr]; exists {
		return node.Address81, nil
	}

	// 2. Check prefix matching in regional cluster (RM-9 / prefix-based match)
	// Match spatial prefix segment (first 27 characters)
	targetPrefix := targetAddr[:27]
	if nextHopCluster, exists := mre.GlobalTable[targetPrefix]; exists {
		return nextHopCluster, nil
	}

	// 3. Fallback to closest local neighbor by metric weight
	var bestHop string
	bestMetric := math.MaxFloat64
	for _, node := range mre.Neighbors {
		if node.Metric < bestMetric {
			bestMetric = node.Metric
			bestHop = node.Address81
		}
	}

	if bestHop != "" {
		return bestHop, nil
	}

	return "", errors.New("no routing path found")
}
