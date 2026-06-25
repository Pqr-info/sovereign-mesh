package governance

import (
	"errors"
	"fmt"
)

// CommissioningPhase specifies the SNCP-27 phases
type CommissioningPhase string

const (
	PhaseDeviceProvisioning CommissioningPhase = "DP-9"
	PhaseIdentityBinding    CommissioningPhase = "IB-9"
	PhaseMeshOnboarding     CommissioningPhase = "MO-9"
)

// SNCPArtifact represents one of the 27 commissioning artifacts produced during the protocol.
type SNCPArtifact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CommissioningState keeps track of the current status of SNCP-27
type CommissioningState struct {
	DeviceProvisioned bool           `json:"device_provisioned"`
	IdentityBound     bool           `json:"identity_bound"`
	MeshOnboarded     bool           `json:"mesh_onboarded"`
	Artifacts         []SNCPArtifact `json:"artifacts"`
}

// CommissioningController handles the SNCP-27 node activation pipeline
type CommissioningController struct {
	State CommissioningState `json:"state"`
}

func NewCommissioningController() *CommissioningController {
	return &CommissioningController{
		State: CommissioningState{
			Artifacts: make([]SNCPArtifact, 0),
		},
	}
}

// RunCommissioningPipeline executes the 3 phases, producing 27 artifacts
func (cc *CommissioningController) RunCommissioningPipeline(spatial, middleware, contextVal string) error {
	// Step 1: Device Provisioning
	dpArtifacts := []string{
		"Secure Element Root Key",
		"Commissioning Entropy Seed",
		"128-bit Register Init (TRICLONE)",
		"Agent State Partition Map",
		"Network Interface Setup (BLE+Wi-Fi+LoRa)",
		"Mesh Identity Pre-Check Seal",
		"NBEP Firmware Integrity Hash",
		"Constitutional Baseline (BL-3/GL-9/ET-27)",
		"Commissioning Mode Hardware Lock",
	}
	for _, art := range dpArtifacts {
		cc.State.Artifacts = append(cc.State.Artifacts, SNCPArtifact{
			Name:  art,
			Value: fmt.Sprintf("DP-Artifact-%d", len(cc.State.Artifacts)+1),
		})
	}
	cc.State.DeviceProvisioned = true

	// Step 2: Identity Binding
	if len(spatial) != 27 || len(middleware) != 27 || len(contextVal) != 27 {
		return errors.New("identity binding failed: segments must be exactly 27 chars each")
	}

	addr, err := NewNodeAddress81(spatial, middleware, contextVal)
	if err != nil {
		return fmt.Errorf("identity binding failed: %w", err)
	}

	ibArtifacts := []string{
		"81-char address: " + addr.FullAddress81,
		"27-char trajectory hash",
		"5-D vertex map (X/Y/Z/V/I)",
		"Secure Element Identity Signature",
		"Arbitration Signing Keypair (AAE)",
		"Lineage Seed Vector",
		"Conformation Baseline (theta=0, epsilon=1, a=0.5)",
		"AMLN Seeding Vector Map",
		"Identity Lock Seal",
	}
	for _, art := range ibArtifacts {
		cc.State.Artifacts = append(cc.State.Artifacts, SNCPArtifact{
			Name:  art,
			Value: fmt.Sprintf("IB-Artifact-%d", len(cc.State.Artifacts)+1),
		})
	}
	cc.State.IdentityBound = true

	// Step 3: Mesh Onboarding
	moArtifacts := []string{
		"Triple-Helix Consensus Warm-Start (BoB/ToB/AoB)",
		"PQR-273 Initialization Block",
		"Slingshot Epoch Token & Buffer",
		"go27 Economic Init (Bal=0, Rate=0.81, Cycles=0)",
		"CSL-27 Safety Engine Activation",
		"REE-27 Runtime Agent State Scheduler",
		"LM-3 Neighbor Table",
		"RM-9 Cluster Membership Table",
		"GM-27 Registry Registration Seal",
	}
	for _, art := range moArtifacts {
		cc.State.Artifacts = append(cc.State.Artifacts, SNCPArtifact{
			Name:  art,
			Value: fmt.Sprintf("MO-Artifact-%d", len(cc.State.Artifacts)+1),
		})
	}
	cc.State.MeshOnboarded = true

	return nil
}
