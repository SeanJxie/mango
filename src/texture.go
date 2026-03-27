package mango

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

type Texture interface {
	GetValue(u, v float64, p Vector3) RGB
}

type SolidColourTexture struct {
	albedo RGB
}

func NewSolidColourTextureRGB(r, g, b float64) *SolidColourTexture {
	return &SolidColourTexture{RGB{r, g, b}}
}

func NewSolidColourTextureAlbedo(albedo RGB) *SolidColourTexture {
	return &SolidColourTexture{albedo}
}

func (tex SolidColourTexture) GetValue(u, v float64, p Vector3) RGB {
	return tex.albedo
}

type CheckeredTexture struct {
	invScale        float64
	texEven, texOdd Texture
}

func NewCheckeredTexture(scale float64, evenTexture, oddTexture Texture) *CheckeredTexture {
	return &CheckeredTexture{1.0 / scale, evenTexture, oddTexture}
}

func NewCheckeredTextureCol(scale float64, evenCol, oddCol RGB) *CheckeredTexture {
	return &CheckeredTexture{
		1.0 / scale,
		NewSolidColourTextureAlbedo(evenCol),
		NewSolidColourTextureAlbedo(oddCol),
	}
}

func (tex CheckeredTexture) GetValue(u, v float64, p Vector3) RGB {
	xInt := int(math.Floor(tex.invScale * p.X))
	yInt := int(math.Floor(tex.invScale * p.Y))
	zInt := int(math.Floor(tex.invScale * p.Z))

	if (xInt+yInt+zInt)%2 == 0 {
		return tex.texEven.GetValue(u, v, p)
	}
	return tex.texOdd.GetValue(u, v, p)
}

type ImageTexture struct {
	img image.Image
}

func NewImageTexture(filepath string) ImageTexture {
	file, err := os.Open(filepath)
	if err != nil {
		return ImageTexture{nil}
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return ImageTexture{nil}
	}

	return ImageTexture{img}
}

func (tex ImageTexture) GetValue(u, v float64, p Vector3) RGB {
	if tex.img == nil {
		// Missing image texture, so we use a placeholder colour.
		return RGB{R: 1, G: 0.08, B: 0.58}
	}

	u = Clamp(u, 0, 1)
	v = 1.0 - Clamp(v, 0, 1)

	i := int(u * float64(tex.img.Bounds().Dx()))
	j := int(v * float64(tex.img.Bounds().Dy()))

	r, g, b, _ := tex.img.At(i, j).RGBA()

	// Shift back to uint8
	return RGB{R: float64(r>>8) * ByteMaxInverse, G: float64(g>>8) * ByteMaxInverse, B: float64(b>>8) * ByteMaxInverse}
}
