package main

import (
	"fmt"
	"image/jpeg"
	"log/slog"
	"os"
)

// Remove the unused function createIconSize
func createIconSize(srcPath string) error {
	fmt.Printf("Creating icon size\n")
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
	fmt.Printf("Image color model: %v\n", img.ColorModel())
	fmt.Printf("Image width: %d, height: %d\n", img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}
