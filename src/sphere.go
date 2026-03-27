package mango

import (
	"math"
)

type Sphere struct {
	Center Vector3
	Radius float64
	Mat    Material
	Bbox   *Aabb
}

func NewSphere(center Vector3, radius float64, material Material) Sphere {
	rvec := Vector3{X: radius, Y: radius, Z: radius}
	return Sphere{center, radius, material, NewAabbFromExtrema(Subtract3(center, rvec), Add3(center, rvec))}
}

func (s Sphere) Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {
	offset := Subtract3(ray.Origin, s.Center)

	a := Dot3(ray.Direction, ray.Direction)
	bHalf := Dot3(offset, ray.Direction)
	c := LengthSquared3(offset) - s.Radius*s.Radius
	d := bHalf*bHalf - a*c

	if d < 0 {
		return false, nil
	}

	dSqrt := math.Sqrt(d)

	aInv := 1 / a

	root := (-bHalf - dSqrt) * aInv
	if root < tMin || tMax < root {
		root = (-bHalf + dSqrt) * aInv
		if root < tMin || tMax < root {
			return false, nil
		}
	}

	var isect ShapeIntersection

	isect.T = root
	isect.Point = ray.At(root)
	outwardNormal := ScalarMultiply3(Subtract3(isect.Point, s.Center), 1/s.Radius)
	isect.SetFaceNormal(ray, outwardNormal)
	isect.Mat = s.Mat
	isect.U, isect.V = getSphereUV(outwardNormal)

	return true, &isect
}

func (s Sphere) IntersectBool(ray *Ray, tMin, tMax float64) bool {
	offset := Subtract3(ray.Origin, s.Center)

	a := Dot3(ray.Direction, ray.Direction)
	bHalf := Dot3(offset, ray.Direction)
	c := LengthSquared3(offset) - s.Radius*s.Radius
	d := bHalf*bHalf - a*c

	if d < 0 {
		return false
	}

	dSqrt := math.Sqrt(d)
	aInv := 1 / a

	root := (-bHalf - dSqrt) * aInv
	if root < tMin || tMax < root {
		root = (-bHalf + dSqrt) * aInv
		if root < tMin || tMax < root {
			return false
		}
	}

	return true
}

func (s Sphere) GetBoundingBox() *Aabb {
	return s.Bbox
}

func (s Sphere) SurfaceArea() float64 {
	return 4 * Pi * s.Radius * s.Radius
}

func (s Sphere) Pdf() float64 {
	return 1 / s.SurfaceArea()
}

func (s Sphere) SamplePoint(sample Vector2) Vector3 {
	// Scale unit sphere to match radius and translate to world space.
	return Add3(s.Center, ScalarMultiply3(SampleSphereUniform(sample), s.Radius))
}

func getSphereUV(p Vector3) (float64, float64) {
	theta := math.Acos(-p.Y)
	phi := math.Atan2(-p.Z, p.X) + Pi

	return 0.5 * phi * PiInverse, theta * PiInverse
}
