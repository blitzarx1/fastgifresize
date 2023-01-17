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
	"time"

	"github.com/nfnt/resize"
)

func resizeGIF(im *gif.GIF, width, height int) error {
	start := time.Now()
	defer func() { fmt.Println(fmt.Sprint("elapsed total: ", time.Since(start))) }()

	img := image.NewRGBA(im.Image[0].Bounds())
	// var img *image.RGBA
	for i, frame := range im.Image {
		fmt.Println(im.Disposal[i])
		frameBounds := frame.Bounds()

		draw.Draw(img, frameBounds, frame, frameBounds.Min, draw.Over)

		var resized image.Image = img
		if width != 0 {
			resized = resize.Resize(uint(width), 0, img, resize.Lanczos3)
		}

		if height != 0 {
			resized = resize.Resize(0, uint(height), img, resize.Lanczos3)
		}

		bounds := resized.Bounds()
		// if img == nil {
		// 	img = image.NewRGBA(bounds)
		// }
		// draw.Draw(img, bounds, resized, bounds.Min, draw.Over)

		newPaletted := image.NewPaletted(bounds, frame.Palette)
		draw.Draw(newPaletted, bounds, resized, image.Point{}, draw.Src)

		*frame = *newPaletted
	}

	im.Config.Width = im.Image[0].Bounds().Dx()
	im.Config.Height = im.Image[0].Bounds().Dy()

	return nil
}

func resizeGIFFrameFixAspect(accum *image.RGBA, frame *image.Paletted, width, height int) image.Image {
	// start := time.Now()
	// defer func() { fmt.Println(fmt.Sprint("elapsed for frame resize: ", time.Since(start))) }()

	frameBounds := frame.Bounds()
	draw.Draw(accum, frameBounds, frame, frameBounds.Min, draw.Over)

	var resized image.Image = accum
	if width != 0 {
		resized = resize.Resize(uint(width), 0, accum, resize.Lanczos3)
	}

	if height != 0 {
		resized = resize.Resize(0, uint(height), accum, resize.Lanczos3)
	}

	return resized
}

func drawToFrame(dst *image.Paletted, resized image.Image) {
	// start := time.Now()
	// defer func() { fmt.Println(fmt.Sprint("elapsed for drawing to frame: ", time.Since(start))) }()

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
