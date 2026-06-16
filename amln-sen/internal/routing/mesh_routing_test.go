package routing

import (
	"testing"
)

func TestMeshRoutingEngine(t *testing.T) {
	mre := NewMeshRoutingEngine("cluster-99")

	addr1 := "addr12345678901234567890123456789012345678901234567890123456789012345678901234567"
	addr2 := "addr23456789012345678901234567890123456789012345678901234567890123456789012345678"
	addr3 := "addr34567890123456789012345678901234567890123456789012345678901234567890123456789"
	addr4 := "addr45678901234567890123456789012345678901234567890123456789012345678901234567890"

	// 1. Add neighbors up to LM-3 limit
	_ = mre.AddNeighbor(MeshNode{Address81: addr1, Class: "Tier-1", Metric: 10.0})
	_ = mre.AddNeighbor(MeshNode{Address81: addr2, Class: "Tier-2", Metric: 2.0}) // Lower metric is preferred
	_ = mre.AddNeighbor(MeshNode{Address81: addr3, Class: "Tier-3", Metric: 15.0})

	// 2. Assert LM-3 limit triggers error on fourth neighbor
	errLimit := mre.AddNeighbor(MeshNode{Address81: addr4, Class: "Tier-1", Metric: 5.0})
	if errLimit == nil {
		t.Error("expected error adding fourth neighbor under LM-3 boundaries, got nil")
	}

	// 3. Resolve direct neighbor route
	hopDirect, err := mre.ResolveNextHop(addr1)
	if err != nil {
		t.Fatalf("failed to resolve next hop: %v", err)
	}
	if hopDirect != addr1 {
		t.Errorf("expected direct next-hop %s, got %s", addr1, hopDirect)
	}

	// 4. Resolve indirect target using GlobalTable prefix mapping
	// Each segment must be exactly 27 characters.
	spatial := "spatial27spatial27spatial27"
	middleware := "middleware27middleware27mid"
	contextVal := "context27context27context27"
	remoteAddr := spatial + middleware + contextVal // 81 characters total

	mre.GlobalTable[spatial] = addr3

	hopRemote, err := mre.ResolveNextHop(remoteAddr)
	if err != nil {
		t.Fatalf("failed to resolve remote address next hop: %v", err)
	}
	if hopRemote != addr3 {
		t.Errorf("expected route through prefix map target %s, got %s", addr3, hopRemote)
	}

	// 5. Fallback check (least metric cost path chosen)
	fallbackAddr := "unmapped27unmapped27unmappd" + middleware + contextVal
	hopFallback, err := mre.ResolveNextHop(fallbackAddr)
	if err != nil {
		t.Fatalf("failed fallback resolution: %v", err)
	}
	// addr2 has the lowest metric (2.0)
	if hopFallback != addr2 {
		t.Errorf("expected fallback path to prefer lower metric neighbor %s, got %s", addr2, hopFallback)
	}
}
