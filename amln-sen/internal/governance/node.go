package governance

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ------------------------------------------------------------
// 1. Identity & Addressing Layer (IAL-81)
// ------------------------------------------------------------

type NodeAddress81 struct {
	Spatial27     string `json:"spatial_27"`
	Middleware27  string `json:"middleware_27"`
	Context27     string `json:"context_27"`
	FullAddress81 string `json:"full_address_81"`
}

func NewNodeAddress81(spatial, middleware, context string) (*NodeAddress81, error) {
	if len(spatial) != 27 || len(middleware) != 27 || len(context) != 27 {
		return nil, errors.New("each address segment must be exactly 27 characters")
	}
	full := spatial + middleware + context
	return &NodeAddress81{
		Spatial27:     spatial,
		Middleware27:  middleware,
		Context27:     context,
		FullAddress81: full,
	}, nil
}

// ------------------------------------------------------------
// 2. Consensus & Lineage Layer (CLL-273)
// ------------------------------------------------------------

type ConformationCertificate struct {
	BasisOfBehavior  float64   `json:"basis_of_behavior"`
	TensionOfBehavior float64   `json:"tension_of_behavior"`
	AngleOfBehavior   float64   `json:"angle_of_behavior"`
	ConformationAngle float64   `json:"conformation_angle"`
	Timestamp         time.Time `json:"timestamp"`
}

// ------------------------------------------------------------
// 3. Sovereign Node Architecture (SNA-27)
// ------------------------------------------------------------

type SovereignNode struct {
	Address        *NodeAddress81          `json:"address"`
	Orchestrator   *GovernanceOrchestrator `json:"-"`
	ExecutionChain []string                `json:"execution_chain"`
}

func NewSovereignNode(addr *NodeAddress81, o *GovernanceOrchestrator) *SovereignNode {
	return &SovereignNode{
		Address:        addr,
		Orchestrator:   o,
		ExecutionChain: []string{},
	}
}

// ExecuteSovereignCycle runs a Triple-Helix validated runtime step
func (n *SovereignNode) ExecuteSovereignCycle(
	ctx context.Context,
	tenantID string,
	agent Agent,
	action RuntimeAction,
	cert ConformationCertificate,
) error {
	// 1. Validate via Triple-Helix Consensus
	if cert.BasisOfBehavior == 0.0 || cert.TensionOfBehavior == 0.0 || cert.AngleOfBehavior == 0.0 {
		return errors.New("SNA-27 Veto: Triple-Helix validation failed")
	}

	// 2. Execute via REE-27 Cascade Gates
	err := n.Orchestrator.REE.ExecuteCycle(ctx, tenantID, agent, action)
	if err != nil {
		return fmt.Errorf("SNA-27 Execution Veto: %w", err)
	}

	n.ExecutionChain = append(n.ExecutionChain, action.Type)
	return nil
}
