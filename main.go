package main

import (
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

const poolsize = 500

type job struct {
	frame *image.Paletted
	accum image.Image
}

func resizeGIF(im *gif.GIF, width, height int) error {
	t := time.Now()
	defer func() { fmt.Println(fmt.Sprint("elapsed: ", time.Since(t))) }()

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
		resized := resize.Resize(uint(width), uint(height), accum, resize.Lanczos3)

		jobs <- job{frame, resized}
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
	drawToFrame(j.frame, j.accum)
}

func drawToFrame(dst *image.Paletted, resized image.Image) {
	bounds := resized.Bounds()
	newPaletted := image.NewPaletted(bounds, dst.Palette)
	draw.Draw(newPaletted, bounds, resized, image.Point{}, draw.Src)

	*dst = *newPaletted
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	dimension := os.Args[2]
	splits := strings.Split(dimension, "x")
	width, _ := strconv.Atoi(splits[0])
	height, _ := strconv.Atoi(splits[1])

	src, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer src.Close()

	g, err := gif.DecodeAll(src)
	if err != nil {
		panic(err)
	}

	dst, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	err = resizeGIF(g, width, height)
	if err != nil {
		panic(err)
	}

	err = gif.EncodeAll(dst, g)
	if err != nil {
		panic(err)
	}
}
