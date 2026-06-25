package cognition

import (
	"math"
)

// CognitiveVectorBuilder aggregates all cognition layers into Ck.
type CognitiveVectorBuilder struct{}

// NewCognitiveVectorBuilder returns a new builder instance.
func NewCognitiveVectorBuilder() *CognitiveVectorBuilder {
	return &CognitiveVectorBuilder{}
}

// ------------------------------------------------------------
// BuildCk merges STMB, LTMS, HDE, PRM, SEL into a single vector.
// ------------------------------------------------------------

func (b *CognitiveVectorBuilder) BuildCk(
	stmbVec, ltmsVec, hdeVec, prmVec, selVec []float64,
) []float64 {

	// Normalize all vectors to same length by padding with zeros.
	maxLen := max(
		len(stmbVec),
		len(ltmsVec),
		len(hdeVec),
		len(prmVec),
		len(selVec),
	)

	stmb := pad(stmbVec, maxLen)
	ltms := pad(ltmsVec, maxLen)
	hde := pad(hdeVec, maxLen)
	prm := pad(prmVec, maxLen)
	sel := pad(selVec, maxLen)

	// Weighted merge:
	// SEL dominates (SEN node)
	// PRM and LTMS secondary
	// STMB and HDE tertiary
	Ck := make([]float64, maxLen)
	for i := 0; i < maxLen; i++ {
		Ck[i] =
			0.50*sel[i] + // Strategy Evolution Layer (dominant)
				0.20*prm[i] + // Pattern Recognition Module
				0.15*ltms[i] + // Long-Term Memory Store
				0.10*stmb[i] + // Short-Term Memory Buffer
				0.05*hde[i] // Hypothesis Delta Engine
	}

	return Ck
}

// ------------------------------------------------------------
// NormalizeCk scales Ck to unit magnitude.
// ------------------------------------------------------------

func (b *CognitiveVectorBuilder) NormalizeCk(Ck []float64) []float64 {
	var sum float64
	for _, v := range Ck {
		sum += v * v
	}
	mag := math.Sqrt(sum)

	if mag == 0 {
		return Ck
	}

	out := make([]float64, len(Ck))
	for i := range Ck {
		out[i] = Ck[i] / mag
	}
	return out
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

func pad(vec []float64, size int) []float64 {
	if len(vec) >= size {
		return vec
	}
	out := make([]float64, size)
	copy(out, vec)
	return out
}

func max(vals ...int) int {
	m := vals[0]
	for _, v := range vals {
		if v > m {
			m = v
		}
	}
	return m
}
