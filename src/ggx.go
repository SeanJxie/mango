package mango

import (
	"math"
)

func TrowbridgeReitzD(half Vector3, alphaX, alphaY float64) float64 {
	tanSquaredTheta := TanSquaredTheta(half)
	if math.IsInf(tanSquaredTheta, 0) {
		return 0
	}

	cosSquaredTheta := CosSquaredTheta(half)
	alphaXY := alphaX * alphaY

	// funny name haha
	cosTesseractedTheta := cosSquaredTheta * cosSquaredTheta
	e := (cosSquaredTheta/alphaXY + SinSquaredTheta(half)/alphaXY) * tanSquaredTheta

	return 1 / (Pi * alphaXY * cosTesseractedTheta * (1 + e) * (1 + e))
}

func TrowbridgeReitzLambda(direction Vector3, alphaX, alphaY float64) float64 {
	absTanTheta := math.Abs(TanTheta(direction))
	if math.IsInf(absTanTheta, 0) {
		return 0
	}

	alpha := math.Sqrt(CosSquaredPhi(direction)*alphaX*alphaX + SinSquaredPhi(direction)*alphaY*alphaY)
	alphaSquaredTanSquaredTheta := alpha * alpha * absTanTheta * absTanTheta

	return (-1 + math.Sqrt(1+alphaSquaredTanSquaredTheta)) * 0.5
}

func TrowbridgeReitzG1(direction Vector3, alphaX, alphaY float64) float64 {
	return 1 / (1 + TrowbridgeReitzLambda(direction, alphaX, alphaY))
}

func TrowbridgeReitzG(outDirection, inDirection Vector3, alphaX, alphaY float64) float64 {
	return 1 / (1 + TrowbridgeReitzLambda(outDirection, alphaX, alphaY) + TrowbridgeReitzLambda(inDirection, alphaX, alphaY))
}

func TrowbridgeReitzPdf(outDirection, half Vector3, alphaX, alphaY float64) float64 {
	return TrowbridgeReitzD(half, alphaX, alphaY) * TrowbridgeReitzG1(outDirection, alphaX, alphaY) * math.Abs(Dot3(outDirection, half)) /
		math.Abs(CosTheta(outDirection))
}

func TrowbridgeReitzSample11(cosTheta, U1, U2 float64) (float64, float64) {
	if cosTheta > .9999 {
		r := math.Sqrt(U1 / (1 - U1))
		phi := 6.28318530718 * U2
		return r * math.Cos(phi), r * math.Sin(phi)
	}

	sinTheta := math.Sqrt(max(0, 1-cosTheta*cosTheta))
	tanTheta := sinTheta / cosTheta
	a := 1 / tanTheta
	G1 := 2 / (1 + math.Sqrt(1+1/(a*a)))

	A := 2*U1/G1 - 1
	tmp := 1 / (A*A - 1)
	if tmp > 1e10 {
		tmp = 1e10
	}
	B := tanTheta
	D := math.Sqrt(max(B*B*tmp*tmp-(A*A-B*B)*tmp, 0))
	slope_x_1 := B*tmp - D
	slope_x_2 := B*tmp + D

	var slopeX float64
	if A < 0 || slope_x_2 > 1/tanTheta {
		slopeX = slope_x_1
	} else {
		slopeX = slope_x_2
	}

	var S float64
	if U2 > 0.5 {
		S = 1
		U2 = 2 * (U2 - 0.5)
	} else {
		S = -1
		U2 = 2 * (0.5 - U2)
	}

	z := (U2 * (U2*(U2*0.27385-0.73369) + 0.46341)) / (U2*(U2*(U2*0.093073+0.309420)-1) + 0.597999)

	slopeY := S * z * math.Sqrt(1+slopeX*slopeX)

	return slopeX, slopeY
}

func TrowbridgeReitzSample(inDirection Vector3, alphaX, alphaY, U1, U2 float64) Vector3 {
	inDirectionStretched := Normalize3(Vector3{alphaX * inDirection.X, alphaY * inDirection.Y, inDirection.Z})

	slopeX, slopeY := TrowbridgeReitzSample11(CosTheta(inDirectionStretched), U1, U2)

	tmp := CosPhi(inDirectionStretched)*slopeX - SinPhi(inDirectionStretched)*slopeY
	slopeY = SinPhi(inDirectionStretched)*slopeX + CosPhi(inDirectionStretched)*slopeY
	slopeX = tmp

	slopeX *= alphaX
	slopeY *= alphaY

	return Normalize3(Vector3{-slopeX, -slopeY, 1})
}

func TrowbridgeReitzSampleWh(outDirection Vector3, sample Vector2, alphaX, alphaY float64) Vector3 {
	var wh Vector3
	flip := outDirection.Z < 0
	if flip {
		wh = TrowbridgeReitzSample(ScalarMultiply3(outDirection, -1), alphaX, alphaY, sample.X, sample.Y)
		wh = ScalarMultiply3(wh, -1)
	} else {
		wh = TrowbridgeReitzSample(outDirection, alphaX, alphaY, sample.X, sample.Y)
	}

	return wh
}

func TrowbridgeReitzRoughnessToAlpha(roughness float64) float64 {
	return roughness * roughness
}
