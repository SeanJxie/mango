package trace

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

type ImageBuffer struct {
	Width, Height int
	Pixels        []float64
}

func NewImageBuffer(width, height int) *ImageBuffer {
	return &ImageBuffer{
		width,
		height,
		make([]float64, width*height*4),
	}
}

func (buf *ImageBuffer) AddSample(x, y int, colour RGB) {
	i := (y*buf.Width + x) * 4

	buf.Pixels[i] += colour.R
	buf.Pixels[i+1] += colour.G
	buf.Pixels[i+2] += colour.B
	buf.Pixels[i+3] = 1 // alpha
}

func (buf *ImageBuffer) WriteToDisk(samplePerPixel float64) {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{buf.Width, buf.Height}})

	samplePerPixelScale := 1.0 / samplePerPixel
	var r, g, b float64

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			i := (y*buf.Width + x) * 4

			r = buf.Pixels[i] * samplePerPixelScale
			g = buf.Pixels[i+1] * samplePerPixelScale
			b = buf.Pixels[i+2] * samplePerPixelScale

			col := color.RGBA{
				// Gamma correction
				uint8(255 * Clamp(math.Pow(r, 1.0/2.2), 0, 0.999)),
				uint8(255 * Clamp(math.Pow(g, 1.0/2.2), 0, 0.999)),
				uint8(255 * Clamp(math.Pow(b, 1.0/2.2), 0, 0.999)),
				255,
			}
			img.Set(x, y, col)
		}
	}

	f, err := os.Create("out.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
