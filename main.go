package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

type job struct {
	frame         *image.Paletted
	accum         image.Image
	width, height int
}

func resizeGIF(im *gif.GIF, poolsize, width, height int) error {
	t := time.Now()
	defer func() { fmt.Println(fmt.Sprint("time spent processing image: ", time.Since(t))) }()

	jobs := make(chan job, poolsize)
	firstFrame := im.Image[0]
	firstFrameBounds := firstFrame.Bounds()
	accum := image.NewRGBA(firstFrameBounds)

	var wg sync.WaitGroup
	go manageWorkerPool(jobs, poolsize, worker, &wg)

	for _, frame := range im.Image {
		frameBounds := frame.Bounds()

		// this Draw call can not be in async part as we need to draw frames into
		// accumulating image in the original order and perform resizing
		// of accumulation result; this Draw call is much cheaper than call in async func
		// that is why we can parallelize the whole process
		draw.Draw(accum, frameBounds, frame, frameBounds.Min, draw.Over)

		// copy image to parallel resize task as well
		newPix := make([]uint8, len(accum.Pix))
		copy(newPix, accum.Pix)

		jobs <- job{frame, &image.RGBA{
			Pix:    newPix,
			Stride: accum.Stride,
			Rect:   accum.Rect,
		}, width, height}
	}

	close(jobs)
	wg.Wait()

	im.Config.Width = im.Image[0].Bounds().Dx()
	im.Config.Height = im.Image[0].Bounds().Dy()

	return nil
}

func manageWorkerPool(jobs <-chan job, limit int, worker func(j job), wg *sync.WaitGroup) {
	workerLimit := make(chan struct{}, limit)
	for j := range jobs {
		wg.Add(1)
		workerLimit <- struct{}{}
		go func(j job) {
			worker(j)
			<-workerLimit
			wg.Done()
		}(j)
	}
	wg.Wait()
}

func worker(j job) {
	drawToFrame(j.frame, j.accum, j.width, j.height)
}

func drawToFrame(dst *image.Paletted, accum image.Image, width, height int) {
	resized := resize.Resize(uint(width), uint(height), accum, resize.Bilinear)
	bounds := resized.Bounds()
	newPaletted := image.NewPaletted(bounds, dst.Palette)
	draw.Draw(newPaletted, bounds, resized, image.Point{}, draw.Src)

	*dst = *newPaletted
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	dim := flag.String("dims", "400x400", "Dimensions of the final image, p.e. 400x400")
	dimSplit := strings.Split(*dim, "x")
	width, _ := strconv.Atoi(dimSplit[0])
	height, _ := strconv.Atoi(dimSplit[1])

	srcPath := flag.String("src", "./src.gif", "Source image path")
	dstPath := flag.String("dst", "./out.gif", "Result path")

	poolsize := flag.Int("poolsize", 5000, "Number of frames processed in parallel")

	flag.Parse()

	src, err := os.Open(*srcPath)
	if err != nil {
		panic(err)
	}
	defer src.Close()

	g, err := gif.DecodeAll(src)
	if err != nil {
		panic(err)
	}

	dst, err := os.Create(*dstPath)
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	err = resizeGIF(g, *poolsize, width, height)
	if err != nil {
		panic(err)
	}

	err = gif.EncodeAll(dst, g)
	if err != nil {
		panic(err)
	}
}
