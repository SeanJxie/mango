package mango

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

	// [r0, g0, b0, a0, r1, g1, b1, a1, ...]
	Pixels []float64

	SamplesPerPixelScale float64
}

func NewImageBuffer(width, height, samplesPerPixel int) *ImageBuffer {
	return &ImageBuffer{
		width,
		height,
		make([]float64, width*height*4),
		1.0 / float64(samplesPerPixel),
	}
}

// Adds a pixel sample to the buffer.
func (buf *ImageBuffer) AddSample(x, y int, colour RGB) {
	i := (y*buf.Width + x) * 4

	buf.Pixels[i] += colour.R
	buf.Pixels[i+1] += colour.G
	buf.Pixels[i+2] += colour.B
	buf.Pixels[i+3] = 1 // alpha
}

func (buf *ImageBuffer) Output(filepath string) {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{buf.Width, buf.Height}})

	gamma := 2.2
	var r, g, b float64

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			i := (y*buf.Width + x) * 4

			// Average the samples.
			r = buf.Pixels[i] * buf.SamplesPerPixelScale
			g = buf.Pixels[i+1] * buf.SamplesPerPixelScale
			b = buf.Pixels[i+2] * buf.SamplesPerPixelScale

			// Gamma encoding.
			r = math.Pow(r, 1/gamma)
			g = math.Pow(g, 1/gamma)
			b = math.Pow(b, 1/gamma)

			col := color.RGBA{
				// Convert to 8-bit int.
				uint8(255 * Clamp(r, 0, 0.999)),
				uint8(255 * Clamp(g, 0, 0.999)),
				uint8(255 * Clamp(b, 0, 0.999)),
				255,
			}

			img.Set(x, y, col)
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func (buf *ImageBuffer) OutputProgressive(filepath string, sampleNumber int) {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{buf.Width, buf.Height}})

	sampleScale := 1.0 / float64(sampleNumber)

	gamma := 2.2
	var r, g, b float64

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			i := (y*buf.Width + x) * 4

			// Average the samples.
			r = buf.Pixels[i] * sampleScale
			g = buf.Pixels[i+1] * sampleScale
			b = buf.Pixels[i+2] * sampleScale

			// Gamma encoding.
			r = math.Pow(r, 1/gamma)
			g = math.Pow(g, 1/gamma)
			b = math.Pow(b, 1/gamma)

			col := color.RGBA{
				// Convert to 8-bit int.
				uint8(255 * Clamp(r, 0, 0.999)),
				uint8(255 * Clamp(g, 0, 0.999)),
				uint8(255 * Clamp(b, 0, 0.999)),
				255,
			}

			img.Set(x, y, col)
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
