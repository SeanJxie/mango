package mango

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"sync"
)

type ImageBuffer struct {
	Width, Height int

	// [r0, g0, b0, a0, r1, g1, b1, a1, ...]
	Pixels []float64

	SamplesPerPixelScale float64

	rowMu []sync.RWMutex
}

func NewImageBuffer(width, height, samplesPerPixel int) *ImageBuffer {
	return &ImageBuffer{
		Width:                width,
		Height:               height,
		Pixels:               make([]float64, width*height*4),
		SamplesPerPixelScale: 1.0 / float64(samplesPerPixel),
		rowMu:                make([]sync.RWMutex, height),
	}
}

// Adds a pixel sample to the buffer.
func (buf *ImageBuffer) AddSample(x, y int, colour RGB) {
	i := (y*buf.Width + x) * 4

	buf.rowMu[y].Lock()
	defer buf.rowMu[y].Unlock()

	buf.Pixels[i] += colour.R
	buf.Pixels[i+1] += colour.G
	buf.Pixels[i+2] += colour.B
	buf.Pixels[i+3] = 1 // alpha
}

func (buf *ImageBuffer) Output(filepath string) {
	img := buf.ToImage(buf.SamplesPerPixelScale)

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func (buf *ImageBuffer) OutputProgressive(filepath string, sampleNumber int) {
	img := buf.ToImage(1.0 / float64(max(sampleNumber, 1)))

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func (buf *ImageBuffer) PNGBytes(sampleScale float64) ([]byte, error) {
	var out bytes.Buffer
	if err := png.Encode(&out, buf.ToImage(sampleScale)); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (buf *ImageBuffer) ToImage(sampleScale float64) *image.RGBA {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{buf.Width, buf.Height}})
	pixels := buf.snapshotPixels()

	gamma := 2.2
	var r, g, b float64

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			i := (y*buf.Width + x) * 4

			// Average the samples.
			r = math.Max(0, pixels[i]*sampleScale)
			g = math.Max(0, pixels[i+1]*sampleScale)
			b = math.Max(0, pixels[i+2]*sampleScale)

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

	return img
}

func (buf *ImageBuffer) snapshotPixels() []float64 {
	pixels := make([]float64, len(buf.Pixels))
	rowStride := buf.Width * 4

	for y := 0; y < buf.Height; y++ {
		rowStart := y * rowStride
		rowEnd := rowStart + rowStride

		buf.rowMu[y].RLock()
		copy(pixels[rowStart:rowEnd], buf.Pixels[rowStart:rowEnd])
		buf.rowMu[y].RUnlock()
	}

	return pixels
}
