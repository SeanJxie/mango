package mango

import (
	"math"
)

type PerspectiveCamera struct {
	// We define a camera by an orthonormal basis rather than a matrix transform.
	position, horizontal, vertical Vector3
	lowerLeft                      Vector3
	u, v, w                        Vector3
	lensRadius                     float64
}

func NewPerspectiveCamera(position, lookAt Vector3, aspectRatio, fieldOfView, lensRadius, focalDistance float64) *PerspectiveCamera {
	var out PerspectiveCamera

	theta := Deg2Rad * fieldOfView
	h := math.Tan(theta * 0.5)
	imagePlaneHeight := 2.0 * h
	imagePlaneWidth := aspectRatio * imagePlaneHeight

	out.w = Normalize3(Subtract3(position, lookAt))             // Camera facing direction
	out.u = Normalize3(Cross(Vector3{X: 0, Y: 1, Z: 0}, out.w)) // Camera right hand direction
	out.v = Cross(out.w, out.u)                                 // Camera up direction

	out.position = position
	out.horizontal = ScalarMultiply3(out.u, imagePlaneWidth*focalDistance)
	out.vertical = ScalarMultiply3(out.v, imagePlaneHeight*focalDistance)

	out.lowerLeft = Subtract3(Subtract3(Subtract3(position,
		ScalarMultiply3(out.horizontal, 0.5)),
		ScalarMultiply3(out.vertical, 0.5)),
		ScalarMultiply3(out.w, focalDistance))

	out.lensRadius = lensRadius

	return &out
}

func (cam *PerspectiveCamera) CastRay(u, v float64, sampler Sampler) *Ray {
	r := ScalarMultiply2(SampleDiskConcentric(sampler.Sample2D()), cam.lensRadius)
	offset := Add3(ScalarMultiply3(cam.u, r.X), ScalarMultiply3(cam.v, r.Y))

	xStep := ScalarMultiply3(cam.horizontal, u)
	yStep := ScalarMultiply3(cam.vertical, v)

	dir := Subtract3(Subtract3(Add3(Add3(cam.lowerLeft, xStep), yStep), cam.position), offset)

	return &Ray{
		Origin:    Add3(cam.position, offset),
		Direction: Normalize3(dir),
	}
}
