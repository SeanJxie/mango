package mango

import "math"

type Sampler interface {
	Sample1D() float64 // on [0, 1)
	Sample2D() Vector2 // on [0, 1)^2
	SamplesPerPixel() int
}

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
func SampleDiskConcentric(sample Vector2) Vector2 {
	// Map sample to [-1, 1)^2.
	sampleUnitSquare := Subtract2(ScalarMultiply2(sample, 2), Ones2) // Scale and shift to origin.

	// Avoid division by zero later on.
	if sampleUnitSquare.X == 0 && sampleUnitSquare.Y == 0 {
		return Zero2
	}

	// Concentric mapping.
	var theta, radius float64
	if math.Abs(sampleUnitSquare.X) > math.Abs(sampleUnitSquare.Y) {
		// Sample in east and west triangular wedges.
		radius = sampleUnitSquare.X
		theta = PiDiv4 * (sampleUnitSquare.Y / sampleUnitSquare.X)
	} else {
		// Sample in north and south triangular wedges.
		radius = sampleUnitSquare.Y
		theta = PiDiv2 - PiDiv4*(sampleUnitSquare.X/sampleUnitSquare.Y)
	}

	return ScalarMultiply2(Vector2{X: math.Cos(theta), Y: math.Sin(theta)}, radius)
}

func SampleHemisphereUniform(sample Vector2) Vector3 {
	z := sample.X
	r := math.Sqrt(1 - z*z)
	phi := 2 * Pi * sample.Y

	return Vector3{X: r * math.Cos(phi), Y: r * math.Sin(phi), Z: z}
}

func SampleHemisphereCosine(sample Vector2) Vector3 {
	d := SampleDiskConcentric(sample)
	z := math.Sqrt(math.Max(0, 1-d.X*d.X-d.Y*d.Y))

	return Vector3{d.X, d.Y, z}
}

func SampleSphereUniform(sample Vector2) Vector3 {
	z := 1 - 2*sample.X
	r := math.Sqrt(max(0, 1-z*z))
	phi := 2 * Pi * sample.Y

	return Vector3{r * math.Cos(phi), r * math.Sin(phi), z}
}

func SampleTriangleUniform(sample Vector2) Vector2 {
	sqrtSampleX := math.Sqrt(sample.X)

	return Vector2{1 - sqrtSampleX, sample.Y * sqrtSampleX}
}

func HemisphereCosineSampePdf(cosTheta float64) float64 {
	return math.Max(0, cosTheta) * PiInverse
}

func PowerHeuristic(nf int, fPdf float64, ng int, gPdf float64) float64 {
	f := float64(nf) * fPdf
	g := float64(ng) * gPdf
	return (f * f) / (f*f + g*g)
}
