package sampler

import (
	"math"

	. "trace/geometry"
	. "trace/util"
)

// Map sample ([0, 1)^2) to a point on the unit disk that
// keeps an nicely-spread distribution.
//
// We can't just pick a random polar theta and r since
// points are clumped together for small r.
//
// So we use a concentric mapping which maps triangular wedges of the
// unit square to rounded triangular wedges of the unit dist.
//
// Desmos visualization: https://www.desmos.com/calculator/cuu4x8ohou
func SampleToDiskConcentric(sample Vector2) Vector2 {
	// Map sample to [-1, 1)^2
	sampleUnitSquare := Subtract2(ScalarMultiply2(sample, 2), Ones2) // Scale and shift to origin

	// Avoid division by zero later on
	if sampleUnitSquare.X == 0 && sampleUnitSquare.Y == 0 {
		return Zero2
	}

	// Concentric mapping
	var theta, radius float64
	if math.Abs(sampleUnitSquare.X) > math.Abs(sampleUnitSquare.Y) {
		// Sample in east and west triangular wedges
		radius = sampleUnitSquare.X
		theta = PiDiv4 * (sampleUnitSquare.Y / sampleUnitSquare.X)
	} else {
		// Sample in north and south triangular wedges
		radius = sampleUnitSquare.Y
		theta = PiDiv2 - PiDiv4*(sampleUnitSquare.X/sampleUnitSquare.Y)
	}

	return ScalarMultiply2(Vector2{X: math.Cos(theta), Y: math.Sin(theta)}, radius)
}
