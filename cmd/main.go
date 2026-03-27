package main

import (
	"math/rand/v2"

	mango "github.com/SeanJxie/mango/src"
)

func main() {
	scene := 8
	switch scene {
	case 1:
		scene1()
	case 2:
		scene2()
	case 3:
		scene3()
	case 4:
		scene4()
	case 5:
		scene5()
	case 6:
		scene6()
	case 7:
		scene7()
	case 8:
		scene8()
	}
}

func scene1() {
	world := make([]mango.Shape, 0)

	checkerTex := mango.NewCheckeredTextureCol(0.32, mango.RGB{R: 0.2, G: 0.3, B: 0.1}, mango.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := mango.Diffuse{Albedo: checkerTex}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1000, Z: 0}, 1000, ground_mat))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			center := mango.Vector3{X: float64(a) + 0.9*rand.Float64(), Y: 0.2, Z: float64(b) + 0.9*rand.Float64()}

			if mango.Length3(mango.Subtract3(center, mango.Vector3{X: 4, Y: 0.2, Z: 0})) > 0.9 {
				var sphere_mat mango.Material

				albedo := mango.Mul(
					mango.RGB{R: rand.Float64(), G: rand.Float64(), B: rand.Float64()},
					mango.RGB{R: rand.Float64(), G: rand.Float64(), B: rand.Float64()},
				)
				sphere_mat = mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(albedo)}
				world = append(world, mango.NewSphere(center, 0.2, sphere_mat))
			}
		}
	}

	mat2 := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.4, G: 0.2, B: 0.1})}
	world = append(world, mango.NewSphere(mango.Vector3{X: -4, Y: 1, Z: 0}, 1.0, mat2))

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 13, Y: 2, Z: 3}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1000.0/740.0, 20, 0, 10)

	samplesPerPixel := 100
	imageBuffer := mango.NewImageBuffer(1000, 740, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.TileRenderProgressiveParallel(32)
	//integrator.ScanlineRenderParallel()
	imageBuffer.Output("out.png")
}

func scene2() {
	world := make([]mango.Shape, 0)
	triangles := mango.ParseOBJ("./lucy.obj")

	for _, t := range triangles {
		world = append(world, t)
	}

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 10, Y: 10, Z: 25}
	lookAt := mango.Vector3{X: 0, Y: 0.5, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 20, 0, 10)

	samplesPerPixel := 100
	imageBuffer := mango.NewImageBuffer(1000, 1000, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.TileRenderProgressiveParallel(32)
	imageBuffer.Output("out.png")
}

func scene3() {
	world := make([]mango.Shape, 0)
	bvh := mango.BuildBVH(world, 0, len(world))

	lookFrom := mango.Vector3{X: 0, Y: 0, Z: 0}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 1}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 20, 0, 10)

	samplesPerPixel := 100
	imageBuffer := mango.NewImageBuffer(1000, 1000, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.TileRenderProgressiveParallel(32)
	imageBuffer.Output("out.png")
}

func scene4() {
	world := make([]mango.Shape, 0)
	triangles := mango.ParseOBJ("./suzanne.obj")

	for _, t := range triangles {
		world = append(world, t)
	}

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 0, Y: 0, Z: 10}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 20, 0, 10)

	samplesPerPixel := 100
	imageBuffer := mango.NewImageBuffer(500, 500, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.ScanlineRenderParallel()
	imageBuffer.Output("out.png")
}

func scene5() {
	world := make([]mango.Shape, 0)

	checkerTex := mango.NewCheckeredTextureCol(0.32, mango.RGB{R: 0.2, G: 0.3, B: 0.1}, mango.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := mango.Diffuse{Albedo: checkerTex}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1000, Z: 0}, 1000, ground_mat))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			center := mango.Vector3{X: float64(a) + 0.9*rand.Float64(), Y: 0.2, Z: float64(b) + 0.9*rand.Float64()}

			if mango.Length3(mango.Subtract3(center, mango.Vector3{X: 4, Y: 0.2, Z: 0})) > 0.9 {
				var sphere_mat mango.Material

				albedo := mango.Mul(
					mango.RGB{R: rand.Float64(), G: rand.Float64(), B: rand.Float64()},
					mango.RGB{R: rand.Float64(), G: rand.Float64(), B: rand.Float64()},
				)
				sphere_mat = mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(albedo)}
				world = append(world, mango.NewSphere(center, 0.2, sphere_mat))
			}
		}
	}

	mat2 := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.4, G: 0.2, B: 0.1})}
	world = append(world, mango.NewSphere(mango.Vector3{X: -4, Y: 1, Z: 0}, 1.0, mat2))

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 13, Y: 2, Z: 3}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1000.0/740.0, 20, 0, 10)

	samplesPerPixel := 10
	imageBuffer := mango.NewImageBuffer(1000, 740, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{mango.NewPointLight(mango.Vector3{X: 0, Y: 5, Z: 0}, mango.RGB{R: 50, G: 50, B: 50})},
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	//integrator.TileRenderProgressiveParallel(32)
	integrator.ScanlineRenderParallel()
	imageBuffer.Output("out.png")
}

