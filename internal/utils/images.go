package utils

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

const (
	WEBP_DEFAULT_QUALITY = 75.0
	WEBP_SMALL_QUALITY = 50.0
)

type ConvertResizeParams struct {
	Width    uint
	Height   uint
	Quality  float32
	Lossless bool
}

func NewConvertResizeParams() ConvertResizeParams {
	return ConvertResizeParams{
		Width:    0,
		Height:   0,
		Quality:  WEBP_DEFAULT_QUALITY,
		Lossless: true,
	}
}

func ConvertAndResizeImage(src, dst string, p ConvertResizeParams) error {
	inputFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		return err
	}

	if p.Width != 0 || p.Height != 0 {
		fmt.Println("Resizing - Width:", p.Width, "Height:", p.Height)
		// If one value is 0, resize.Resize maintains aspect ratio
		img = resize.Resize(p.Width, p.Height, img, resize.Lanczos3)
	}

	outputFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	fmt.Println("Converting -  Quality:", p.Quality, "Lossless:", p.Lossless)
	if p.Quality <= 0.0 || p.Quality > 100.0 {
		fmt.Println("Converting -  Quality is not a float32 value or out of range... Setting default value")
		p.Quality = NewConvertResizeParams().Quality
	}
	err = webp.Encode(outputFile, img, &webp.Options{
		Lossless: p.Lossless,
		Quality:  p.Quality,
	})
	if err != nil {
		return err
	}

	return nil
}
