package addressing

// WaferAlignment computes a simple tolerance metric between two 5-D coordinates.
func WaferAlignment(a, b Address5DCoord) float64 {
	dx := abs64(a.X - b.X)
	dy := abs64(a.Y - b.Y)
	dz := abs64(a.Z - b.Z)
	dpsi := abs64(a.Psi - b.Psi)
	// simple heuristic; replace with your wafer model
	return float64(dx + dy + dz + dpsi) // lower is better
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
