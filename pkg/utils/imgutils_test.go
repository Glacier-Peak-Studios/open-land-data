package utils_test

import (
	"fmt"
	"image"
	"testing"
)

func TestImageOverview(t *testing.T) {
	imgSize := 256
	for i := 0; i < 4; i++ {
		rect := image.Rect((i%2)*imgSize, (i/2)*imgSize, (i%2+1)*imgSize, (i/2+1)*imgSize)
		fmt.Printf("Rect %d: %v\n", i, rect)
		// draw.Draw(bgImg, rect, imgRef, image.Point{}, draw.Over)
	}

}
