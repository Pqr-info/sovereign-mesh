package routing

import (
	"testing"
	"time"
)

func TestSRRPRouting(t *testing.T) {
	mre := NewMeshRoutingEngine("test-cluster")

	// 1. Verify adding neighbors
	n1 := SRRPNeighbor{
		ID:        "peer1",
		Addr:      "udp4://127.0.0.1:9999",
		Transport: "LAN_MULTICAST",
		Metric:    1.0,
		LastSeen:  time.Now(),
	}
	n2 := SRRPNeighbor{
		ID:        "peer2",
		Addr:      "udp4://127.0.0.1:9998",
		Transport: "LAN_MULTICAST",
		Metric:    2.0,
		LastSeen:  time.Now(),
	}

	mre.AddOrUpdateNeighbor(n1)
	mre.AddOrUpdateNeighbor(n2)

	if mre.LocalNeighborCount() != 2 {
		t.Errorf("expected 2 neighbors, got %d", mre.LocalNeighborCount())
	}

	// 2. Resolve next hop prefers best metric (n1 with 1.0)
	targetAddr := "peer2-extra-dummy-padding-to-81-characters-to-satisfy-the-address-length-limit-81" // exactly 81 chars
	nextHop, err := mre.ResolveNextHop(targetAddr)
	if err != nil {
		t.Fatalf("failed to resolve next hop: %v", err)
	}
	if nextHop != n1.Addr {
		t.Errorf("expected best metric target %s, got %s", n1.Addr, nextHop)
	}

	// 3. Verify RemoveNeighbor
	mre.RemoveNeighbor("peer1")
	if mre.LocalNeighborCount() != 1 {
		t.Errorf("expected 1 neighbor after removal, got %d", mre.LocalNeighborCount())
	}
}
