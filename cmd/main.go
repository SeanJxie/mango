package main

import (
	"math"
	"math/rand/v2"

	mango "github.com/SeanJxie/mango/src"
)

func main() {
	scene := 10
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
	case 9:
		scene9()
	case 10:
		scene10()
	}
}

func appendQuad(world []mango.Shape, a, b, c, d mango.Vector3, material mango.Material) []mango.Shape {
	return append(world,
		mango.NewTriangle(a, b, c, material),
		mango.NewTriangle(a, c, d, material),
	)
}

func appendEmissiveQuad(world []mango.Shape, a, b, c, d mango.Vector3, emittedRadiance mango.RGB) []mango.Shape {
	black := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.Black)}
	return append(world,
		mango.NewEmissiveTriangle(a, b, c, black, emittedRadiance),
		mango.NewEmissiveTriangle(a, c, d, black, emittedRadiance),
	)
}

func appendRotatedBox(world []mango.Shape, minPoint, maxPoint mango.Vector3, rotationY float64, material mango.Material) []mango.Shape {
	center := mango.Vector3{
		X: 0.5 * (minPoint.X + maxPoint.X),
		Y: 0.5 * (minPoint.Y + maxPoint.Y),
		Z: 0.5 * (minPoint.Z + maxPoint.Z),
	}

	cosTheta := math.Cos(rotationY)
	sinTheta := math.Sin(rotationY)
	return appendBoxWithTransform(world, minPoint, maxPoint, material, func(p mango.Vector3) mango.Vector3 {
		x := p.X - center.X
		z := p.Z - center.Z
		return mango.Vector3{
			X: center.X + x*cosTheta + z*sinTheta,
			Y: p.Y,
			Z: center.Z - x*sinTheta + z*cosTheta,
		}
	})
}

func appendBoxWithTransform(world []mango.Shape, minPoint, maxPoint mango.Vector3, material mango.Material, transform func(mango.Vector3) mango.Vector3) []mango.Shape {
	x0, y0, z0 := minPoint.X, minPoint.Y, minPoint.Z
	x1, y1, z1 := maxPoint.X, maxPoint.Y, maxPoint.Z

	p000 := transform(mango.Vector3{X: x0, Y: y0, Z: z0})
	p001 := transform(mango.Vector3{X: x0, Y: y0, Z: z1})
	p010 := transform(mango.Vector3{X: x0, Y: y1, Z: z0})
	p011 := transform(mango.Vector3{X: x0, Y: y1, Z: z1})
	p100 := transform(mango.Vector3{X: x1, Y: y0, Z: z0})
	p101 := transform(mango.Vector3{X: x1, Y: y0, Z: z1})
	p110 := transform(mango.Vector3{X: x1, Y: y1, Z: z0})
	p111 := transform(mango.Vector3{X: x1, Y: y1, Z: z1})

	world = appendQuad(world, p000, p100, p110, p010, material)
	world = appendQuad(world, p101, p001, p011, p111, material)
	world = appendQuad(world, p001, p000, p010, p011, material)
	world = appendQuad(world, p100, p101, p111, p110, material)
	world = appendQuad(world, p010, p110, p111, p011, material)
	world = appendQuad(world, p001, p101, p100, p000, material)

	return world
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
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 1),
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
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 2),
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
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 3),
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
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 4),
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	integrator.TileRenderProgressiveParallel(32)
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
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 5),
		Buffer:     imageBuffer,
		MaxBounces: 50,
	}

	//integrator.TileRenderProgressiveParallel(32)
	integrator.ScanlineRenderParallel()
	imageBuffer.Output("out.png")
}

