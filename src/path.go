package mango

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Lighting utilities.

const (
	russianRouletteStartBounce = 3
	minSurvivalProbability     = 0.05
	maxPathThroughput          = 10.0
	maxSampleRadiance          = 10.0
)

// https://pbr-book.org/3ed-2018/Light_Transport_I_Surface_Reflection/Direct_Lighting#EstimateDirect
func DirectLighting(light Light, lightSample, scatterSample Vector2, woLocal Vector3, intersection *ShapeIntersection, world Shape) RGB {
	Ld := Black
	Li, wiWorld, lightPdf, visibilityTarget := light.SampleLi(intersection, lightSample)

	// Light source multiple important sampling.
	if lightPdf <= 0 || math.IsNaN(lightPdf) || math.IsInf(lightPdf, 0) || IsBlack(Li) || !IsValidRGB(Li) {
		return Ld
	}

	cosTheta := Dot3(wiWorld, intersection.SurfaceNormal)
	if cosTheta <= 0 || math.IsNaN(cosTheta) || math.IsInf(cosTheta, 0) {
		return Ld
	}

	bsdf := intersection.GetBSDF()

	localBasis := NewLocalBasis(intersection.SurfaceNormal)
	wiLocal := StandardToLocalBasis(wiWorld, localBasis)

	f := bsdf.Value(woLocal, wiLocal)
	f = Scale(f, cosTheta)
	if IsBlack(f) || !IsValidRGB(f) {
		return Ld
	}

	scatterPdf := bsdf.Pdf(woLocal, wiLocal)
	if scatterPdf < 0 || math.IsNaN(scatterPdf) || math.IsInf(scatterPdf, 0) {
		scatterPdf = 0
	}

	if !Visible(intersection.Point, visibilityTarget, world) {
		return Ld
	}

	if light.GetType() == Point {
		return Mul(f, Scale(Li, 1/lightPdf))
	}

	weight := PowerHeuristic(1, lightPdf, 1, scatterPdf)
	Ld = Mul(f, Scale(Li, weight/lightPdf))
	if !IsValidRGB(Ld) {
		return Black
	}

	// BSDF multiple importance sampling (implement for non delta lights)

	return Ld
}

func SampleOneLight(lights []Light, woLocal Vector3, intersection *ShapeIntersection, world Shape, sampler Sampler) RGB {
	// Choose a random light in the scene.
	numLights := len(lights)
	if numLights == 0 {
		return Black
	}

	lightIndex := min(int(sampler.Sample1D()*float64(numLights)), numLights-1)
	chosenLight := lights[lightIndex]

	lightSample := sampler.Sample2D()
	scatterSample := sampler.Sample2D()

	return Scale(DirectLighting(chosenLight, lightSample, scatterSample, woLocal, intersection, world), float64(numLights))
}

type PathIntegrator struct {
	World      Shape
	Lights     []Light
	Camera     *PerspectiveCamera
	Sampler    Sampler
	Buffer     *ImageBuffer
	MaxBounces int
}

