package addressing

import (
	"math"
	"sort"
)

// roundFloat ensures deterministic cross-platform behavior
func roundFloat(v float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(v*factor) / factor
}

// harmonic checks if two phase angles are harmonically related in the 27-state space
func harmonic(a, b int64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	// Direct alignment (0), neighbors (1), or wrap-around boundaries (26, 25)
	return diff == 0 || diff == 1 || diff == 25 || diff == 26
}

// CalculateAffinity computes a normalized affinity score [0.0–1.0] between two 5D coordinates
func CalculateAffinity(a, b Address5DCoord) float64 {
	// Euclidean distance in X/Y/Z space (bounds are [0, 1023])
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	dz := float64(a.Z - b.Z)

	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

	// Max possible distance in 1024x1024x1024 cube is sqrt(3 * 1023^2)
	maxDist := math.Sqrt(3.0 * 1023.0 * 1023.0)
	normDist := dist / maxDist

	// Convert distance → affinity (inverse relationship)
	affinity := 1.0 - normDist
	if affinity < 0.0 {
		affinity = 0.0
	}

	// Phase lock boost
	if harmonic(a.Psi, b.Psi) {
		affinity += 0.15 // small but meaningful boost
	}

	// Clamp to [0, 1]
	if affinity > 1.0 {
		affinity = 1.0
	}

	return roundFloat(affinity, 6)
}

// FindBestPartners returns the top N partners sorted by affinity
func FindBestPartners(anchor Address5DCoord, pool []Address5DCoord, limit int) []Address5DCoord {
	type scored struct {
		coord    Address5DCoord
		affinity float64
	}

	var scoredList []scored

	for _, p := range pool {
		score := CalculateAffinity(anchor, p)
		scoredList = append(scoredList, scored{coord: p, affinity: score})
	}

	// Sort by affinity descending
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].affinity > scoredList[j].affinity
	})

	// Limit results
	if limit > len(scoredList) {
		limit = len(scoredList)
	}

	result := make([]Address5DCoord, limit)
	for i := 0; i < limit; i++ {
		result[i] = scoredList[i].coord
	}

	return result
}
