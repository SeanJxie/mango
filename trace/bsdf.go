package trace

type BxDFType int

const (
	Reflection   BxDFType = 1 << 0
	Transmission          = 1 << 1
	Diffuse               = 1 << 2
	Gloss                 = 1 << 3
	Specular              = 1 << 4

	DiffuseReflection    = Diffuse | Reflection
	DiffuseTransmission  = Diffuse | Transmission
	GlossyReflection     = Gloss | Reflection
	GlossyTransmission   = Gloss | Transmission
	SpecularReflection   = Specular | Reflection
	SpecularTransmission = Specular | Transmission
	All                  = Diffuse | Gloss | Specular | Reflection | Transmission
)

// Defined on a local basis determined by point of intersection.
type BxDF interface {
	F(wo Vector3, wi Vector3) RGB
	Pdf(wo Vector3, wi Vector3) float64
	SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64)
	GetType() BxDFType
}

type BSDF struct {
	BxDFs []BxDF // TODO: only supports 1 BxDF so far.
}

func (b *BSDF) AddBxDF(bxdf BxDF) {
	b.BxDFs = append(b.BxDFs, bxdf)
}

func (b *BSDF) F(wo, wi Vector3) RGB {
	return b.BxDFs[0].F(wo, wi)
}

func (b *BSDF) SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64) {
	return b.BxDFs[0].SampleF(wo, u)
}

// BxDF definitions

type LambertianReflection struct {
	Colour RGB
}

func (b LambertianReflection) F(wo Vector3, wi Vector3) RGB {
	if wi.Z <= 0 {
		return Black
	}
	return Scale(b.Colour, PiInverse)
}

func (b LambertianReflection) Pdf(wo Vector3, wi Vector3) float64 {
	if wi.Z <= 0 {
		return 0
	}
	return wi.Z * PiInverse
}

func (b LambertianReflection) SampleF(wo Vector3, u Vector2) (RGB, Vector3, float64) {
	wi := SampleHemisphereCosine(u)

	if wo.Z <= 0 {
		wi.Z *= -1
	}

	pdf := b.Pdf(wo, wi)
	if pdf == 0 {
		return Black, wi, 0
	}

	return b.F(wo, wi), wi, pdf
}

func (b LambertianReflection) GetType() BxDFType {
	return DiffuseReflection
}
