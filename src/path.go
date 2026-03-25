package mango

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type PathIntegrator struct {
	World      Shape
	Camera     *PerspectiveCamera
	Sampler    Sampler
	Buffer     *ImageBuffer
	MaxBounces int
}

func (integ *PathIntegrator) Li(ray Ray) RGB {
	L := Black    // Incoming radiance.
	beta := White // Contribution weight.

	var foundIntersection bool
	var intersection *ShapeIntersection

	for b := 0; b < integ.MaxBounces; b++ {

		foundIntersection, intersection = integ.World.Intersect(&ray, Epsilon, math.Inf(0))
		if !foundIntersection {
			// Missed the scene, hit skybox.
			L = Add(L, ScaleComponentsRGB(SkyBox(ray.Direction), beta))
			break
		}

		bsdf := intersection.GetBSDF()

		// Work in intersection-local coordinates.
		localBasis := NewLocalBasis(intersection.SurfaceNormal)
		wo := ScalarMultiply3(ray.Direction, -1)
		woLocal := StandardToLocalBasis(wo, localBasis)
		f, wiLocal, pdf := bsdf.SampleF(woLocal, integ.Sampler.Sample2D())
		wiWorld := LocalToStandardBasis(wiLocal, localBasis)

		tmp := ScaleRGB(f, math.Abs(Dot3(wiWorld, intersection.SurfaceNormal))/pdf)
		beta = ScaleComponentsRGB(beta, tmp)

		ray = intersection.CastRay(wiWorld)
	}

	return L
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

func (integ *PathIntegrator) renderTile(t Tile) {
	sampler := integ.Sampler
	imageBuffer := integ.Buffer
	camera := integ.Camera

	width := imageBuffer.Width
	height := imageBuffer.Height

	for y := t.Ymin; y < t.Ymax; y++ {
		for x := t.Xmin; x < t.Xmax; x++ {

			u := (float64(x) + sampler.Sample1D()) / float64(width)
			v := (float64(height-(y+1)) + sampler.Sample1D()) / float64(height)

			camRay := camera.CastRay(u, v, sampler)
			sampleColour := integ.Li(*camRay)

			imageBuffer.AddSample(x, y, sampleColour)
		}
	}
}

func (integ *PathIntegrator) TileRenderProgressiveParallel(tileSize int) {
	// Set up SDL for UI preview.
	// var renderer *sdl.Renderer
	// var previewTexture *sdl.Texture
	// sdl.Init(uint32(sdl.INIT_EVERYTHING))
	// defer sdl.Quit()

	// sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "0") // "0" uses nearest-neighbour

	// window, _ := sdl.CreateWindow(
	// 	fmt.Sprintf("Render Preview (%d x %d)", integ.Buffer.Width, integ.Buffer.Height),
	// 	int32(sdl.WINDOWPOS_CENTERED),
	// 	int32(sdl.WINDOWPOS_CENTERED),
	// 	int32(integ.Buffer.Width/2), int32(integ.Buffer.Height/2),
	// 	uint32(sdl.WINDOW_RESIZABLE),
	// )
	// defer window.Destroy()

	// renderer, _ = sdl.CreateRenderer(window, -1, uint32(sdl.RENDERER_ACCELERATED))
	// defer renderer.Destroy()

	// renderer.SetLogicalSize(int32(integ.Buffer.Width), int32(integ.Buffer.Height)) // Maintain aspect ratio

	// previewTexture, _ = renderer.CreateTexture(
	// 	uint32(sdl.PIXELFORMAT_ABGR8888),
	// 	int(sdl.TEXTUREACCESS_STREAMING),
	// 	int32(integ.Buffer.Width), int32(integ.Buffer.Height),
	// )
	// defer previewTexture.Destroy()

	// Build tile group.
	totalSamples := integ.Sampler.SamplesPerPixel()
	tiles := MakeTiles(integ.Buffer.Width, integ.Buffer.Height, tileSize)
	numWorkers := runtime.NumCPU() // Use all possible logical CPUs.

	previewReady := make(chan int, 1)

	start := time.Now()

	for wave := 0; wave < totalSamples; wave++ {

		tileChan := make(chan Tile, len(tiles))
		for _, t := range tiles {
			tileChan <- t
		}
		close(tileChan)

		var wg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for tile := range tileChan {
					integ.renderTile(tile)
				}
			}()
		}

		wg.Wait()
		log.Printf("Finished wave %d / %d", wave+1, totalSamples)

		// Signal the UI thread that we have new data
		select {
		case previewReady <- wave:
		default:
		}
	}

	dt := time.Since(start)
	fmt.Printf("\nRender time: %v\n", dt)

	// for {
	// 	// Handle SDL Events (Prevents "Not Responding" status)
	// 	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
	// 		switch event.(type) {
	// 		case *sdl.QuitEvent:
	// 			integ.Buffer.Output("out.png")
	// 			return
	// 		}
	// 	}

	// 	select {
	// 	case currentWave := <-previewReady:
	// 		tmpBuffer8bit := make([]uint8, len(integ.Buffer.Pixels))
	// 		sampleScale := 1.0 / float64(currentWave+1)
	// 		gamma := 2.2

	// 		for i := 0; i < len(tmpBuffer8bit); i++ {
	// 			if (i+1)%4 == 0 { // Alpha channel.
	// 				tmpBuffer8bit[i] = 255
	// 				continue
	// 			}
	// 			c := integ.Buffer.Pixels[i] * sampleScale
	// 			c = math.Pow(c, 1/gamma)
	// 			tmpBuffer8bit[i] = uint8(255 * Clamp(c, 0, 0.999))
	// 		}
	// 		previewTexture.Update(nil, unsafe.Pointer(&tmpBuffer8bit[0]), integ.Buffer.Width*4)

	// 	default:
	// 		// Just keep the window alive
	// 	}

	// 	renderer.Clear()
	// 	renderer.Copy(previewTexture, nil, nil)
	// 	renderer.Present()
	// }
}
