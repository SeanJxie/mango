package mango

type Interval struct {
	Min, Max float64
}

func NewInterval(Min, Max float64) Interval {
	return Interval{Min, Max}
}

func NewIntervalFromUnion(a, b Interval) Interval {
	var uMin, uMax float64
	if a.Min <= b.Min {
		uMin = a.Min
	} else {
		uMin = b.Min
	}

	if a.Max >= b.Max {
		uMax = a.Max
	} else {
		uMax = b.Max
	}

	return NewInterval(uMin, uMax)
}

func (inter Interval) GetSize() float64 {
	return inter.Max - inter.Min
}

func (inter Interval) Clamp(x float64) float64 {
	return Clamp(x, inter.Min, inter.Max)
}

func (inter Interval) Expand(delta float64) Interval {
	pad := delta * 0.5
	return Interval{inter.Min - pad, inter.Max + pad}
}

// A 3D interval
type Aabb struct {
	x, y, z Interval
}

func NewAabbFromExtrema(a Vector3, b Vector3) *Aabb {
	// Sort coords
	var x, y, z Interval
	if a.X <= b.X {
		x = NewInterval(a.X, b.X)
	} else {
		x = NewInterval(b.X, a.X)
	}

	if a.Y <= b.Y {
		y = NewInterval(a.Y, b.Y)
	} else {
		y = NewInterval(b.Y, a.Y)
	}

	if a.Z <= b.Z {
		z = NewInterval(a.Z, b.Z)
	} else {
		z = NewInterval(b.Z, a.Z)
	}

	ret := &Aabb{x, y, z}
	ret.thickenIfTooThin()

	return ret
}

func NewAabbFromUnion(a *Aabb, b *Aabb) *Aabb {
	ret := &Aabb{
		NewIntervalFromUnion(a.x, b.x),
		NewIntervalFromUnion(a.y, b.y),
		NewIntervalFromUnion(a.z, b.z),
	}
	ret.thickenIfTooThin()

	return ret
}

func (box *Aabb) GetAxisInterval(n int) Interval {
	if n == 0 {
		return box.x
	}
	if n == 1 {
		return box.y
	}
	return box.z
}

func (box *Aabb) GetLongestAxis() int {
	xs := box.x.GetSize()
	ys := box.y.GetSize()
	zs := box.z.GetSize()

	if xs > ys {
		if xs > zs {
			return 0
		} else {
			return 2
		}
	} else {
		if ys > zs {
			return 1
		} else {
			return 2
		}
	}

}

func (box *Aabb) Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {
	rayOrigin := ray.Origin
	rayDir := ray.Direction

	for a := 0; a < 3; a++ {
		axis := box.GetAxisInterval(a)
		var rayDirAxis, rayOriginAxis float64
		switch a {
		case 0:
			rayDirAxis = rayDir.X
			rayOriginAxis = rayOrigin.X
		case 1:
			rayDirAxis = rayDir.Y
			rayOriginAxis = rayOrigin.Y
		case 2:
			rayDirAxis = rayDir.Z
			rayOriginAxis = rayOrigin.Z
		}
		rayDirAxisInv := 1.0 / rayDirAxis

		t0 := (axis.Min - rayOriginAxis) * rayDirAxisInv
		t1 := (axis.Max - rayOriginAxis) * rayDirAxisInv

		if t0 < t1 {
			if t0 > tMin {
				tMin = t0
			}
			if t1 < tMax {
				tMax = t1
			}
		} else {
			if t1 > tMin {
				tMin = t1
			}
			if t0 < tMax {
				tMax = t0
			}
		}

		if tMax <= tMin {
			return false, nil
		}
	}

	return true, nil
}

func (box *Aabb) IntersectBool(ray *Ray, tMin, tMax float64) bool {
	// Always returns nil *ShapeIntersection anyways.
	hit, _ := box.Intersect(ray, tMin, tMax)
	return hit
}

func (box *Aabb) thickenIfTooThin() {
	// If any dimension of the box gets too thin
	// (like in the case of planar primitves), expand them
	// them a little to avoid numerical issues.
	if box.x.GetSize() < Epsilon {
		box.x = box.x.Expand(Epsilon)
	}
	if box.y.GetSize() < Epsilon {
		box.y = box.y.Expand(Epsilon)
	}
	if box.z.GetSize() < Epsilon {
		box.z = box.z.Expand(Epsilon)
	}
}
