package mango

import "math"

type Cylinder struct {
	Bottom Vector3
	Top    Vector3
	Radius float64
	Open   bool
	Mat    Material
	Bbox   *Aabb
}

func NewCylinder(bottom, top Vector3, radius float64, open bool, material Material) *Cylinder {
	axis := Normalize3(Subtract3(top, bottom))

	ex := radius * math.Sqrt(max(0, 1-axis.X*axis.X))
	ey := radius * math.Sqrt(max(0, 1-axis.Y*axis.Y))
	ez := radius * math.Sqrt(max(0, 1-axis.Z*axis.Z))

	ext := Vector3{ex, ey, ez}

	pMin := Subtract3(Min3(bottom, top), ext)
	pMax := Subtract3(Max3(bottom, top), ext)

	return &Cylinder{
		Bottom: bottom,
		Top:    top,
		Radius: radius,
		Open:   open,
		Mat:    material,
		Bbox:   NewAabbFromExtrema(pMin, pMax),
	}
}

// func (c *Cylinder) Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {

// }
