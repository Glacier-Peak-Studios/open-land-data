package utils

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/nfnt/resize"
)

func CombineImages(img1 string, img2 string, outImg string) error {
	imgFile1, err := os.Open(img1)
	if err != nil {
		log.Error("Could not open img1")
		return err
	}
	defer imgFile1.Close()
	imgFile2, err := os.Open(img2)
	if err != nil {
		log.Error("Could not open img2")
		return err
	}
	defer imgFile2.Close()
	img1D, err := png.Decode(imgFile1)
	if err != nil {
		log.Error("Could not decode img1")
		return err
	}
	img2D, err := png.Decode(imgFile2)
	if err != nil {
		log.Error("Could not decode img2")
		return err
	}

	bgWidth, bgHeight := 256, 256
	bgImg := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))

	draw.Draw(bgImg, bgImg.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	draw.Draw(bgImg, img1D.Bounds(), img1D, image.ZP, draw.Over)
	draw.Draw(bgImg, img2D.Bounds(), img2D, image.ZP, draw.Over)

	out, err := os.Create(outImg)
	if err != nil {
		log.Error("Could not create output image")
		return err
	}
	defer out.Close()
	// var opt jpeg.Options
	// opt.Quality = 80
	// err = jpeg.Encode(out, bgImg, &opt)
	err = png.Encode(out, bgImg)
	if err != nil {
		log.Error("Could not encode output image")
	}
	return err
}

func GenerateOverview(outName string, img1 string, img2 string, img3 string, img4 string) error {
	imgLocs := []string{img1, img2, img3, img4}
	imgs := make([]image.Image, 4)
	for i, imgLoc := range imgLocs {
		img, err := os.Open(imgLoc)
		defer img.Close()
		var imgDec image.Image
		if err != nil {
			log.Error("Could not open image: ", imgLoc)
			imgDec = image.NewUniform(color.White)
		} else {
			imgDec, err = png.Decode(img)
		}
		imgs[i] = imgDec
	}

	bgWidth, bgHeight := 512, 512
	bgImg := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))
	// draw.Draw(bgImg, bgImg.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	draw.Draw(bgImg, image.Rect(0, 0, 256, 256), imgs[0], image.ZP, draw.Over)
	draw.Draw(bgImg, image.Rect(256, 0, 512, 256), imgs[1], image.ZP, draw.Over)
	draw.Draw(bgImg, image.Rect(0, 256, 256, 512), imgs[2], image.ZP, draw.Over)
	draw.Draw(bgImg, image.Rect(256, 256, 512, 512), imgs[3], image.ZP, draw.Over)

	imgOut := resize.Resize(256, 256, bgImg, resize.MitchellNetravali)

	os.MkdirAll(filepath.Dir(outName), 0755)
	out, err := os.Create(outName)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	png.Encode(out, imgOut)

	return nil

}
