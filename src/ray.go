package mango

type Ray struct {
	Origin    Vector3
	Direction Vector3
}

func (r *Ray) At(t float64) Vector3 {
	return Add3(r.Origin, ScalarMultiply3(r.Direction, t))
}
