package sovereign

import (
	"math"
)

// Factor27Regularizer calculates the deviation of a training trajectory from the 27-neutral invariant.
// The vector v(e) is a composite of Encoding, Positioning, and Persona/Tagging.
// We penalize the 'Phase Drift' that violates the geometric constraints of the swarm.
func Factor27Regularizer(v_actual, v_target []float64, phaseDrift float64) float64 {
	// 1. Geometric Loss: Euclidean distance from the 27-neutral trajectory
	var geomLoss float64
	for i := range v_actual {
		diff := v_actual[i] - v_target[i]
		geomLoss += diff * diff
	}

	// 2. Phase-Neutral Constraint: Exponential penalty for drift
	// Drift is defined as the scalar distance from the Factor-27 attractor
	const λ = 0.27 // The coupling constant to the Genome
	phasePenalty := λ * math.Exp(phaseDrift)

	return geomLoss + phasePenalty
}

// GradientVitality computes the 'Slope' of intelligence flux across the ledger.
func GradientVitality(loss float64, prevLoss float64) float64 {
	return (prevLoss - loss) / prevLoss
}
