package governance

import (
	"fmt"
	"math/rand"
)

// SimulationMode specifies the active simulation behavior (SM-3)
type SimulationMode string

const (
	ModeDeterministic SimulationMode = "SM-1"
	ModeStochastic    SimulationMode = "SM-2"
	ModeSovereign     SimulationMode = "SM-3"
)

// SFSLSimulationResult captures the metrics produced by the foresight engine (SFSL-27)
type SFSLSimulationResult struct {
	Mode               SimulationMode `json:"mode"`
	StepsExecuted      int            `json:"steps_executed"`
	AverageConvergence float64        `json:"average_convergence"`
	EntropyVariance    float64        `json:"entropy_variance"`
	SovereigntySustained bool         `json:"sovereignty_sustained"`
}

// SovereignFieldSimulationLayer handles multi-mode foresight forecasts
type SovereignFieldSimulationLayer struct {
	Mode SimulationMode `json:"mode"`
}

func NewSovereignFieldSimulationLayer(mode SimulationMode) *SovereignFieldSimulationLayer {
	return &SovereignFieldSimulationLayer{Mode: mode}
}

// RunForecast runs a simulated foresight projection over the specified runtime steps
func (s *SovereignFieldSimulationLayer) RunForecast(steps int) SFSLSimulationResult {
	avgConv := 0.95
	entVar := 0.02
	sovSustained := true

	if s.Mode == ModeStochastic {
		// Stochastic variance simulation (SM-2)
		r := rand.New(rand.NewSource(int64(steps)))
		avgConv = 0.5 + r.Float64()*0.45
		entVar = r.Float64() * 0.9
		if entVar > 0.7 {
			sovSustained = false
		}
	} else if s.Mode == ModeSovereign {
		// Sovereign mode starts at absolute equilibrium (SM-3)
		avgConv = 1.0
		entVar = 0.0
		sovSustained = true
	}

	return SFSLSimulationResult{
		Mode:                 s.Mode,
		StepsExecuted:        steps,
		AverageConvergence:   avgConv,
		EntropyVariance:      entVar,
		SovereigntySustained: sovSustained,
	}
}

// GenerateSimulationArtifacts outputs the canonical 27 visualization maps
func (s *SovereignFieldSimulationLayer) GenerateSimulationArtifacts() []string {
	artifacts := make([]string, 27)
	for i := 0; i < 9; i++ {
		artifacts[i] = fmt.Sprintf("LA-%d: Heatmap/Drift Lineage Curve", i+1)
		artifacts[i+9] = fmt.Sprintf("CA-%d: Conformation Field Sync Map", i+1)
		artifacts[i+18] = fmt.Sprintf("AA-%d: Agentic Weight / Mutation Log", i+1)
	}
	return artifacts
}
