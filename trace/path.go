package trace

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

type PathIntegrator struct {
	World      Shape
	Camera     *PerspectiveCamera
	Sampler    Sampler
	Buffer     *ImageBuffer
	MaxBounces int
}

func (integ *PathIntegrator) Li(ray Ray) RGB {
	radiance := Black
	contributionWeight := White

	var foundIntersection bool
	var intersection *ShapeIntersection

	for b := 0; b < integ.MaxBounces; b++ {

		foundIntersection, intersection = integ.World.Intersect(&ray, 0.001, math.Inf(0))
		if !foundIntersection {
			// Missed the scene, sample skybox.
			radiance = Add(radiance, ScaleComponents(SkyBox(ray.Direction), contributionWeight))
			break
		}

		bsdf := intersection.GetBSDF()
		localBasis := NewLocalBasis(intersection.SurfaceNormal)

		wo := ScalarMultiply3(ray.Direction, -1)
		woLocal := StandardToLocalBasis(wo, localBasis)

		f, wiLocal, pdf := bsdf.SampleF(woLocal, integ.Sampler.Sample2D())
		wiWorld := LocalToStandardBasis(wiLocal, localBasis)

		tmp := Scale(f, Dot3(wiWorld, intersection.SurfaceNormal)/pdf)
		contributionWeight = ScaleComponents(contributionWeight, tmp)

		ray = intersection.CastRay(wiWorld)
	}

	return radiance
}

func (integ *PathIntegrator) SlowRender() {
	imageBuffer := integ.Buffer
	width := integ.Buffer.Width
	height := integ.Buffer.Height
	camera := integ.Camera
	sampler := integ.Sampler
	samplesPerPixel := sampler.SamplesPerPixel()

	var wg sync.WaitGroup
	var completedRows int64

	for y := 0; y < height; y++ {
		wg.Add(1)

		go func(y int) {
			defer wg.Done()

			localSampler := UniformSampler{samplesPerPixel}

			for x := 0; x < width; x++ {
				pixelColour := Black
				for s := 0; s < samplesPerPixel; s++ {
					u := (float64(x) + localSampler.Sample1D()) / float64(width)
					v := (float64(height-(y+1)) + localSampler.Sample1D()) / float64(height)

					camRay := camera.CastRay(u, v, localSampler)
					pixelColour = Add(pixelColour, integ.Li(*camRay))
				}

				imageBuffer.AddSample(x, y, pixelColour)
			}

			count := atomic.AddInt64(&completedRows, 1)

			fmt.Printf("\rRender progress: %.2f%% (%d/%d rows)",
				100.0*float64(count)/float64(height), count, height)
		}(y)
	}

	// Wait for all rows to finish computing
	wg.Wait()
}
