package mango

type StratifiedSampler struct {
	Spp                          int
	Seed                         uint64
	gridX, gridY                 int
	x, y, sampleIndex, dimension int
}

func NewStratifiedSampler(samplesPerPixel int, seed uint64) *StratifiedSampler {
	gridX, gridY := stratifiedGrid(samplesPerPixel)
	return &StratifiedSampler{
		Spp:   samplesPerPixel,
		Seed:  seed,
		gridX: gridX,
		gridY: gridY,
	}
}

func stratifiedGrid(samplesPerPixel int) (int, int) {
	if samplesPerPixel <= 1 {
		return 1, 1
	}

	gridX := 1
	for gridX*gridX < samplesPerPixel {
		gridX++
	}

	gridY := samplesPerPixel / gridX
	if gridX*gridY < samplesPerPixel {
		gridY++
	}

	return gridX, gridY
}

func (s *StratifiedSampler) SamplesPerPixel() int {
	return s.Spp
}

func (s *StratifiedSampler) StartPixelSample(x, y, sampleIndex int) {
	s.x = x
	s.y = y
	s.sampleIndex = sampleIndex
	s.dimension = 0
}

func (s *StratifiedSampler) Clone(seed uint64) Sampler {
	return NewStratifiedSampler(s.Spp, splitMix64(s.Seed^seed))
}

func (s *StratifiedSampler) Sample1D() float64 {
	dimension := s.dimension
	s.dimension++
	return s.sampleDimension(dimension)
}

func (s *StratifiedSampler) Sample2D() Vector2 {
	dimension := s.dimension
	s.dimension += 2

	if dimension == 0 {
		return s.pixelSample()
	}

	return Vector2{
		X: s.sampleDimension(dimension),
		Y: s.sampleDimension(dimension + 1),
	}
}

func (s *StratifiedSampler) pixelSample() Vector2 {
	sampleIndex := s.permutedSampleIndex(0)
	cellX := sampleIndex % s.gridX
	cellY := sampleIndex / s.gridX

	jitterX := unitFloatFromHash(samplerHash(s.Seed, s.x, s.y, s.sampleIndex, 0))
	jitterY := unitFloatFromHash(samplerHash(s.Seed, s.x, s.y, s.sampleIndex, 1))

	return Vector2{
		X: sampleClamp((float64(cellX) + jitterX) / float64(s.gridX)),
		Y: sampleClamp((float64(cellY) + jitterY) / float64(s.gridY)),
	}
}

func (s *StratifiedSampler) sampleDimension(dimension int) float64 {
	prime := samplerPrimes[dimension%len(samplerPrimes)]
	index := uint64(s.permutedSampleIndex(dimension) + 1)
	shift := unitFloatFromHash(samplerHash(s.Seed, s.x, s.y, 0, dimension))
	u := radicalInverse(index, prime) + shift
	u -= float64(int(u))
	return sampleClamp(u)
}

func (s *StratifiedSampler) permutedSampleIndex(dimension int) int {
	if s.Spp <= 1 {
		return 0
	}
	offset := int(samplerHash(s.Seed, s.x, s.y, 0, dimension) % uint64(s.Spp))
	return (s.sampleIndex + offset) % s.Spp
}
