package governance

import (
	"fmt"
	"math"
)

// GlobalEvolutionState tracks the metrics of SNGEM-27
type GlobalEvolutionState struct {
	ExpansionFactor float64 `json:"expansion_factor"`
	ConvergenceRate float64 `json:"convergence_rate"`
	EmergenceIndex  float64 `json:"emergence_index"`
	LineageDensity  float64 `json:"lineage_density"`
	EvolutionEpoch  int     `json:"evolution_epoch"`
}

// GlobalEvolutionModel implements SNGEM-27 dynamics
type GlobalEvolutionModel struct {
	State GlobalEvolutionState `json:"state"`
}

func NewGlobalEvolutionModel() *GlobalEvolutionModel {
	return &GlobalEvolutionModel{
		State: GlobalEvolutionState{
			ExpansionFactor: 1.0,
			ConvergenceRate: 0.5,
			EmergenceIndex:  0.1,
			LineageDensity:  1.0,
			EvolutionEpoch:  0,
		},
	}
}

// EvolveStep applies triadic evolution dynamics: Expansion (GED-1) -> Convergence (GED-2) -> Emergence (GED-3)
func (gem *GlobalEvolutionModel) EvolveStep(nodeCount int, avgWeight float64) {
	gem.State.EvolutionEpoch++
	
	// GED-1: Expansion scales with nodeCount
	gem.State.ExpansionFactor = math.Log1p(float64(nodeCount))

	// GED-2: Convergence scales with line density & average agentic weight
	gem.State.LineageDensity = float64(nodeCount) * 1.5
	gem.State.ConvergenceRate = (1.0 - math.Exp(-0.1*gem.State.LineageDensity)) * avgWeight

	// GED-3: Emergence represents the complexity offset
	gem.State.EmergenceIndex = gem.State.ExpansionFactor * gem.State.ConvergenceRate
}

// GenerateEvolutionArtifacts produces 27 artifacts mapping global evolutionary progress
func (gem *GlobalEvolutionModel) GenerateEvolutionArtifacts() []string {
	artifacts := make([]string, 27)
	for i := 0; i < 9; i++ {
		artifacts[i] = fmt.Sprintf("LEA-%d: Lineage Density / Epoch %d / Exp %.2f", i+1, gem.State.EvolutionEpoch, gem.State.ExpansionFactor)
		artifacts[i+9] = fmt.Sprintf("CEA-%d: Conformation Sync / Epoch %d / Conv %.2f", i+1, gem.State.EvolutionEpoch, gem.State.ConvergenceRate)
		artifacts[i+18] = fmt.Sprintf("AEA-%d: Agentic Spec / Epoch %d / Emerg %.2f", i+1, gem.State.EvolutionEpoch, gem.State.EmergenceIndex)
	}
	return artifacts
}
