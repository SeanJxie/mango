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
	Albedo    Texture
	Roughness float64
}

func (m Glossy) GetBSDF(intersection *ShapeIntersection) *BSDF {
	bsdf := &BSDF{}
	bsdf.AddBxDF(
		GlossyReflectionSimple{
			m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point),
			m.Roughness * m.Roughness,
		},
	)

	return bsdf
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
