package mango

import "math"

// Implementation of Fresnel equations to describe amount if reflection/transmission of light.
func FresnelConductor(cosThetaI float64, etaI, etaT, k RGB) RGB {

	cosThetaI = Clamp(cosThetaI, -1, 1)

	eta := Div(etaT, etaI)
	etaK := Div(k, etaI)

	cosSquaredThetaI := cosThetaI * cosThetaI
	sinSquaredThetaI := 1 - cosSquaredThetaI

	etaSquared := Mul(eta, eta)
	etaKSquared := Mul(etaK, etaK)

	t0 := Subtract(Subtract(etaSquared, etaKSquared), RGB{sinSquaredThetaI, sinSquaredThetaI, sinSquaredThetaI})
	a2Plusb2 := SquareRoot(Add(Mul(t0, t0), Scale(Mul(etaSquared, etaKSquared), 4)))
	t1 := AddGrey(a2Plusb2, cosSquaredThetaI)
	a := SquareRoot(Scale(Add(a2Plusb2, t0), 0.5))
	t2 := Scale(a, 2*cosThetaI)
	rs := Div(Subtract(t1, t2), Add(t1, t2))

	t3 := AddGrey(Scale(a2Plusb2, cosSquaredThetaI), sinSquaredThetaI*sinSquaredThetaI)
	t4 := Scale(t2, sinSquaredThetaI)
	rp := Mul(rs, Div(Subtract(t3, t4), Add(t3, t4)))

	return Scale(Add(rp, rs), 0.5)
}

func FresnelDielectric(cosThetaI, etaI, etaT float64) float64 {
	cosThetaI = Clamp(cosThetaI, -1, 1)

	entering := cosThetaI > 0
	if !entering {
		etaI, etaT = etaT, etaI
		cosThetaI = math.Abs(cosThetaI)
	}

	sinThetaI := math.Sqrt(math.Max(0, 1-cosThetaI*cosThetaI))
	sinThetaT := etaI / etaT * sinThetaI
	if sinThetaT >= 1 {
		return 1
	}

	cosThetaT := math.Sqrt(math.Max(0, 1-sinThetaT*sinThetaT))
	rParallel := ((etaT * cosThetaI) - (etaI * cosThetaT)) / ((etaT * cosThetaI) + (etaI * cosThetaT))
	rPerpendicular := ((etaI * cosThetaI) - (etaT * cosThetaT)) / ((etaI * cosThetaI) + (etaT * cosThetaT))

	return 0.5 * (rParallel*rParallel + rPerpendicular*rPerpendicular)
}
