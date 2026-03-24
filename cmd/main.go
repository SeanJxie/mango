package main

import (
	"math/rand/v2"

	"github.com/SeanJxie/trace/trace"
)

func main() {
	world := make([]trace.Shape, 0)

	checkerTex := trace.NewCheckeredTextureCol(0.32, trace.RGB{R: 0.2, G: 0.3, B: 0.1}, trace.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := trace.Lambertian{Albedo: checkerTex}
	world = append(world, trace.NewSphere(trace.Vector3{X: 0, Y: -1000, Z: 0}, 1000, ground_mat))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			center := trace.Vector3{X: float64(a) + 0.9*rand.Float64(), Y: 0.2, Z: float64(b) + 0.9*rand.Float64()}

			if trace.Length3(trace.Subtract3(center, trace.Vector3{X: 4, Y: 0.2, Z: 0})) > 0.9 {
				var sphere_mat trace.Material

				albedo := trace.ScaleComponents(
					trace.RGB{R: rand.Float64(), G: rand.Float64(), B: rand.Float64()},
					trace.RGB{R: rand.Float64(), G: rand.Float64(), B: rand.Float64()},
				)
				sphere_mat = trace.Lambertian{Albedo: trace.NewSolidColourTextureAlbedo(albedo)}
				world = append(world, trace.NewSphere(center, 0.2, sphere_mat))
			}
		}
	}

	mat2 := trace.Lambertian{Albedo: trace.NewSolidColourTextureAlbedo(trace.RGB{R: 0.4, G: 0.2, B: 0.1})}
	world = append(world, trace.NewSphere(trace.Vector3{X: -4, Y: 1, Z: 0}, 1.0, mat2))

	bvh := trace.NewBVH(world, 0, len(world))

	// Set up camera
	lookFrom := trace.Vector3{X: 13, Y: 2, Z: 3}
	lookAt := trace.Vector3{X: 0, Y: 0, Z: 0}
	camera := trace.NewPerspectiveCamera(lookFrom, lookAt, 1000.0/740.0, 20, 0, 10)

	imageBuffer := trace.NewImageBuffer(1000, 740)
	samplesPerPixel := 100

	integrator := trace.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Sampler:    &trace.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.SlowRender()
	imageBuffer.WriteToDisk(float64(samplesPerPixel))

}
