package sampler

import (
	"math/rand"

	. "trace/geometry"
)

// Samples using Go's standard random library.
type UniformSampler struct{}

func (s *UniformSampler) Sample1D() float64 {
	return rand.Float64()
}

func (s *UniformSampler) Sample2D() Vector2 {
	return Vector2{X: rand.Float64(), Y: rand.Float64()}
}
