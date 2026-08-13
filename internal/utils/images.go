package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"os"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

const (
	WEBP_DEFAULT_QUALITY = 75.0
	WEBP_SMALL_QUALITY   = 50.0
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
	defer func() {
		if err := inputFile.Close(); err != nil {
			slog.Error("Error closing input file", "err", err)
		}
	}()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		return err
	}

	if p.Width != 0 || p.Height != 0 {
		slog.Debug("Resizing image", "width", p.Width, "height", p.Height)
		// If one value is 0, resize.Resize maintains aspect ratio
		img = resize.Resize(p.Width, p.Height, img, resize.Lanczos3)
	}

	outputFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			slog.Error("Error closing output file", "err", err)
		}
	}()

	slog.Debug("Converting image", "quality", p.Quality, "lossless", p.Lossless)
	if p.Quality <= 0.0 || p.Quality > 100.0 {
		slog.Warn("image quality out of range, setting default value", "quality", p.Quality)
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
