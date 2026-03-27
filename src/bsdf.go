package mango

import (
	"math"
)

type BxDFType int

const (
	// TODO: transmission
	ReflectionBxDF BxDFType = 1 << 0

	DiffuseBxDF  = 1 << 2 // Basically random scattering
	GlossBxDF    = 1 << 3 // Scatters roughly in some direction
	SpecularBxDF = 1 << 4 // Perfectly reflective

	AllBxDF = ReflectionBxDF | DiffuseBxDF | GlossBxDF | SpecularBxDF
)

// Some helper trig functions defined on the local intersection space (where intersection normal is z-axis).

func CosTheta(u Vector3) float64 {
	return u.Z
}

func SinTheta(u Vector3) float64 {
	return math.Sqrt(SinSquaredTheta(u))
}

func TanTheta(u Vector3) float64 {
	return SinTheta(u) / CosTheta(u)
}

func CosSquaredTheta(u Vector3) float64 {
	return u.Z * u.Z
}

func SinSquaredTheta(u Vector3) float64 {
	return max(0, 1-CosSquaredTheta(u))
}

func TanSquaredTheta(u Vector3) float64 {
	return SinSquaredTheta(u) / CosSquaredTheta(u)
}

func CosPhi(u Vector3) float64 {
	sinTheta := SinTheta(u)
	if sinTheta == 0 {
		return 1
	}
	return Clamp(u.X/sinTheta, -1, 1)
}

func SinPhi(u Vector3) float64 {
	sinTheta := SinTheta(u)
	if sinTheta == 0 {
		return 0
	}
	return Clamp(u.Y/sinTheta, -1, 1)
}

func CosSquaredPhi(u Vector3) float64 {
	return CosPhi(u) * CosPhi(u)
}

func SinSquaredPhi(u Vector3) float64 {
	return SinPhi(u) * SinPhi(u)
}

// Checks if u and v are in the same side (hemisphere).
func SameSide(u, v Vector3) bool {
	return u.Z*v.Z > 0
}

func SphericalDirection(sinTheta, cosTheta, phi float64) Vector3 {
	return Vector3{sinTheta * math.Cos(phi), sinTheta * math.Sin(phi), cosTheta}
}

func Reflect(outDirection, normal Vector3) Vector3 {
	return Add3(ScalarMultiply3(outDirection, -1), ScalarMultiply3(normal, 2*Dot3(outDirection, normal)))
}

func Refract(inDirection, normal Vector3, eta float64) (bool, Vector3) {
	cosThetaI := Dot3(normal, inDirection)
	sinSquaredThetaI := max(0, 1-cosThetaI*cosThetaI)
	sinSquaredThetaT := eta * eta * sinSquaredThetaI

	if sinSquaredThetaT >= 1 {
		return false, Zero3
	}

	cosThetaT := math.Sqrt(1 - sinSquaredThetaT)

	return true, Add3(ScalarMultiply3(inDirection, -eta), ScalarMultiply3(normal, eta*cosThetaI-cosThetaT))
}

// Defined on a local basis where "up" is the normal of intersection.
type BxDF interface {

	// Given incoming and outgoing light directions, returns the fraction of light reflected/transmitted.
	Value(outDirection, inDirection Vector3) (weight RGB)

	// Returns the probability density that SampleAndValue would choose
	// the incoming light direction, given the outgoing light direction.
	Pdf(outDirection, inDirection Vector3) (probability float64)

	// Given an incoming light direction based on the outgoing light direction a random 2D sample,
	// returns the BxDF value for those directions and the probability density of the choice.
	SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64)

	// Returns type flag.
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

func (b *BSDF) Value(outDirection, inDirection Vector3) (weight RGB) {
	// The goal is to blend all the BxDFs in an even way via sampling.
	return b.BxDFs[0].Value(outDirection, inDirection)
}

func (b *BSDF) Pdf(outDirection, inDirection Vector3) (probability float64) {
	return b.BxDFs[0].Pdf(outDirection, inDirection)
}

func (b *BSDF) SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64) {
	return b.BxDFs[0].SampleAndValue(outDirection, sample)
}

func (b *BSDF) GetType() BxDFType {
	return b.Type
}

// BxDF definitions

// Lambertian (light scatters generally in all directions (hemisphere)).
type DiffuseReflection struct {
	Reflectance RGB
}

func (b DiffuseReflection) Value(outDirection, inDirection Vector3) (weight RGB) {
	return Scale(b.Reflectance, PiInverse)
}