func scene6() {
	world := make([]mango.Shape, 0)

	checkerTex := mango.NewCheckeredTextureCol(0.5, mango.RGB{R: 0.1, G: 0.1, B: 0.1}, mango.RGB{R: 0.9, G: 0.9, B: 0.9})
	ground_mat := mango.Diffuse{Albedo: checkerTex}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1001, Z: 0}, 1000, ground_mat))

	glossMat := mango.Glossy{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness:         0.1,
		ClearCoat:         1,
		IndexOfRefraction: 1.5,
	}
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: 0, Z: 0}, 1, glossMat))

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 0, Y: 5, Z: 10}
	lookAt := mango.Vector3{X: 0, Y: 0, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 20, 0, 2)

	samplesPerPixel := 10
	imageBuffer := mango.NewImageBuffer(500, 500, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:  bvh,
		Camera: camera,
		//Lights:     []mango.Light{mango.NewPointLight(mango.Vector3{X: -1, Y: 5, Z: 1}, mango.RGB{R: 500, G: 500, B: 500})},
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 6),
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

	glossMat1 := mango.Metal{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Absorption:        mango.RGB{R: 3.42, G: 2.45, B: 1.91},
		IndexOfRefraction: mango.RGB{R: 0.18, G: 0.43, B: 1.38},
		Roughness:         0.1,
	}

	glossMat2 := mango.Metal{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Absorption:        mango.RGB{R: 3.42, G: 2.45, B: 1.91},
		IndexOfRefraction: mango.RGB{R: 0.18, G: 0.43, B: 1.38},
		Roughness:         0.2,
	}

	glossMat3 := mango.Metal{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Absorption:        mango.RGB{R: 3.42, G: 2.45, B: 1.91},
		IndexOfRefraction: mango.RGB{R: 0.18, G: 0.43, B: 1.38},
		Roughness:         0.3,
	}

	glossMat4 := mango.Metal{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Absorption:        mango.RGB{R: 3.42, G: 2.45, B: 1.91},
		IndexOfRefraction: mango.RGB{R: 0.18, G: 0.43, B: 1.38},
		Roughness:         0.4,
	}

	glossMat5 := mango.Metal{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Absorption:        mango.RGB{R: 3.42, G: 2.45, B: 1.91},
		IndexOfRefraction: mango.RGB{R: 0.18, G: 0.43, B: 1.38},
		Roughness:         0.9,
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
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 7),
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
	world = append(world, mango.NewSphere(mango.Vector3{X: 0, Y: -1003.6, Z: 0}, 1000, ground_mat))

	triangles := mango.ParseOBJ("./lucy.obj")

	for _, t := range triangles {
		world = append(world, t)
	}

	bvh := mango.BuildBVH(world, 0, len(world))

	// Set up camera
	lookFrom := mango.Vector3{X: 10, Y: 10, Z: 25}
	lookAt := mango.Vector3{X: 0, Y: 0.5, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 18, 0, 10)

	samplesPerPixel := 5
	imageBuffer := mango.NewImageBuffer(100, 100, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{mango.NewPointLight(mango.Vector3{X: 0, Y: 50, Z: 30}, mango.RGB{R: 10000, G: 10000, B: 10000})},
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 8),
		Buffer:     imageBuffer,
		MaxBounces: 1000,
	}

	integrator.TileRenderProgressiveParallel(32)
	imageBuffer.Output("out.png")
}

func scene9() {
	world := make([]mango.Shape, 0)

	white := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.78, G: 0.78, B: 0.74})}
	red := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.65, G: 0.08, B: 0.05})}
	green := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.12, G: 0.45, B: 0.15})}
	diffuseBlock := mango.Glossy{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 1, G: 1, B: 1}),
		Roughness:         0.001,
		ClearCoat:         1,
		IndexOfRefraction: 2.2,
	}
	// microfacetMetal := mango.Metal{
	// 	Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.96, G: 0.96, B: 0.94}),
	// 	Absorption:        mango.RGB{R: 4.10, G: 3.10, B: 2.30},
	// 	IndexOfRefraction: mango.RGB{R: 0.16, G: 0.15, B: 0.14},
	// 	Roughness:         0.1,
	// }

	// Cornell box, open toward +Z.
	world = appendQuad(world,
		mango.Vector3{X: -1, Y: 0, Z: 1},
		mango.Vector3{X: 1, Y: 0, Z: 1},
		mango.Vector3{X: 1, Y: 0, Z: -1},
		mango.Vector3{X: -1, Y: 0, Z: -1},
		white,
	)
	world = appendQuad(world,
		mango.Vector3{X: -1, Y: 2, Z: -1},
		mango.Vector3{X: 1, Y: 2, Z: -1},
		mango.Vector3{X: 1, Y: 2, Z: 1},
		mango.Vector3{X: -1, Y: 2, Z: 1},
		white,
	)
	world = appendQuad(world,
		mango.Vector3{X: -1, Y: 0, Z: -1},
		mango.Vector3{X: 1, Y: 0, Z: -1},
		mango.Vector3{X: 1, Y: 2, Z: -1},
		mango.Vector3{X: -1, Y: 2, Z: -1},
		white,
	)
	world = appendQuad(world,
		mango.Vector3{X: -1, Y: 0, Z: -1},
		mango.Vector3{X: -1, Y: 0, Z: 1},
		mango.Vector3{X: -1, Y: 2, Z: 1},
		mango.Vector3{X: -1, Y: 2, Z: -1},
		red,
	)
	world = appendQuad(world,
		mango.Vector3{X: 1, Y: 0, Z: 1},
		mango.Vector3{X: 1, Y: 0, Z: -1},
		mango.Vector3{X: 1, Y: 2, Z: -1},
		mango.Vector3{X: 1, Y: 2, Z: 1},
		green,
	)

	world = appendRotatedBox(world,
		mango.Vector3{X: -0.72, Y: 0, Z: -0.48},
		mango.Vector3{X: -0.18, Y: 1.3, Z: 0.08},
		18*mango.Deg2Rad,
		diffuseBlock,
	)
	world = append(world, mango.NewSphere(mango.Vector3{X: 0.42, Y: 0.36, Z: -0.28}, 0.36, diffuseBlock))

	lightRadiance := mango.RGB{R: 18 * 0.5, G: 17 * 0.5, B: 15 * 0.5}
	lightV0 := mango.Vector3{X: -0.34, Y: 2, Z: -0.26}
	lightV1 := mango.Vector3{X: 0.34, Y: 2, Z: -0.26}
	lightV2 := mango.Vector3{X: 0.34, Y: 2, Z: 0.26}
	lightV3 := mango.Vector3{X: -0.34, Y: 2, Z: 0.26}
	world = appendEmissiveQuad(world, lightV0, lightV1, lightV2, lightV3, lightRadiance)
	areaLight := mango.NewDiffuseAreaLight(lightV0, lightV1, lightV2, lightV3, lightRadiance)

	bvh := mango.BuildBVH(world, 0, len(world))

	lookFrom := mango.Vector3{X: 0, Y: 1, Z: 4}
	lookAt := mango.Vector3{X: 0, Y: 1, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 33, 0, 4)

	samplesPerPixel := 1000
	imageBuffer := mango.NewImageBuffer(700, 700, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{areaLight},
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 9),
		Buffer:     imageBuffer,
		MaxBounces: 24,
	}

	integrator.TileRenderProgressiveParallel(32)
	imageBuffer.Output("out.png")
}

