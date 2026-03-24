package trace

type ShapeIntersection struct {
	Point, SurfaceNormal Vector3
	Mat                  Material
	Le                   RGB
	T                    float64
	U, V                 float64
	FrontFace            bool
}

func (si *ShapeIntersection) GetBSDF() *BSDF {
	return si.Mat.GetBSDF(si)
}

func (si *ShapeIntersection) SetFaceNormal(ray *Ray, outwardNormal Vector3) {
	si.FrontFace = Dot3(ray.Direction, outwardNormal) < 0.0
	if si.FrontFace {
		si.SurfaceNormal = outwardNormal
	} else {
		si.SurfaceNormal = ScalarMultiply3(outwardNormal, -1)
	}
}

func (si *ShapeIntersection) CastRay(direction Vector3) Ray {
	return Ray{si.Point, direction}
}

type Shape interface {
	Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection)
	GetBoundingBox() *Aabb
}
