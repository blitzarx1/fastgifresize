package main

import (
	"image"
	"image/draw"
	"image/gif"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

func resizeGIF(im *gif.GIF, width, height int) error {
	img := image.NewRGBA(im.Image[0].Bounds())
	wg := sync.WaitGroup{}

	for _, frame := range im.Image {
		resized := resizeFrame(img, frame, width, height)
		wg.Add(1)
		go func(frame *image.Paletted, resized image.Image) {
			defer wg.Done()
			drawToFrame(frame, resized)
		}(frame, resized)
	}
	wg.Wait()

	im.Config.Width = im.Image[0].Bounds().Dx()
	im.Config.Height = im.Image[0].Bounds().Dy()

	return nil
}

func resizeFrame(accum *image.RGBA, frame *image.Paletted, width, height int) image.Image {
	frameBounds := frame.Bounds()
	draw.Draw(accum, frameBounds, frame, frameBounds.Min, draw.Over)

	return resize.Resize(uint(width), uint(height), accum, resize.Lanczos3)
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