func (integ *PathIntegrator) Li(ray Ray, sampler Sampler) RGB {
	L := Black                    // Incoming radiance.
	beta := White                 // Contribution weight.
	emissionVisibleBounce := true // Camera rays should see emissive geometry.

	var foundIntersection bool
	var intersection *ShapeIntersection

	for b := 0; b < integ.MaxBounces; b++ {

		foundIntersection, intersection = integ.World.Intersect(&ray, Epsilon, math.Inf(0))
		if !foundIntersection {
			// Missed the scene, hit skybox.
			//L = Add(L, Mul(SkyBox(ray.Direction), beta))
			break
		}

		// Found an intersection. The goal now is to pretend the ray
		// is bouncing on its way towards the camera and ask how much
		// light is coming from this direction.

		if !IsBlack(intersection.Le) {
			if emissionVisibleBounce {
				L = Add(L, Mul(beta, intersection.Le))
			}
			break
		}

		bsdf := intersection.GetBSDF()

		// BSDFs work in intersection-local coordinates (where intersection normal is "upwards" z-axis).
		woWorld := ScalarMultiply3(ray.Direction, -1)
		localBasis := NewLocalBasis(intersection.SurfaceNormal)
		woLocal := StandardToLocalBasis(woWorld, localBasis)

		// Direct lighting for all BSDFs that are not purely specular (only SpecularBxDF)
		//
		// For perfect mirrors (pure specular), the probability that there is a light ray that reflects
		// to exactly our outgoing ray is zero.
		if len(integ.Lights) > 0 && bsdf.GetType()&SpecularBxDF == 0 {
			directLight := Mul(beta, SampleOneLight(integ.Lights, woLocal, intersection, integ.World, sampler))
			L = Add(L, directLight)
		}

		// Get the radiance value carried by the outgoing ray, and a random incoming ray that contributed
		// to that radiance, along with the probability that it would do so.
		f, wiLocal, pdf, sampledType := bsdf.SampleAndValueWithType(woLocal, sampler.Sample2D())
		if pdf <= 0 || math.IsNaN(pdf) || math.IsInf(pdf, 0) {
			break
		}

		wiWorld := LocalToStandardBasis(wiLocal, localBasis)

		cosTheta := Dot3(wiWorld, intersection.SurfaceNormal)
		if cosTheta <= 0 || math.IsNaN(cosTheta) || math.IsInf(cosTheta, 0) {
			break
		}
		f = Scale(f, cosTheta/pdf)
		if !IsValidRGB(f) {
			break
		}

		beta = Mul(beta, f)

		if IsBlack(beta) || !IsValidRGB(beta) {
			break
		}

		beta = ClampMaxComponent(beta, maxPathThroughput)

		// Russian roulette: keep low-contribution paths from running forever without
		// giving rare survivors enormous weights.
		if b >= russianRouletteStartBounce {
			survivalProbability := Clamp(MaxComponent(beta), minSurvivalProbability, 1)
			if sampler.Sample1D() >= survivalProbability {
				break
			}

			beta = Scale(beta, 1/survivalProbability)
		}

		emissionVisibleBounce = sampledType&(SpecularBxDF|GlossBxDF) != 0
		ray = intersection.CastRay(wiWorld)
	}

	if !IsValidRGB(L) {
		return Black
	}
	return ClampMaxComponent(L, maxSampleRadiance)
}

func (integ *PathIntegrator) ScanlineRenderSlow() {

	imageBuffer := integ.Buffer
	width := integ.Buffer.Width
	height := integ.Buffer.Height
	camera := integ.Camera
	sampler := integ.Sampler
	samplesPerPixel := sampler.SamplesPerPixel()

	start := time.Now()
	preview := startPreview(imageBuffer, "Scanline render")
	preview.Update(samplesPerPixel, 0, fmt.Sprintf("0/%d rows", height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixelColour := Black
			for s := 0; s < samplesPerPixel; s++ {
				sampler.StartPixelSample(x, y, s)
				pixelSample := sampler.Sample2D()
				u := (float64(x) + pixelSample.X) / float64(width)
				v := (float64(height-(y+1)) + pixelSample.Y) / float64(height)

				camRay := camera.CastRay(u, v, sampler)
				pixelColour = Add(pixelColour, integ.Li(*camRay, sampler))
			}

			imageBuffer.AddSample(x, y, pixelColour)
		}

		fmt.Printf("\rRender progress: %.2f%% (%d/%d rows)",
			100.0*float64(y+1)/float64(height), y+1, height)
		preview.Update(samplesPerPixel, float64(y+1)/float64(height), fmt.Sprintf("%d/%d rows", y+1, height))
	}

	dt := time.Since(start)
	fmt.Printf("\nRender time: %v\n", dt)
}

func (integ *PathIntegrator) ScanlineRenderParallel() {

	imageBuffer := integ.Buffer
	width := integ.Buffer.Width
	height := integ.Buffer.Height
	camera := integ.Camera
	sampler := integ.Sampler
	samplesPerPixel := sampler.SamplesPerPixel()

	var wg sync.WaitGroup
	var completedRows int64

	start := time.Now()
	preview := startPreview(imageBuffer, "Parallel scanline render")
	preview.Update(samplesPerPixel, 0, fmt.Sprintf("0/%d rows", height))

	for y := 0; y < height; y++ {
		wg.Add(1)

		go func(y int) {
			defer wg.Done()

			localSampler := integ.Sampler.Clone(0)

			for x := 0; x < width; x++ {
				pixelColour := Black
				for s := 0; s < samplesPerPixel; s++ {
					localSampler.StartPixelSample(x, y, s)
					pixelSample := localSampler.Sample2D()
					u := (float64(x) + pixelSample.X) / float64(width)
					v := (float64(height-(y+1)) + pixelSample.Y) / float64(height)

					camRay := camera.CastRay(u, v, localSampler)
					pixelColour = Add(pixelColour, integ.Li(*camRay, localSampler))
				}

				imageBuffer.AddSample(x, y, pixelColour)
			}

			count := atomic.AddInt64(&completedRows, 1)

			fmt.Printf("\rRender progress: %.2f%% (%d/%d rows)",
				100.0*float64(count)/float64(height), count, height)
			preview.Update(samplesPerPixel, float64(count)/float64(height), fmt.Sprintf("%d/%d rows", count, height))
		}(y)
	}

	// Wait for all rows to finish computing
	wg.Wait()

	dt := time.Since(start)
	fmt.Printf("\nRender time: %v\n", dt)
}

