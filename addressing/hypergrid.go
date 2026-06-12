package addressing

// Quantize maps a 5-D coordinate into coarse hypergrid buckets.
type HypergridBucket struct {
	GX   int64
	GY   int64
	GZ   int64
	GT   int64
	GPsi int64
}

func Quantize(c Address5DCoord, step int64) HypergridBucket {
	if step <= 0 {
		step = 1
	}
	return HypergridBucket{
		GX:   c.X / step,
		GY:   c.Y / step,
		GZ:   c.Z / step,
		GT:   c.T / step,
		GPsi: c.Psi / step,
	}
}
