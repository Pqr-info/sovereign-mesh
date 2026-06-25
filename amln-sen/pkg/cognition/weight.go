package cognition

// WeightEngine computes the agentic weight α for the AMLN node.
type WeightEngine struct {
	sel *SEL
}

// NewWeightEngine returns a new weight engine bound to SEL.
func NewWeightEngine(sel *SEL) *WeightEngine {
	return &WeightEngine{
		sel: sel,
	}
}

// ------------------------------------------------------------
// ComputeAlpha returns the agentic weight α ∈ [0,1].
// ------------------------------------------------------------

func (w *WeightEngine) ComputeAlpha() float64 {
	// Delegates to SEL's magnitude-based normalization.
	return w.sel.AgenticWeight()
}