func (b DiffuseReflection) Pdf(outDirection, inDirection Vector3) (probability float64) {
	// Cosine-weighted scattering. i.e. rays tend to scatter towards the top of the hemisphere
	// which is good because those are the rays that contribute the most (compared to the ones
	// that only graze the surface).
	if SameSide(inDirection, inDirection) {
		return math.Abs(CosTheta(inDirection)) * PiInverse
	}
	return 0
}

func (b DiffuseReflection) SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64) {
	inDirection = SampleHemisphereCosine(sample)
	return b.Value(outDirection, inDirection), inDirection, b.Pdf(outDirection, inDirection)
}

func (b DiffuseReflection) GetType() BxDFType {
	return ReflectionBxDF | DiffuseBxDF
}

// Cook-Torrance microfacet reflection using GGX (Trowbridge-Reitz) distribution.
// TODO: debug
type MicrofacetReflectionGGX struct {
	Reflectance RGB
	Alpha       float64
}

func (b MicrofacetReflectionGGX) Value(outDirection, inDirection Vector3) (weight RGB) {
	cosThetaOut := math.Abs(CosTheta(outDirection))
	cosThetaIn := math.Abs(CosTheta(inDirection))

	half := Add3(outDirection, inDirection)

	if cosThetaOut == 0 || cosThetaIn == 0 {
		return Black
	}

	if IsZero3(half) {
		return Black
	}

	half = Normalize3(half)
	half = FaceDirection(half, Vector3{0, 0, 1}) // Force to live in same hemsphere as surface normal.

	f := FresnelConductor(math.Abs(Dot3(inDirection, half)), White, RGB{0.18, 0.43, 1.38}, RGB{3.42, 2.45, 1.91})
	d := TrowbridgeReitzD(half, b.Alpha, b.Alpha)
	g := TrowbridgeReitzG(outDirection, inDirection, b.Alpha, b.Alpha)

	// Cook-Torrance.
	fdg := Scale(f, d*g)

	value := Scale(Mul(b.Reflectance, fdg), 1/(4*cosThetaIn*cosThetaOut))

	return value
}

func (b MicrofacetReflectionGGX) Pdf(outDirection, inDirection Vector3) (probability float64) {
	if !SameSide(outDirection, inDirection) {
		return 0
	}

	half := Normalize3(Add3(outDirection, inDirection))

	return TrowbridgeReitzPdf(outDirection, half, b.Alpha, b.Alpha) / (4 * Dot3(outDirection, half))
}

func (b MicrofacetReflectionGGX) SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64) {
	if outDirection.Z == 0 {
		return Black, Zero3, 0
	}

	wh := TrowbridgeReitzSampleWh(outDirection, sample, b.Alpha, b.Alpha)

	if Dot3(outDirection, wh) < 0 {
		wh = FaceDirection(wh, outDirection) // TODO: temporary fix.
	}

	inDirection = Reflect(outDirection, wh)

	if !SameSide(outDirection, inDirection) {
		inDirection = FaceDirection(inDirection, outDirection) // TODO: temporary fix.
	}

	probability = TrowbridgeReitzPdf(outDirection, wh, b.Alpha, b.Alpha) / (4 * Dot3(outDirection, wh))

	return b.Value(outDirection, inDirection), inDirection, probability
}

func (b MicrofacetReflectionGGX) GetType() BxDFType {
	return ReflectionBxDF | GlossBxDF
}

// Simple "cone-scattering" glossy BRDF.
type SimpleGlossyReflection struct {
	Reflectance RGB
	Roughness   float64
}

// Perfect mirror surface (light scatters in one direction only).
// type SpecularReflection struct {
// 	R RGB
// }

// func (b SpecularReflection) F(wo Vector3, wi Vector3) RGB {
// 	// When light just grazes the surface, underlying colour comes through less.
// 	cosTheta := CosTheta(wi)
// 	if cosTheta == 0 {
// 		return Black
// 	}
// 	return Scale(b.R, 1.0/math.Abs(cosTheta))
// }

// func (b SpecularReflection) Pdf(wo Vector3, wi Vector3) float64 {
// 	return 1.0
// }

// func (b SpecularReflection) SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64) {
// 	// Local space reflection (Z-axis is normal).
// 	wi := Vector3{-wo.X, -wo.Y, wo.Z}
// 	return b.F(wo, wi), wi, b.Pdf(wo, wi)
// }

// func (b SpecularReflection) GetType() BxDFType {
// 	return ReflectionBxDF | SpecularBxDF
// }
