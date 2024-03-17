package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"os"
)

func copyImage(img *image.Image) error {
	// create destination file
	dstFile, err := os.Create("copy-test.jpeg")
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// create new image
	maxX := (*img).Bounds().Dx() / 4
	maxY := (*img).Bounds().Dy() / 4
	// newImg := image.NewRGBA((*img).Bounds())
	newImg := image.NewRGBA(image.Rect(0, 0, maxX, maxY))
	fmt.Printf("New image bounds: %v\n", newImg.Bounds())

	for y := range (*newImg).Bounds().Dy() {
		for x := range (*newImg).Bounds().Dx() {
			newImg.Set(x, y, (*img).At(x*4, y*4))
		}
	}

	err = jpeg.Encode(dstFile, newImg, nil)
	if err != nil {
		return err
	}
	return nil
}

// Remove the unused function createIconSize
func createIconSize(srcPath string) error {
	fmt.Printf("Creating icon size for %s\n", srcPath)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	img, err := jpeg.Decode(srcFile)
	if err != nil {
		slog.Error("Error decoding image")
		return err
	}

	fmt.Printf("Image bounds: %v\n", img.Bounds())
	fmt.Printf("Image color model: %s\n", img.ColorModel())
	fmt.Printf("Image width: %d, height: %d\n", img.Bounds().Dx(), img.Bounds().Dy())
	err = copyImage(&img)
	if err != nil {
		return err
	}
	return nil
}
