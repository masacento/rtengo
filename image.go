package rtengo

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"golang.org/x/image/draw"
)

// ImageNet normalization constants.
var (
	ImageNetMean = [3]float32{0.485, 0.456, 0.406}
	ImageNetStd  = [3]float32{0.229, 0.224, 0.225}
)

// Option configures image preprocessing.
type Option func(*imageConfig)

type imageConfig struct {
	Mean [3]float32
	Std  [3]float32
}

// WithNormalization overrides the mean and std used for normalization.
func WithNormalization(mean, std [3]float32) Option {
	return func(cfg *imageConfig) {
		cfg.Mean = mean
		cfg.Std = std
	}
}

// LoadAndPreprocess loads an image from a reader, resizes it, and converts to NCHW tensor format.
func LoadAndPreprocess(ctx context.Context, r io.Reader, width, height int, opts ...Option) ([]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cfg := imageConfig{Mean: ImageNetMean, Std: ImageNetStd}
	for _, opt := range opts {
		opt(&cfg)
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)

	data := make([]float32, 3*height*width)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()

			rf := float32(r) / 65535.0
			gf := float32(g) / 65535.0
			bf := float32(b) / 65535.0

			data[0*(height*width)+y*width+x] = (rf - cfg.Mean[0]) / cfg.Std[0]
			data[1*(height*width)+y*width+x] = (gf - cfg.Mean[1]) / cfg.Std[1]
			data[2*(height*width)+y*width+x] = (bf - cfg.Mean[2]) / cfg.Std[2]
		}
	}

	return data, nil
}

// LoadAndPreprocessFile loads an image from a file path and preprocesses it.
func LoadAndPreprocessFile(ctx context.Context, path string, width, height int, opts ...Option) ([]float32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	return LoadAndPreprocess(ctx, file, width, height, opts...)
}
