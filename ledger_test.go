package sovereign

import (
	"testing"
)

// TestExecuteQuantumFlattening_Veto simulates a perfectly divergent matrix (Chaos)
// where an Orbital Agent uses its high TrustScore to issue a Minority Report Veto.
func TestExecuteQuantumFlattening_Veto(t *testing.T) {
	// Initialize controller with mock shared memory bus
	c := &Controller{
		agents:    make(map[string]*Agent),
		memoryBus: make([]byte, BusSize),
	}

	// 1. Setup a Satellite Agent with high stake (L4 primitive)
	agentID := "sat-01-orbital"
	c.agents[agentID] = &Agent{
		ID:           agentID,
		SatelliteID:  "STARLINK-SHELL-1",
		StakedAmount: 50000, // Qualifies for Veto authority (> 10k)
		MemoryOffset: 0,
	}

	// 2. Set high conviction in the Shared Memory Bus (L7 conviction variable)
	// This simulates the agent "ray-tracing" a preferred path through the chaos.
	state := c.GetAgentState(0)
	state.Conviction = 0.9 // High conviction (> 0.8)

	// 3. Define a perfectly divergent environment
	var matrix ConvergenceMatrix // Empty/Divergent matrix
	entropy := EntropyBalance{
		StabilityIndex: 0.05, // Critical Chaos: Below 0.1 threshold
	}

	// 4. Execute Quantum Flattening
	_, _, resolved := c.ExecuteQuantumFlattening(matrix, entropy)

	// 5. Assert that the Veto was triggered to prevent a system stall
	if !resolved {
		t.Errorf("Quantum Stall detected! Veto should have been triggered by Satellite %s but was not.", agentID)
	}

	// 6. Verify Stall behavior if conviction is low
	state.Conviction = 0.1
	_, _, resolvedLowConviction := c.ExecuteQuantumFlattening(matrix, entropy)
	if resolvedLowConviction {
		t.Error("System should have stalled (returned false) due to low orbital conviction during chaos.")
	}
}