package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"os"
)

func writeNewImg(img *image.Image, scale int, filename string) error {
	// create destination file
	dstFile, err := os.Create(fmt.Sprintf("%s-%d.jpeg", filename, scale))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// create new image
	maxX := (*img).Bounds().Dx() / scale
	maxY := (*img).Bounds().Dy() / scale
	// newImg := image.NewRGBA((*img).Bounds())
	newImg := image.NewRGBA(image.Rect(0, 0, maxX, maxY))
	fmt.Printf("New image bounds: %v\n", newImg.Bounds())

	for y := range (*newImg).Bounds().Dy() {
		for x := range (*newImg).Bounds().Dx() {
			newImg.Set(x, y, (*img).At(x*scale, y*scale))
		}
	}

	err = jpeg.Encode(dstFile, newImg, nil)
	if err != nil {
		return err
	}
	return nil
}

func createAltSizes(srcPath string) {
	slog.Info(fmt.Sprintf("Creating alt sizes for %s", srcPath))
	srcFile, err := os.Open(srcPath)
	if err != nil {
		panic(err)
	}
	defer srcFile.Close()

	img, err := jpeg.Decode(srcFile)
	if err != nil {
		slog.Error("Error decoding image")
		panic(err)
	}

	filename := srcFile.Name()
	// Create medium img at 1/4 size
	err = writeNewImg(&img, 4, filename)
	if err != nil {
		panic(err)
	}
	// Create icon image at 1/10 size
	err = writeNewImg(&img, 10, filename)
	if err != nil {
		panic(err)
	}
}
