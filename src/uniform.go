package mango

// Hash-based white-noise sampler. Mostly useful as a baseline.
type UniformSampler struct {
	Spp                          int
	Seed                         uint64
	x, y, sampleIndex, dimension int
}

func (s *UniformSampler) SamplesPerPixel() int {
	return s.Spp
}

func (s *UniformSampler) StartPixelSample(x, y, sampleIndex int) {
	s.x = x
	s.y = y
	s.sampleIndex = sampleIndex
	s.dimension = 0
}

func (s *UniformSampler) Clone(seed uint64) Sampler {
	return &UniformSampler{Spp: s.Spp, Seed: splitMix64(s.Seed ^ seed)}
}

func (s *UniformSampler) Sample1D() float64 {
	dimension := s.dimension
	s.dimension++
	return sampleClamp(unitFloatFromHash(samplerHash(s.Seed, s.x, s.y, s.sampleIndex, dimension)))
}

func (s *UniformSampler) Sample2D() Vector2 {
	return Vector2{X: s.Sample1D(), Y: s.Sample1D()}
}
