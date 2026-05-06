package mango

import "math"

type LightType int

const (
	Point LightType = iota
	DiffuseArea
)

func Visible(from, to Vector3, world Shape) bool {
	toLight := Subtract3(to, from)
	distance := Length3(toLight)
	if distance <= Epsilon {
		return false
	}

	ray := &Ray{from, ScalarMultiply3(toLight, 1/distance)}
	return !world.IntersectBool(ray, Epsilon, distance-Epsilon)
}

type Light interface {
	SampleLi(intersection *ShapeIntersection, sample Vector2) (RGB, Vector3, float64, Vector3)
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

func (light *PointLight) SampleLi(intersection *ShapeIntersection, sample Vector2) (RGB, Vector3, float64, Vector3) {
	wi := Normalize3(Subtract3(light.Position, intersection.Point))
	pdf := 1.0
	L := Scale(light.Intensity, 1.0/DistanceSquared3(light.Position, intersection.Point))

	return L, wi, pdf, light.Position
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

type DiffuseAreaLight struct {
	Type             LightType
	V0, EdgeU, EdgeV Vector3
	Normal, Center   Vector3
	Area             float64
	EmittedRadiance  RGB
}

func NewDiffuseAreaLight(v0, v1, v2, v3 Vector3, emittedRadiance RGB) *DiffuseAreaLight {
	edgeU := Subtract3(v1, v0)
	edgeV := Subtract3(v3, v0)
	cross := Cross(edgeU, edgeV)

	return &DiffuseAreaLight{
		Type:            DiffuseArea,
		V0:              v0,
		EdgeU:           edgeU,
		EdgeV:           edgeV,
		Normal:          Normalize3(cross),
		Center:          ScalarMultiply3(Add3(Add3(v0, v1), Add3(v2, v3)), 0.25),
		Area:            Length3(cross),
		EmittedRadiance: emittedRadiance,
	}
}

func (light *DiffuseAreaLight) SampleLi(intersection *ShapeIntersection, sample Vector2) (RGB, Vector3, float64, Vector3) {
	point := Add3(light.V0, Add3(ScalarMultiply3(light.EdgeU, sample.X), ScalarMultiply3(light.EdgeV, sample.Y)))
	toLight := Subtract3(point, intersection.Point)
	distanceSquared := LengthSquared3(toLight)
	if distanceSquared <= Epsilon*Epsilon || light.Area <= 0 {
		return Black, Zero3, 0, point
	}

	wi := ScalarMultiply3(toLight, 1/math.Sqrt(distanceSquared))
	cosThetaLight := Dot3(light.Normal, ScalarMultiply3(wi, -1))
	if cosThetaLight <= 0 {
		return Black, Zero3, 0, point
	}

	pdf := distanceSquared / (cosThetaLight * light.Area)
	return light.EmittedRadiance, wi, pdf, point
}

func (light *DiffuseAreaLight) Power() RGB {
	return Scale(light.EmittedRadiance, light.Area*Pi)
}

func (light *DiffuseAreaLight) GetPosition() Vector3 {
	return light.Center
}

func (light *DiffuseAreaLight) GetType() LightType {
	return light.Type
}
