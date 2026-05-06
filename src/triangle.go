package mango

type Triangle struct {
	v0, v1, v2 Vector3
	normal     Vector3
	Mat        Material
	Le         RGB
	bbox       *Aabb
}

func NewTriangle(v0, v1, v2 Vector3, material Material) *Triangle {
	return newTriangle(v0, v1, v2, material, Black)
}

func NewEmissiveTriangle(v0, v1, v2 Vector3, material Material, emittedRadiance RGB) *Triangle {
	return newTriangle(v0, v1, v2, material, emittedRadiance)
}

func newTriangle(v0, v1, v2 Vector3, material Material, emittedRadiance RGB) *Triangle {
	minVec := Vector3{min(v0.X, v1.X, v2.X), min(v0.Y, v1.Y, v2.Y), min(v0.Z, v1.Z, v2.Z)}
	maxVec := Vector3{max(v0.X, v1.X, v2.X), max(v0.Y, v1.Y, v2.Y), max(v0.Z, v1.Z, v2.Z)}

	edge1 := Subtract3(v1, v0)
	edge2 := Subtract3(v2, v0)
	normal := Normalize3(Cross(edge1, edge2)) // Assuming CCW winding order

	return &Triangle{
		v0:     v0,
		v1:     v1,
		v2:     v2,
		normal: normal,
		Mat:    material,
		Le:     emittedRadiance,
		bbox:   NewAabbFromExtrema(minVec, maxVec),
	}
}

func (tri Triangle) Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {
	// Moller-Trumbore
	// https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm

	a := tri.v0
	b := tri.v1
	c := tri.v2

	o := ray.Origin
	d := ray.Direction

	edge1 := Subtract3(b, a)
	edge2 := Subtract3(c, a)

	rayCrossEdge2 := Cross(d, edge2)
	det := Dot3(edge1, rayCrossEdge2)

	// Ray parallel to triangle
	if det > -Epsilon && det < Epsilon {
		return false, nil
	}

	invDet := 1.0 / det
	s := Subtract3(o, a)
	u := invDet * Dot3(s, rayCrossEdge2)

	// Ray passes outside edge2 bounds
	if u < 0.0 || u > 1.0 {
		return false, nil
	}

	sCrossEdge1 := Cross(s, edge1)
	v := invDet * Dot3(d, sCrossEdge1)

	// Ray passes outside edge1 bounds
	if v < 0.0 || u+v > 1.0 {
		return false, nil
	}

	// Ray intersects triangle
	// Compute intersection param t
	t := invDet * Dot3(edge2, sCrossEdge1)

	// Invalid t
	if t < tMin || tMax < t {
		return false, nil
	}

	var rec ShapeIntersection

	rec.T = t
	rec.Point = ray.At(t)
	rec.SetFaceNormal(ray, tri.normal)
	rec.Mat = tri.Mat
	if rec.IsFrontFace {
		rec.Le = tri.Le
	}

	return true, &rec
}

func (tri Triangle) IntersectBool(ray *Ray, tMin, tMax float64) bool {
	a := tri.v0
	b := tri.v1
	c := tri.v2

	o := ray.Origin
	d := ray.Direction

	edge1 := Subtract3(b, a)
	edge2 := Subtract3(c, a)

	rayCrossEdge2 := Cross(d, edge2)
	det := Dot3(edge1, rayCrossEdge2)

	// Ray parallel to triangle
	if det > -Epsilon && det < Epsilon {
		return false
	}

	invDet := 1.0 / det
	s := Subtract3(o, a)
	u := invDet * Dot3(s, rayCrossEdge2)

	// Ray passes outside edge2 bounds
	if u < 0.0 || u > 1.0 {
		return false
	}

	sCrossEdge1 := Cross(s, edge1)
	v := invDet * Dot3(d, sCrossEdge1)

	// Ray passes outside edge1 bounds
	if v < 0.0 || u+v > 1.0 {
		return false
	}

	// Ray intersects triangle
	// Compute intersection param t
	t := invDet * Dot3(edge2, sCrossEdge1)

	// Invalid t
	if t < tMin || tMax < t {
		return false
	}

	return true
}

func (tri Triangle) GetBoundingBox() *Aabb {
	return tri.bbox
}

func (tri Triangle) SurfaceArea() float64 {
	edge1 := Subtract3(tri.v1, tri.v0)
	edge2 := Subtract3(tri.v2, tri.v0)

	return 0.5 * Length3(Cross(edge1, edge2))
}

func (tri Triangle) Pdf() float64 {
	return 1 / tri.SurfaceArea()
}

func (tri Triangle) SamplePoint(sample Vector2) Vector3 {
	return Zero3
}
