package mango

import "math"

type BxDFType int

const (
	// TODO: transmission
	ReflectionBxDF BxDFType = 1 << 0

	DiffuseBxDF  = 1 << 2 // Basically random scattering
	GlossBxDF    = 1 << 3 // Scatters roughly in some direction
	SpecularBxDF = 1 << 4 // Perfectly reflective

	AllBxDF = ReflectionBxDF | DiffuseBxDF | GlossBxDF | SpecularBxDF
)

// Functions defined on the local space (normal is z-axis).
//
//         z
//         |     u
//         |----*-----> cos(theta) is just the Z value of u
//         |   /
//         |  /
//         | /
//         |/ <--- Angle here is theta
//   -------------
//
func CosTheta(u Vector3) float64 {
	return u.Z
}

func SameSide(u, v Vector3) bool {
	return u.Z*v.Z > 0
}

// Defined on a local basis determined by point of intersection.
type BxDF interface {
	F(wo, wi Vector3) RGB
	Pdf(wo, wi Vector3) float64
	SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64)
	GetType() BxDFType
}

// Sum of BxDFs
type BSDF struct {
	Type  BxDFType
	BxDFs []BxDF // TODO: only supports 1 BxDF so far.
}

func (b *BSDF) AddBxDF(bxdf BxDF) {
	b.Type |= bxdf.GetType()
	b.BxDFs = append(b.BxDFs, bxdf)
}

func (b *BSDF) F(wo, wi Vector3) RGB {
	// The goal is to blend all the BxDFs in an even way via sampling.
	return b.BxDFs[0].F(wo, wi)
}

func (b *BSDF) Pdf(wo, wi Vector3) float64 {
	return b.BxDFs[0].Pdf(wo, wi)
}

func (b *BSDF) SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64) {

	return b.BxDFs[0].SampleF(wo, u)
}

func (b *BSDF) GetType() BxDFType {
	return b.Type
}

// BxDF definitions

// Cosine-weighted scattering. i.e. rays tend to scatter towards the top of the hemisphere
// which is good because those are the rays that contribute the most (compared to the ones
// that only graze the surface).
//
// func DefaultPdf(wo, wi Vector3) float64 {
// 	if SameSide(wi, wo) {
// 		return math.Abs(CosTheta(wi)) * PiInverse
// 	}
// 	return 0
// }

// Lambertian (light scatters generally in all directions (hemisphere)).
type DiffuseReflection struct {
	R RGB
}

func (b DiffuseReflection) F(wo, wi Vector3) RGB {
	return ScaleRGB(b.R, PiInverse)
}

func (b DiffuseReflection) Pdf(wo, wi Vector3) float64 {
	if SameSide(wi, wo) {
		return math.Abs(CosTheta(wi)) * PiInverse
	}
	return 0
}

func (b DiffuseReflection) SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64) {
	wi := SampleHemisphereCosine(u)
	return b.F(wo, wi), wi, b.Pdf(wo, wi)
}

func (b DiffuseReflection) GetType() BxDFType {
	return ReflectionBxDF | DiffuseBxDF
}

// Perfect mirror surface (light scatters in one direction only).
type SpecularReflection struct {
	R RGB
}

func (b SpecularReflection) F(wo Vector3, wi Vector3) RGB {
	// When light just grazes the surface, underlying colour comes through more.
	cosTheta := CosTheta(wi)
	if cosTheta == 0 {
		return Black
	}
	return ScaleRGB(b.R, 1.0/math.Abs(cosTheta))
}

func (b SpecularReflection) Pdf(wo Vector3, wi Vector3) float64 {
	return 1.0
}

func (b SpecularReflection) SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64) {
	// Local space reflection (Z-axis is normal).
	wi := Vector3{-wo.X, -wo.Y, wo.Z}
	return b.F(wo, wi), wi, b.Pdf(wo, wi)
}

func (b SpecularReflection) GetType() BxDFType {
	return ReflectionBxDF | SpecularBxDF
}
