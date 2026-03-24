package trace

var (
	Black = RGB{0, 0, 0}
	White = RGB{1, 1, 1}
)

type RGB struct {
	R, G, B float64
}

func AsVector3(c RGB) Vector3 {
	return Vector3{c.R, c.G, c.B}
}

func IsBlack(c RGB) bool {
	return c.R == 0 && c.G == 0 && c.B == 0
}

func SkyBox(rayDirection Vector3) RGB {
	blueWeight := 0.5 * (rayDirection.Y + 1.0)
	return Blend(RGB{R: 0.5, G: 0.7, B: 1.0}, White, blueWeight)
}

func Add(c1, c2 RGB) RGB {
	return RGB{c1.R + c2.R, c1.G + c2.G, c1.B + c2.B}
}

func ScaleComponents(c RGB, scale RGB) RGB {
	return RGB{c.R * scale.R, c.G * scale.G, c.B * scale.B}
}

func Scale(c RGB, s float64) RGB {
	return RGB{c.R * s, c.G * s, c.B * s}
}

func Blend(c1, c2 RGB, factor float64) RGB {
	return Add(Scale(c1, factor), Scale(c2, 1-factor))
}

func Luminance(c RGB) float64 {
	return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
}
