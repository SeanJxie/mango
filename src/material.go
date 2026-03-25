package mango

type Material interface {
	GetBSDF(intersection *ShapeIntersection) *BSDF
}

type Lambertian struct {
	Albedo Texture
}

func (m Lambertian) GetBSDF(intersection *ShapeIntersection) *BSDF {
	bsdf := &BSDF{}
	bsdf.AddBxDF(
		DiffuseReflection{
			m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point),
		},
	)

	return bsdf
}

type PerfectMirror struct {
	Albedo Texture
}

func (m PerfectMirror) GetBSDF(intersection *ShapeIntersection) *BSDF {
	bsdf := &BSDF{}
	bsdf.AddBxDF(
		SpecularReflection{
			m.Albedo.GetValue(intersection.U, intersection.V, intersection.Point),
		},
	)

	return bsdf
}

// type Metal struct {
// 	Albedo Vector3
// 	Fuzz   float64
// }

// func NewMetal(albedo Vector3, fuzz float64) Metal {
// 	if fuzz > 1.0 {
// 		fuzz = 1.0
// 	}
// 	return Metal{albedo, fuzz}
// }

// func (m Metal) Scatter(rIn Ray, rec HitInfo) (bool, Vector3, Ray) {
// 	reflected := Vector3reflect(rIn.D.Unit(), rec.Norm)
// 	attenuation := m.Albedo
// 	scattered := Ray{rec.P, reflected.Add(Vector3randInUnitSphere().Smul(m.Fuzz))}

// 	return scattered.D.Dot(rec.Norm) > 0.0, attenuation, scattered
// }

// type Dielectric struct {
// 	IndexOfRefraction float64
// }

// func NewDielectric(ior float64) Dielectric {
// 	return Dielectric{ior}
// }

// func dielectricReflectance(cosine float64, refIdx float64) float64 {
// 	r0 := (1.0 - refIdx) / (1.0 + refIdx)
// 	r0 = r0 * r0

// 	return r0 + (1.0-r0)*math.Pow(1.0-cosine, 5.0)
// }

// func (m Dielectric) Scatter(rIn Ray, rec HitInfo) (bool, Vector3, Ray) {
// 	var refractionRatio float64
// 	if rec.FrontFace {
// 		refractionRatio = 1.0 / m.IndexOfRefraction
// 	} else {
// 		refractionRatio = m.IndexOfRefraction
// 	}

// 	unitDir := rIn.D.Unit()
// 	cosTheta := math.Min(unitDir.Neg().Dot(rec.Norm), 1.0)
// 	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)

// 	cannotRefract := refractionRatio*sinTheta > 1.0

// 	var direction Vector3
// 	dr := dielectricReflectance(cosTheta, refractionRatio)
// 	if cannotRefract || (dr > rand.Float64()) {
// 		direction = Vector3reflect(unitDir, rec.Norm)
// 	} else {
// 		direction = Vector3refract(unitDir, rec.Norm, refractionRatio)
// 	}

// 	attenuation := Col{1.0, 1.0, 1.0}
// 	scattered := Ray{rec.P, direction}

// 	return true, attenuation, scattered
// }