func scene10() {
	world := make([]mango.Shape, 0)

	wall := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.72, G: 0.72, B: 0.68})}
	floor := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.5, G: 0.52, B: 0.48})}
	leftWall := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.58, G: 0.13, B: 0.1})}
	rightWall := mango.Diffuse{Albedo: mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.12, G: 0.36, B: 0.18})}
	mirrorMetal := mango.Metal{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.98, G: 0.98, B: 0.96}),
		Absorption:        mango.RGB{R: 4.8, G: 3.1, B: 2.1},
		IndexOfRefraction: mango.RGB{R: 0.16, G: 0.14, B: 0.13},
		Roughness:         0.015,
	}
	lucyGlossy := mango.Glossy{
		Albedo:            mango.NewSolidColourTextureAlbedo(mango.RGB{R: 0.84, G: 0.8, B: 0.72}),
		Roughness:         0.015,
		ClearCoat:         1,
		IndexOfRefraction: 2.0,
	}

	world = appendQuad(world,
		mango.Vector3{X: -2.2, Y: 0.01, Z: 4.8},
		mango.Vector3{X: 2.2, Y: 0.01, Z: 4.8},
		mango.Vector3{X: 2.2, Y: 0.01, Z: -1.6},
		mango.Vector3{X: -2.2, Y: 0.01, Z: -1.6},
		floor,
	)
	world = appendQuad(world,
		mango.Vector3{X: -2.2, Y: -0.02, Z: -1.6},
		mango.Vector3{X: 2.2, Y: -0.02, Z: -1.6},
		mango.Vector3{X: 2.2, Y: 2.7, Z: -1.6},
		mango.Vector3{X: -2.2, Y: 2.7, Z: -1.6},
		wall,
	)
	world = appendQuad(world,
		mango.Vector3{X: -2.2, Y: -0.02, Z: 4.8},
		mango.Vector3{X: -2.2, Y: -0.02, Z: -1.6},
		mango.Vector3{X: -2.2, Y: 2.7, Z: -1.6},
		mango.Vector3{X: -2.2, Y: 2.7, Z: 4.8},
		leftWall,
	)
	world = appendQuad(world,
		mango.Vector3{X: 2.2, Y: -0.02, Z: -1.6},
		mango.Vector3{X: 2.2, Y: -0.02, Z: 4.8},
		mango.Vector3{X: 2.2, Y: 2.7, Z: 4.8},
		mango.Vector3{X: 2.2, Y: 2.7, Z: -1.6},
		rightWall,
	)
	world = appendQuad(world,
		mango.Vector3{X: -2.2, Y: 2.7, Z: -1.6},
		mango.Vector3{X: 2.2, Y: 2.7, Z: -1.6},
		mango.Vector3{X: 2.2, Y: 2.7, Z: 4.8},
		mango.Vector3{X: -2.2, Y: 2.7, Z: 4.8},
		wall,
	)
	world = appendQuad(world,
		mango.Vector3{X: 2.2, Y: -0.02, Z: 4.8},
		mango.Vector3{X: -2.2, Y: -0.02, Z: 4.8},
		mango.Vector3{X: -2.2, Y: 2.7, Z: 4.8},
		mango.Vector3{X: 2.2, Y: 2.7, Z: 4.8},
		wall,
	)

	leftLightRadiance := mango.RGB{R: 3.5, G: 7.5, B: 18}
	leftLightV0 := mango.Vector3{X: -2.19, Y: 0.7, Z: 0.7}
	leftLightV1 := mango.Vector3{X: -2.19, Y: 0.7, Z: -0.7}
	leftLightV2 := mango.Vector3{X: -2.19, Y: 2.1, Z: -0.7}
	leftLightV3 := mango.Vector3{X: -2.19, Y: 2.1, Z: 0.7}
	world = appendEmissiveQuad(world, leftLightV0, leftLightV1, leftLightV2, leftLightV3, leftLightRadiance)
	leftAreaLight := mango.NewDiffuseAreaLight(leftLightV0, leftLightV1, leftLightV2, leftLightV3, leftLightRadiance)

	rightLightRadiance := mango.RGB{R: 18, G: 6.5, B: 3}
	rightLightV0 := mango.Vector3{X: 2.19, Y: 0.7, Z: -0.7}
	rightLightV1 := mango.Vector3{X: 2.19, Y: 0.7, Z: 0.7}
	rightLightV2 := mango.Vector3{X: 2.19, Y: 2.1, Z: 0.7}
	rightLightV3 := mango.Vector3{X: 2.19, Y: 2.1, Z: -0.7}
	world = appendEmissiveQuad(world, rightLightV0, rightLightV1, rightLightV2, rightLightV3, rightLightRadiance)
	rightAreaLight := mango.NewDiffuseAreaLight(rightLightV0, rightLightV1, rightLightV2, rightLightV3, rightLightRadiance)

	for _, center := range []mango.Vector3{
		{X: -0.68, Y: 0.55, Z: -1.18},
		{X: 0.68, Y: 0.55, Z: -1.18},
		{X: -0.68, Y: 1.15, Z: -1.18},
		{X: 0.68, Y: 1.15, Z: -1.18},
	} {
		world = append(world, mango.NewSphere(center, 0.22, mirrorMetal))
	}

	lucyTriangles := loadLucyTriangles(lucyGlossy)
	if len(lucyTriangles) == 0 {
		panic("failed to load lucy.obj from cmd/lucy.obj or ./lucy.obj")
	}
	for _, triangle := range lucyTriangles {
		world = append(world, triangle)
	}

	bvh := mango.BuildBVH(world, 0, len(world))

	lookFrom := mango.Vector3{X: 0, Y: 1.05, Z: 3.6}
	lookAt := mango.Vector3{X: 0, Y: 0.9, Z: 0}
	camera := mango.NewPerspectiveCamera(lookFrom, lookAt, 1, 32, 0, mango.Length3(mango.Subtract3(lookFrom, lookAt)))

	samplesPerPixel := 1000
	imageBuffer := mango.NewImageBuffer(1000, 1000, samplesPerPixel)

	integrator := mango.PathIntegrator{
		World:      bvh,
		Camera:     camera,
		Lights:     []mango.Light{leftAreaLight, rightAreaLight},
		Sampler:    mango.NewStratifiedSampler(samplesPerPixel, 10),
		Buffer:     imageBuffer,
		MaxBounces: 100,
	}

	integrator.TileRenderProgressiveParallel(32)
	imageBuffer.Output("out.png")
}

func loadLucyTriangles(material mango.Material) []*mango.Triangle {
	transform := func(p mango.Vector3) mango.Vector3 {
		const (
			minZ    = -605.890015
			centerX = 690.9680175
			centerY = -121.5364915
			scale   = 1.8 / 1597.221985
		)

		return mango.Vector3{
			X: (p.X - centerX) * scale,
			Y: (p.Z - minZ) * scale,
			Z: (p.Y - centerY) * scale,
		}
	}

	triangles := mango.ParseOBJWithMaterialTransform("cmd/lucy.obj", material, transform)
	if len(triangles) == 0 {
		triangles = mango.ParseOBJWithMaterialTransform("./lucy.obj", material, transform)
	}
	return triangles
}
