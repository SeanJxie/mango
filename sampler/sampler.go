package sampler

import (
	. "trace/geometry"
)

type Sampler interface {
	Sample1D() float64 // on [0, 1)
	Sample2D() Vector2 // on [0, 1)^2
}
