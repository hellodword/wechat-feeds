package main

import (
	"bytes"
	"fmt"
	"github.com/h2non/bimg"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
)

// ToPng converts an image to png
// https://gist.github.com/tizz98/fb15f8dd0c55ac8d2be0e3c4bd8249c3
func ToPng(imageBytes []byte) ([]byte, error) {
	contentType := http.DetectContentType(imageBytes)

	switch contentType {
	case "image/png":
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unable to convert %#v to png", contentType)
}

func formatAndResize(b []byte, size int) []byte {
	var err error

	bPng, _ := ToPng(b)
	if bPng != nil {
		b = bPng
	}

	img := bimg.NewImage(b)
	if img.Type() != bimg.ImageTypeName(bimg.PNG) {
		var newImage []byte
		newImage, err = img.Convert(bimg.PNG)
		if err != nil {
			return b
		}
		img = bimg.NewImage(newImage)
	}

	inputSize, err := img.Size()
	if err != nil {
		return b
	}

	if inputSize.Width != inputSize.Height {
		return b
	}

	if size > inputSize.Width {
		size = int(math.Min(SizeNormal, float64(inputSize.Width)))
	}

	newIcon, err := img.Resize(size, size)

	if err != nil {
		return b
	} else {
		return newIcon
	}
}
