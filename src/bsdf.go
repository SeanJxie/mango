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
	BxDFs []BxDF
}

func (b *BSDF) AddBxDF(bxdf BxDF) {
	b.Type |= bxdf.GetType()
	b.BxDFs = append(b.BxDFs, bxdf)
}

func (b *BSDF) Value(outDirection, inDirection Vector3) (weight RGB) {
	for _, bxdf := range b.BxDFs {
		weight = Add(weight, bxdf.Value(outDirection, inDirection))
	}
	return weight
}

func (b *BSDF) Pdf(outDirection, inDirection Vector3) (probability float64) {
	if len(b.BxDFs) == 0 {
		return 0
	}

	for _, bxdf := range b.BxDFs {
		probability += bxdf.Pdf(outDirection, inDirection)
	}
	return probability / float64(len(b.BxDFs))
}

func (b *BSDF) SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64) {
	weight, inDirection, probability, _ = b.SampleAndValueWithType(outDirection, sample)
	return weight, inDirection, probability
}

func (b *BSDF) SampleAndValueWithType(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64, sampledType BxDFType) {
	if len(b.BxDFs) == 0 {
		return Black, Zero3, 0, 0
	}

	component := min(int(sample.X*float64(len(b.BxDFs))), len(b.BxDFs)-1)
	sample.X = sample.X*float64(len(b.BxDFs)) - float64(component)
	sample.X = sampleClamp(sample.X)

	sampledBxDF := b.BxDFs[component]
	_, inDirection, sampledPdf := sampledBxDF.SampleAndValue(outDirection, sample)
	if sampledPdf <= 0 || math.IsNaN(sampledPdf) || math.IsInf(sampledPdf, 0) {
		return Black, Zero3, 0, 0
	}

	probability = b.Pdf(outDirection, inDirection)
	if probability <= 0 || math.IsNaN(probability) || math.IsInf(probability, 0) {
		return Black, Zero3, 0, 0
	}

	return b.Value(outDirection, inDirection), inDirection, probability, sampledBxDF.GetType()
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
	if SameSide(outDirection, inDirection) {
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

// Lambertian body reflection seen through a dielectric clearcoat.
type CoatedDiffuseReflection struct {
	Reflectance  RGB
	CoatStrength float64
	CoatEta      float64
}

func (b CoatedDiffuseReflection) Value(outDirection, inDirection Vector3) (weight RGB) {
	if !SameSide(outDirection, inDirection) {
		return Black
	}

	coatIn := b.coatFresnel(math.Abs(CosTheta(inDirection)))
	coatOut := b.coatFresnel(math.Abs(CosTheta(outDirection)))
	transmission := (1 - coatIn) * (1 - coatOut)

	return Scale(b.Reflectance, transmission*PiInverse)
}

func (b CoatedDiffuseReflection) Pdf(outDirection, inDirection Vector3) (probability float64) {
	if SameSide(outDirection, inDirection) {
		return math.Abs(CosTheta(inDirection)) * PiInverse
	}
	return 0
}

func (b CoatedDiffuseReflection) SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64) {
	inDirection = SampleHemisphereCosine(sample)
	if outDirection.Z < 0 {
		inDirection.Z *= -1
	}

	return b.Value(outDirection, inDirection), inDirection, b.Pdf(outDirection, inDirection)
}

func (b CoatedDiffuseReflection) GetType() BxDFType {
	return ReflectionBxDF | DiffuseBxDF
}

func (b CoatedDiffuseReflection) coatFresnel(cosTheta float64) float64 {
	strength := Clamp(b.CoatStrength, 0, 1)
	return strength * FresnelDielectric(cosTheta, 1, clearCoatEta(b.CoatEta))
}

// Cook-Torrance microfacet reflection using GGX (Trowbridge-Reitz) distribution.
type MicrofacetReflectionGGX struct {
	Reflectance       RGB
	Absorption        RGB
	IndexOfRefraction RGB
	Alpha             float64
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

	f := FresnelConductor(math.Abs(Dot3(inDirection, half)), White, b.IndexOfRefraction, b.Absorption)
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

	halfVector := Add3(outDirection, inDirection)
	if IsZero3(halfVector) {
		return 0
	}

	half := Normalize3(halfVector)
	denominator := 4 * math.Abs(Dot3(outDirection, half))
	if denominator == 0 {
		return 0
	}

	return TrowbridgeReitzPdf(outDirection, half, b.Alpha, b.Alpha) / denominator
}

func (b MicrofacetReflectionGGX) SampleAndValue(outDirection Vector3, sample Vector2) (weight RGB, inDirection Vector3, probability float64) {
	if outDirection.Z == 0 {
		return Black, Zero3, 0
	}

	wh := TrowbridgeReitzSampleWh(outDirection, sample, b.Alpha, b.Alpha)

	outDotHalf := Dot3(outDirection, wh)
	if outDotHalf <= 0 {
		return Black, Zero3, 0
	}

	inDirection = Reflect(outDirection, wh)

	if !SameSide(outDirection, inDirection) {
		return Black, Zero3, 0
	}

	probability = TrowbridgeReitzPdf(outDirection, wh, b.Alpha, b.Alpha) / (4 * math.Abs(outDotHalf))
	if probability <= 0 || math.IsNaN(probability) || math.IsInf(probability, 0) {
		return Black, Zero3, 0
	}

	return b.Value(outDirection, inDirection), inDirection, probability
}

func (b MicrofacetReflectionGGX) GetType() BxDFType {
	return ReflectionBxDF | GlossBxDF
}

// Dielectric GGX reflection lobe for a glossy clearcoat layer.
type ClearCoatReflection struct {
	Strength  float64
	Roughness float64
	Eta       float64
}

func (b ClearCoatReflection) Value(outDirection, inDirection Vector3) RGB {
	if !SameSide(outDirection, inDirection) {
		return Black
	}

	cosThetaOut := math.Abs(CosTheta(outDirection))
	cosThetaIn := math.Abs(CosTheta(inDirection))
	if cosThetaOut == 0 || cosThetaIn == 0 {
		return Black
	}

	halfVector := Add3(outDirection, inDirection)
	if IsZero3(halfVector) {
		return Black
	}

	half := FaceDirection(Normalize3(halfVector), Vector3{0, 0, 1})
	alpha := clearCoatAlpha(b.Roughness)
	f := Clamp(b.Strength, 0, 1) * FresnelDielectric(math.Abs(Dot3(inDirection, half)), 1, clearCoatEta(b.Eta))
	d := TrowbridgeReitzD(half, alpha, alpha)
	g := TrowbridgeReitzG(outDirection, inDirection, alpha, alpha)

	return Scale(White, f*d*g/(4*cosThetaIn*cosThetaOut))
}

func (b ClearCoatReflection) Pdf(outDirection, inDirection Vector3) float64 {
	if !SameSide(outDirection, inDirection) {
		return 0
	}

	halfVector := Add3(outDirection, inDirection)
	if IsZero3(halfVector) {
		return 0
	}

	half := Normalize3(halfVector)
	denominator := 4 * math.Abs(Dot3(outDirection, half))
	if denominator == 0 {
		return 0
	}

	alpha := clearCoatAlpha(b.Roughness)
	return TrowbridgeReitzPdf(outDirection, half, alpha, alpha) / denominator
}

func (b ClearCoatReflection) SampleAndValue(outDirection Vector3, sample Vector2) (RGB, Vector3, float64) {
	if outDirection.Z == 0 {
		return Black, Zero3, 0
	}

	alpha := clearCoatAlpha(b.Roughness)
	half := TrowbridgeReitzSampleWh(outDirection, sample, alpha, alpha)
	outDotHalf := Dot3(outDirection, half)
	if outDotHalf <= 0 {
		return Black, Zero3, 0
	}

	inDirection := Reflect(outDirection, half)

	if !SameSide(outDirection, inDirection) {
		return Black, Zero3, 0
	}

	pdf := TrowbridgeReitzPdf(outDirection, half, alpha, alpha) / (4 * math.Abs(outDotHalf))
	if pdf <= 0 || math.IsNaN(pdf) || math.IsInf(pdf, 0) {
		return Black, Zero3, 0
	}

	return b.Value(outDirection, inDirection), inDirection, pdf
}

func (b ClearCoatReflection) GetType() BxDFType {
	return ReflectionBxDF | GlossBxDF
}

func clearCoatEta(eta float64) float64 {
	if eta <= 0 {
		return 1.5
	}
	return math.Max(1.001, eta)
}

func clearCoatAlpha(roughness float64) float64 {
	return math.Max(0.001, TrowbridgeReitzRoughnessToAlpha(Clamp(roughness, 0.001, 1)))
}
