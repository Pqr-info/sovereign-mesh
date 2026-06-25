package cognition

import "math"

type Lineage struct {
	vector []float64
	lambda float64 // decay factor in (0,1)
}

func NewLineage(size int, lambda float64) *Lineage {
	return &Lineage{
		vector: make([]float64, size),
		lambda: lambda,
	}
}

func (l *Lineage) Update(Ck []float64) {
	n := len(l.vector)
	if len(Ck) < n {
		n = len(Ck)
	}
	for i := 0; i < n; i++ {
		// L(t+1) = λ L(t) + (1-λ) Ck(t)
		l.vector[i] = l.lambda*l.vector[i] + (1-l.lambda)*Ck[i]
	}
}

func (l *Lineage) Vector() []float64 {
	return l.vector
}

// Optional: scalar "lineage coherence"
func (l *Lineage) Magnitude() float64 {
	var sum float64
	for _, v := range l.vector {
		sum += v * v
	}
	return math.Sqrt(sum)
}
