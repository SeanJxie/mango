package mango

import (
	"math"
	"testing"
)

func TestBVHMatchesBruteForceIntersections(t *testing.T) {
	material := Diffuse{Albedo: NewSolidColourTextureAlbedo(RGB{R: 0.8, G: 0.8, B: 0.8})}
	shapes := []Shape{
		NewSphere(Vector3{X: -1.2, Y: 0, Z: -3}, 0.5, material),
		NewSphere(Vector3{X: 0.2, Y: -0.1, Z: -2}, 0.35, material),
		NewSphere(Vector3{X: 1.1, Y: 0.3, Z: -4}, 0.8, material),
		NewTriangle(
			Vector3{X: -2, Y: -0.75, Z: -1.5},
			Vector3{X: 2, Y: -0.75, Z: -1.5},
			Vector3{X: 0, Y: -0.75, Z: -4},
			material,
		),
	}
	bvh := BuildBVH(shapes, 0, len(shapes))

	rays := []Ray{
		{Origin: Vector3{X: 0, Y: 0, Z: 0}, Direction: Normalize3(Vector3{X: -0.35, Y: 0, Z: -1})},
		{Origin: Vector3{X: 0, Y: 0, Z: 0}, Direction: Normalize3(Vector3{X: 0.1, Y: -0.1, Z: -1})},
		{Origin: Vector3{X: 0, Y: 0.2, Z: 0}, Direction: Normalize3(Vector3{X: 0.25, Y: 0, Z: -1})},
		{Origin: Vector3{X: 0, Y: 1.5, Z: 0}, Direction: Normalize3(Vector3{X: 0, Y: -1, Z: -1})},
		{Origin: Vector3{X: 0, Y: 2, Z: 0}, Direction: Normalize3(Vector3{X: 0, Y: 1, Z: -1})},
	}

	for _, ray := range rays {
		bvhHit, bvhIntersection := bvh.Intersect(&ray, Epsilon, math.Inf(1))
		bruteHit, bruteIntersection := bruteForceIntersect(shapes, &ray, Epsilon, math.Inf(1))

		if bvhHit != bruteHit {
			t.Fatalf("BVH hit %v, brute-force hit %v for ray %+v", bvhHit, bruteHit, ray)
		}
		if bvhHit && math.Abs(bvhIntersection.T-bruteIntersection.T) > 1e-9 {
			t.Fatalf("BVH t %v, brute-force t %v for ray %+v", bvhIntersection.T, bruteIntersection.T, ray)
		}

		if bvh.IntersectBool(&ray, Epsilon, math.Inf(1)) != bruteForceIntersectBool(shapes, &ray, Epsilon, math.Inf(1)) {
			t.Fatalf("BVH boolean intersection mismatch for ray %+v", ray)
		}
	}
}

func bruteForceIntersect(shapes []Shape, ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {
	var closestIntersection *ShapeIntersection
	for _, shape := range shapes {
		hit, intersection := shape.Intersect(ray, tMin, tMax)
		if hit {
			tMax = intersection.T
			closestIntersection = intersection
		}
	}
	return closestIntersection != nil, closestIntersection
}

func bruteForceIntersectBool(shapes []Shape, ray *Ray, tMin, tMax float64) bool {
	for _, shape := range shapes {
		if shape.IntersectBool(ray, tMin, tMax) {
			return true
		}
	}
	return false
}
