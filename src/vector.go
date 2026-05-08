package mango

import (
	"math"
)

var (
	Zero2 = Vector2{}
	Ones2 = Vector2{1, 1}
	Zero3 = Vector3{}
	Ones3 = Vector3{1, 1, 1}
)

type Vector2 struct {
	X, Y float64
}

func Add2(u, v Vector2) Vector2 {
	return Vector2{u.X + v.X, u.Y + v.Y}
}

func Subtract2(u, v Vector2) Vector2 {
	return Vector2{u.X - v.X, u.Y - v.Y}
}

func ScalarMultiply2(u Vector2, c float64) Vector2 {
	return Vector2{c * u.X, c * u.Y}
}

func ElementMultiply2(u, v Vector2) Vector2 {
	return Vector2{u.X * v.X, u.Y * v.Y}
}

func Dot2(u, v Vector2) float64 {
	return u.X*v.X + u.Y*v.Y
}

func Length2(u Vector2) float64 {
	return math.Sqrt(u.X*u.X + u.Y*u.Y)
}

func Normalize2(u Vector2) Vector2 {
	return ScalarMultiply2(u, 1.0/Length2(u))
}

func NearZero2(u Vector2) bool {
	return u.X < Epsilon && u.Y < Epsilon
}

type Vector3 struct {
	X, Y, Z float64
}

func Add3(u, v Vector3) Vector3 {
	return Vector3{u.X + v.X, u.Y + v.Y, u.Z + v.Z}
}

func Subtract3(u, v Vector3) Vector3 {
	return Vector3{u.X - v.X, u.Y - v.Y, u.Z - v.Z}
}

func ScalarMultiply3(u Vector3, c float64) Vector3 {
	return Vector3{c * u.X, c * u.Y, c * u.Z}
}

func ElementMultiply3(u, v Vector3) Vector3 {
	return Vector3{u.X * v.X, u.Y * v.Y, u.Z * v.Z}
}

func ElementDivide3(u, v Vector3) Vector3 {
	return Vector3{u.X / v.X, u.Y / v.Y, u.Z / v.Z}
}

func Dot3(u, v Vector3) float64 {
	return u.X*v.X + u.Y*v.Y + u.Z*v.Z
}

func Cross(u, v Vector3) Vector3 {
	return Vector3{u.Y*v.Z - u.Z*v.Y, u.Z*v.X - u.X*v.Z, u.X*v.Y - u.Y*v.X}
}

func Length3(u Vector3) float64 {
	return math.Sqrt(u.X*u.X + u.Y*u.Y + u.Z*u.Z)
}
func LengthSquared3(u Vector3) float64 {
	return u.X*u.X + u.Y*u.Y + u.Z*u.Z
}

func DistanceSquared3(u, v Vector3) float64 {
	return LengthSquared3(Subtract3(u, v))
}

func Normalize3(u Vector3) Vector3 {
	return ScalarMultiply3(u, 1.0/Length3(u))
}

func IsZero3(u Vector3) bool {
	return u.X == 0 && u.Y == 0 && u.Z == 0
}

func NearZero3(u Vector3) bool {
	return u.X < Epsilon && u.Y < Epsilon && u.Z < Epsilon
}

func Min3(u, v Vector3) Vector3 {
	return Vector3{
		min(u.X, v.X),
		min(u.Y, v.Y),
		min(u.Z, v.Z),
	}
}

func Max3(u, v Vector3) Vector3 {
	return Vector3{
		max(u.X, v.X),
		max(u.Y, v.Y),
		max(u.Z, v.Z),
	}
}

func FaceDirection(u, direction Vector3) Vector3 {
	if Dot3(u, direction) < 0 {
		return ScalarMultiply3(u, -1)
	}
	return u
}
