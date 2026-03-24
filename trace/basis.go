package trace

import "math"

type LocalBasis struct {
	U, V, W Vector3
}

func NewLocalBasis(normal Vector3) *LocalBasis {
	w := Normalize3(normal)

	var a Vector3
	if math.Abs(w.X) > 0.9 {
		a = Vector3{0, 1, 0}
	} else {
		a = Vector3{1, 0, 0}
	}

	v := Normalize3(Cross(w, a))
	u := Cross(w, v)

	return &LocalBasis{u, v, w}
}

// Change of basis from {U, V, W} to standard {(1, 0, 0), (0, 1, 0), (0, 0, 1)}.
// Given, local coordinates (x, y, z), we have standard coordinates U*x + V*y + W*z.
func LocalToStandardBasis(u Vector3, b *LocalBasis) Vector3 {
	return Add3(Add3(ScalarMultiply3(b.U, u.X), ScalarMultiply3(b.V, u.Y)), ScalarMultiply3(b.W, u.Z))
}

func StandardToLocalBasis(u Vector3, b *LocalBasis) Vector3 {
	return Vector3{Dot3(u, b.U), Dot3(u, b.V), Dot3(u, b.W)}
}