func scene6() {
	world := make([]mango.Shape, 0)

	checkerTex := mango.NewCheckeredTextureCol(0.32, mango.RGB{R: 0.2, G: 0.3, B: 0.1}, mango.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := mango.Glossy{Albedo: checkerTex, Roughness: 0.9}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1001, Z: 0}, 1000, ground_mat))

	glossMat := mango.Glossy{
		Albedo:    mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness: 0.1,
	}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: 0, Z: 0}, 1, glossMat))

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 0, Y: 0, Z: 10}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 20, 0, 2)

	samplesPerPixel := 10
	imageBuffer := mango.NewImageBuffer(500, 500, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{mango.NewPointLight(mango.Vector3{X: -1, Y: 5, Z: 1}, mango.RGB{R: 50, G: 50, B: 50})},
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 10,
	}

	integrator.TileRenderProgressiveParallel(32)
	//integrator.ScanlineRenderParallel()
	//integrator.ScanlineRenderSlow()
	imageBuffer.Output("out.png")
}

func scene7() {
	world := make([]mango.Shape, 0)

	checkerTex := mango.NewCheckeredTextureCol(0.5, mango.RGB{R: 0.1, G: 0.1, B: 0.1}, mango.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := mango.Diffuse{Albedo: checkerTex}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1001, Z: 0}, 1000, ground_mat))

	glossMat1 := mango.Glossy{
		Albedo:    mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness: 0.1,
	}

	glossMat2 := mango.Glossy{
		Albedo:    mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness: 0.3,
	}

	glossMat3 := mango.Glossy{
		Albedo:    mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness: 0.5,
	}

	glossMat4 := mango.Glossy{
		Albedo:    mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness: 0.7,
	}

	glossMat5 := mango.Glossy{
		Albedo:    mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness: 0.9,
	}

	world = append(world, mango.NewSphere(mango.Vector3{X: -2, Y: 0, Z: 0}, 0.4, glossMat1))
	world = append(world, mango.NewSphere(mango.Vector3{X: -1, Y: 0, Z: 0}, 0.4, glossMat2))
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: 0, Z: 0}, 0.4, glossMat3))
	world = append(world, mango.NewSphere(mango.Vector3{X: 1, Y: 0, Z: 0}, 0.4, glossMat4))
	world = append(world, mango.NewSphere(mango.Vector3{X: 2, Y: 0, Z: 0}, 0.4, glossMat5))

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 0, Y: 10, Z: 10}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1000.0/400.0, 9, 0, 2)

	samplesPerPixel := 1000
	imageBuffer := mango.NewImageBuffer(1000, 400, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{mango.NewPointLight(mango.Vector3{X: -1, Y: 5, Z: 1}, mango.RGB{R: 50, G: 50, B: 50})},
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 100,
	}

	integrator.TileRenderProgressiveParallel(32)
	//integrator.ScanlineRenderParallel()
	//integrator.ScanlineRenderSlow()
	imageBuffer.Output("out.png")
}

func scene8() {
	world := make([]mango.Shape, 0)

	checkerTex := mango.NewCheckeredTextureCol(4, mango.RGB{R: 0.1, G: 0.1, B: 0.1}, mango.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := mango.Diffuse{Albedo: checkerTex}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1003, Z: 0}, 1000, ground_mat))

	triangles := mango.ParseOBJ("./lucy.obj")

	for _, t := range triangles {
		world = append(world, t)
	}

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 10, Y: 10, Z: 25}
	lookAt := mango.Vector3{X: 0, Y: 0.5, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 18, 0, 10)

	samplesPerPixel := 10
	imageBuffer := mango.NewImageBuffer(100, 100, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{mango.NewPointLight(mango.Vector3{X: 0, Y: 30, Z: 25}, mango.RGB{R: 10000, G: 10000, B: 10000})},
		Sampler:    &mango.UniformSampler{Spp: samplesPerPixel},
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.TileRenderProgressiveParallel(32)
	imageBuffer.Output("out.png")
}