type Tile struct {
	Xmin, Xmax, Ymin, Ymax int
}

func MakeTiles(width, height, tileSize int) []Tile {
	tiles := make([]Tile, 0)
	for y := 0; y < height; y += tileSize {
		for x := 0; x < width; x += tileSize {
			tile := Tile{
				Xmin: x,
				Ymin: y,
				Xmax: min(x+tileSize, width),
				Ymax: min(y+tileSize, height),
			}
			tiles = append(tiles, tile)
		}
	}

	// rand.Shuffle(len(tiles), func(i, j int) {
	// 	tiles[i], tiles[j] = tiles[j], tiles[i]
	// })

	return tiles
}

func (integ *PathIntegrator) renderTile(t Tile, sampleIndex int, sampler Sampler) {
	imageBuffer := integ.Buffer
	camera := integ.Camera

	width := imageBuffer.Width
	height := imageBuffer.Height

	for y := t.Ymin; y < t.Ymax; y++ {
		for x := t.Xmin; x < t.Xmax; x++ {

			sampler.StartPixelSample(x, y, sampleIndex)
			pixelSample := sampler.Sample2D()
			u := (float64(x) + pixelSample.X) / float64(width)
			v := (float64(height-(y+1)) + pixelSample.Y) / float64(height)

			camRay := camera.CastRay(u, v, sampler)
			sampleColour := integ.Li(*camRay, sampler)

			imageBuffer.AddSample(x, y, sampleColour)
		}
	}
}

func (integ *PathIntegrator) TileRenderProgressiveParallel(tileSize int) {
	// Build tile group.
	totalSamples := integ.Sampler.SamplesPerPixel()
	tiles := MakeTiles(integ.Buffer.Width, integ.Buffer.Height, tileSize)
	numWorkers := runtime.NumCPU() // Use all possible logical CPUs.

	start := time.Now()
	preview := startPreview(integ.Buffer, "Tile wave render")
	preview.Update(1, 0, fmt.Sprintf("0/%d waves", totalSamples))

	for wave := 0; wave < totalSamples; wave++ {

		tileChan := make(chan Tile, len(tiles))
		for _, t := range tiles {
			tileChan <- t
		}
		close(tileChan)

		var wg sync.WaitGroup
		var completedTiles int64
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				localSampler := integ.Sampler.Clone(0)
				for tile := range tileChan {
					integ.renderTile(tile, wave, localSampler)
					count := atomic.AddInt64(&completedTiles, 1)
					progress := (float64(wave) + float64(count)/float64(len(tiles))) / float64(totalSamples)
					preview.Update(wave+1, progress, fmt.Sprintf("Wave %d/%d, tile %d/%d", wave+1, totalSamples, count, len(tiles)))
				}
			}()
		}

		wg.Wait()
		fmt.Printf("Finished wave %d / %d\n", wave+1, totalSamples)
		preview.Update(wave+1, float64(wave+1)/float64(totalSamples), fmt.Sprintf("%d/%d waves", wave+1, totalSamples))
	}

	dt := time.Since(start)
	fmt.Printf("\nRender time: %v\n", dt)
	preview.Update(totalSamples, 1, "Complete")
}

func startPreview(buffer *ImageBuffer, mode string) *RenderPreview {
	preview, err := StartRenderPreview(buffer, mode)
	if err != nil {
		fmt.Printf("Render preview unavailable: %v\n", err)
		return nil
	}

	fmt.Printf("Render preview: %s\n", preview.URL())
	if err := preview.Open(); err != nil {
		fmt.Printf("Render preview could not open automatically: %v\n", err)
	}
	return preview
}
