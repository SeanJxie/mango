package mango

type Material interface {
	GetBSDF(intersection *ShapeIntersection) *BSDF
}

type Diffuse struct {
	Albedo Texture
}

func (m Diffuse) GetBSDF(intersection *ShapeIntersection) *BSDF {
	bsdf := &BSDF{}
	bsdf.AddBxDF(
		DiffuseReflection{
			m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point),
		},
	)

	return bsdf
}

type Metal struct {
	Albedo            Texture
	Absorption        RGB
	IndexOfRefraction RGB
	Roughness         float64
}

func (m Metal) GetBSDF(intersection *ShapeIntersection) *BSDF {
	bsdf := &BSDF{}
	bsdf.AddBxDF(
		MicrofacetReflectionGGX{
			m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point),
			m.Absorption,
			m.IndexOfRefraction,
			TrowbridgeReitzRoughnessToAlpha(m.Roughness),
		},
	)

	return bsdf
}

type Glossy struct {
	Albedo            Texture
	Roughness         float64
	ClearCoat         float64
	IndexOfRefraction float64
}

func (m Glossy) GetBSDF(intersection *ShapeIntersection) *BSDF {
	bsdf := &BSDF{}
	albedo := m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point)
	coatStrength := m.clearCoatStrength()
	coatEta := m.indexOfRefraction()

	bsdf.AddBxDF(
		CoatedDiffuseReflection{
			Reflectance:  albedo,
			CoatStrength: coatStrength,
			CoatEta:      coatEta,
		},
	)
	if coatStrength > 0 {
		bsdf.AddBxDF(
			ClearCoatReflection{
				Strength:  coatStrength,
				Roughness: m.roughness(),
				Eta:       coatEta,
			},
		)
	}

	return bsdf
}

func (m Glossy) clearCoatStrength() float64 {
	if m.ClearCoat == 0 {
		return 1
	}
	return Clamp(m.ClearCoat, 0, 1)
}

func (m Glossy) roughness() float64 {
	if m.Roughness <= 0 {
		return 0.1
	}
	return Clamp(m.Roughness, 0.001, 1)
}

func (m Glossy) indexOfRefraction() float64 {
	if m.IndexOfRefraction <= 0 {
		return 1.5
	}
	return m.IndexOfRefraction
}

// type PerfectMirror struct {
// 	Albedo Texture
// }

// func (m PerfectMirror) GetBSDF(intersection *ShapeIntersection) *BSDF {
// 	bsdf := &BSDF{}
// 	bsdf.AddBxDF(
// 		SpecularReflection{
// 			m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point),
// 		},
// 	)

// 	return bsdf
// }
