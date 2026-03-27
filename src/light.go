package mango

import "math"

type LightType int

const (
	Point LightType = iota
	DiffuseArea
)

func Visible(from, to Vector3, world Shape) bool {
	ray := &Ray{from, Normalize3(Subtract3(to, from))}
	return !world.IntersectBool(ray, Epsilon, math.Inf(0))
}

type Light interface {
	SampleLi(intersection *ShapeIntersection, sample Vector2) (RGB, Vector3, float64)
	GetPosition() Vector3
	Power() RGB
	GetType() LightType
}

type PointLight struct {
	Type      LightType
	Position  Vector3
	Intensity RGB
}

func NewPointLight(position Vector3, intensity RGB) *PointLight {
	return &PointLight{
		Point, position, intensity,
	}
}

func (light *PointLight) SampleLi(intersection *ShapeIntersection, sample Vector2) (RGB, Vector3, float64) {
	wi := Normalize3(Subtract3(light.Position, intersection.Point))
	pdf := 1.0
	L := Scale(light.Intensity, 1.0/DistanceSquared3(light.Position, intersection.Point))

	return L, wi, pdf
}

func (light *PointLight) Power() RGB {
	return Scale(light.Intensity, 4*Pi)
}

func (light *PointLight) GetPosition() Vector3 {
	return light.Position
}

func (light *PointLight) GetType() LightType {
	return light.Type
}

// type DiffuseAreaLight struct {
// 	EmittedRadiance RGB
// 	LightShape      Shape
// }

// func (light *DiffuseAreaLight) SampleLi(intersection *ShapeIntersection, sample Vector2) (RGB, Vector3, float64) {
// 	wi := Normalize3(Subtract3(light.Position, intersection.Point))
// 	pdf := 1.0
// 	L := Scale(light.Intensity, 1.0/DistanceSquared3(light.Position, intersection.Point))

// 	return L, wi, pdf
// }

// func (light *DiffuseAreaLight) Radiance(intersection *ShapeIntersection, outDirection Vector3) RGB {
// 	if Dot3(intersection.SurfaceNormal, outDirection) > 0 {
// 		return light.EmittedRadiance
// 	}
// 	return Black
// }

// func (light *DiffuseAreaLight) Power() RGB {
// 	return Scale(light.EmittedRadiance, light.LightShape.SurfaceArea()*Pi)
// }
