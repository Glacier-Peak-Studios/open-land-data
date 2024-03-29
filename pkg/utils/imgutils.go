package utils

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"math"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
)

var WHITE = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
var TRANSPARENT = color.NRGBA{R: 0, G: 0, B: 0, A: 0}

// GenerateOverviewTile takes in 4 image paths
// and creates a quad layout image of size 512x512.
// Layout format is:
//
// | img1 | img2 |
//
// | img3 | img4 |
func GenerateOverviewTile(imgOutPath, img1Path, img2Path, img3Path, img4Path string) error {
	imgPaths := []string{img1Path, img2Path, img3Path, img4Path}
	imgRefs := make([]image.Image, 4)
	for i, imgPath := range imgPaths {
		img, err := os.Open(imgPath)
		if err != nil {
			defer img.Close()
		}
		var imgDec image.Image
		if err != nil {
			log.Debug().Msgf("Could not open image, using white: %v", imgPath)
			imgDec = image.NewUniform(TRANSPARENT)
		} else {
			imgDec, _ = png.Decode(img)
		}
		imgRefs[i] = imgDec
	}

	bgWidth, bgHeight := 512, 512
	bgImg := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))

	imgSize := 256
	for i, imgRef := range imgRefs {
		if imgRef != nil {
			rect := image.Rect((i%2)*imgSize, (i/2)*imgSize, (i%2+1)*imgSize, (i/2+1)*imgSize)
			draw.Draw(bgImg, rect, imgRef, image.Point{}, draw.Over)

		}
	}

	imgOut := resize.Resize(uint(imgSize), uint(imgSize), bgImg, resize.NearestNeighbor)

	var perms fs.FileMode = 0755 // read/execute all, write owner
	os.MkdirAll(filepath.Dir(imgOutPath), perms)
	err := EncodePNGToPath(imgOutPath, imgOut)

	return err
}

func MergeNTiles0(imgPaths []string, tile Tile, basePath string, outImg string) error {
	whiteTolerance := 0.1
	bgImg := image.NewRGBA(image.Rect(0, 0, 256, 256))
	draw.Draw(bgImg, bgImg.Bounds(), &image.Uniform{TRANSPARENT}, image.Point{}, draw.Src)
	for _, imgPath := range imgPaths {
		base := basePath + imgPath
		imgPathWBase := appendTileToBase(base, tile) + ".png"
		img, err := DecodePNGFromPath(imgPathWBase)
		if err == nil {
			// blend.BlendImage(bgImg, img, blend.Multiply)
			if !imgIsWhite(img, whiteTolerance) && !imgIsTransparent(img) {
				draw.Draw(bgImg, img.Bounds(), img, image.Point{}, draw.Over)
			}
			// log.Debug().Msgf("Could not open image, using transparent: %v", imgPathWBase)
			// img = image.NewUniform(TRANSP)
		}
		// blend.BlendImage(bgImg, img, blend.Multiply)
		// draw.Draw(bgImg, img.Bounds(), img, image.Point{}, draw.Over)
	}
	err := EncodePNGToPath(outImg, bgImg)
	return err
}

func MergeTiles(img1 string, img2 string, outImg string) error {

	img1D, err := DecodePNGFromPath(img1)
	if err != nil {
		return err
	}

	img2D, err := DecodePNGFromPath(img2)
	if err != nil {
		return err
	}

	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	transp := color.NRGBA{R: 0, G: 0, B: 0, A: 0}

	bgWidth, bgHeight := 256, 256
	bgImg := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))

	draw.Draw(bgImg, bgImg.Bounds(), &image.Uniform{transp}, image.Point{}, draw.Src)

	img1WhiteP := GetPixelPercent(img1D, white)
	img2WhiteP := GetPixelPercent(img2D, white)
	img1TransP := GetPixelPercent(img1D, transp)
	img2TransP := GetPixelPercent(img2D, transp)

	if math.Abs(img1WhiteP-img2WhiteP) < 0.25 {
		img1D = ReplaceColor(img1D, white, transp)
		img2D = ReplaceColor(img2D, white, transp)
	}

	// Determine drawing order by white content
	if img1WhiteP > img2WhiteP || img1TransP > img2TransP {
		draw.Draw(bgImg, img1D.Bounds(), img1D, image.Point{}, draw.Over)
		draw.Draw(bgImg, img2D.Bounds(), img2D, image.Point{}, draw.Over)
	} else {
		draw.Draw(bgImg, img2D.Bounds(), img2D, image.Point{}, draw.Over)
		draw.Draw(bgImg, img1D.Bounds(), img1D, image.Point{}, draw.Over)
	}
	err = EncodePNGToPath(outImg, bgImg)
	return err
}

// GetPixelPercent calculates the percentage makeup of a specific color in a given image
func GetPixelPercent(img image.Image, col color.Color) float64 {
	countColor := 0
	size := img.Bounds().Max
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			pxCol := img.At(x, y)
			if pxCol == col {
				countColor++
			}
		}
	}
	return float64(countColor) / float64(size.X*size.Y)
}

func isImgWhiteOrTransparent(imgPath string) bool {
	img, err := DecodePNGFromPath(imgPath)
	if err != nil {
		return false
	}
	notWhiteCount := 0
	size := img.Bounds().Max
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			pxCol := img.At(x, y)
			if !pixelIsTransparent(pxCol) && !pixelIsWhite(pxCol, 0.1) {
				notWhiteCount++
			}
		}
	}
	return notWhiteCount == 0
}

// GetColorDistance computes the distance between two colors
func GetColorDistance(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return math.Sqrt(math.Pow(float64(r1)-float64(r2), 2) + math.Pow(float64(g1)-float64(g2), 2) + math.Pow(float64(b1)-float64(b2), 2))
}

func pixelIsTransparent(col color.Color) bool {
	_, _, _, a := col.RGBA()
	return a == 0
}

func pixelIsWhite(col color.Color, tolerance float64) bool {
	diff := GetColorDistance(col, WHITE)
	maxDistance := 113509.949674
	percentDiff := diff / maxDistance
	// fmt.Printf("Color distance: %f\n", a)
	// fmt.Printf("Color space distance percent: %f\n\n", percent)
	return percentDiff < tolerance
}

// Checks if an image contains all transparent
// pixels
func imgIsTransparent(img image.Image) bool {
	size := img.Bounds().Max
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			pxCol := img.At(x, y)
			if !pixelIsTransparent(pxCol) {
				return false
			}
		}
	}
	return true
}

// Checks if an image contains all white
// pixels within a certain tolerance
func imgIsWhite(img image.Image, tolerance float64) bool {
	size := img.Bounds().Max
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			pxCol := img.At(x, y)
			if !pixelIsWhite(pxCol, tolerance) {
				return false
			}
		}
	}
	return true
}

func ImgOverRects(img image.Image, rects []image.Rectangle) image.Image {
	bgWidth, bgHeight := 256, 256
	bgImg := image.NewRGBA(image.Rect(0, 0, bgWidth, bgHeight))

	for _, rect := range rects {
		draw.Draw(bgImg, rect, &image.Uniform{color.White}, image.Point{}, draw.Over)
	}

	draw.Draw(bgImg, img.Bounds(), img, image.Point{}, draw.Over)

	return bgImg
}

func GetCoverageRectSide(img image.Image, edge int) (image.Rectangle, error) {
	// img, _ := decodePNGFromPath(imgPath)
	pxWhiteTolerance := 0.1
	x, y := 0, 0
	pxRng := IntRange(0, 256)
	if edge%2 == 1 {
		pxRng = IntRange(256, 0)
	}
	outer := &x
	inner := &y
	if edge > 1 {
		outer = &y
		inner = &x
	}

	// size := img.Bounds().Max
	// m := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	edgeFound := false
	for _, *outer = range pxRng {
		for _, *inner = range pxRng {
			pxCol := img.At(x, y)
			if !pixelIsTransparent(pxCol) && !pixelIsWhite(pxCol, pxWhiteTolerance) {
				edgeFound = true
				break
			}
		}
		if edgeFound {
			*inner = pxRng[0]
			break
		}
	}
	pxRngLastIdx := len(pxRng) - 1
	covgRect := image.Rect(x, y, pxRng[pxRngLastIdx], pxRng[pxRngLastIdx])

	return covgRect, nil

}

func GetCoverageRectCorner(img image.Image, corner int) ([]image.Rectangle, error) {
	// img, _ := decodePNGFromPath(imgPath)
	// x, y := 0, 0
	pxWhiteTolerance := 0.1

	xRng := IntRange(0, 256)
	if corner%2 == 1 {
		xRng = IntRange(256, 0)
	}
	yRng := IntRange(0, 256)
	if corner > 1 {
		yRng = IntRange(256, 0)
	}
	xFound, yFound := false, false
	xIdx, yIdx := 0, 0
	for !(xFound && yFound) {
		for revIdx := xIdx; revIdx >= 0 && !xFound; revIdx-- {
			x := xRng[revIdx]
			y := yRng[yIdx]
			pxCol := img.At(x, y)
			if !pixelIsTransparent(pxCol) && !pixelIsWhite(pxCol, pxWhiteTolerance) {
				xFound = true
			}

		}
		if xIdx == 256 {
			xFound = true
		}

		for revIdx := yIdx; revIdx >= 0 && !yFound; revIdx-- {
			x := xRng[xIdx]
			y := yRng[revIdx]
			pxCol := img.At(x, y)
			if !pixelIsTransparent(pxCol) && !pixelIsWhite(pxCol, pxWhiteTolerance) {
				yFound = true
			}
		}
		if yIdx == 256 {
			yFound = true
		}

		if !xFound {
			xIdx++
		}
		if !yFound {
			yIdx++
		}

	}

	rect1 := image.Rect(xRng[xIdx], yRng[0], xRng[len(xRng)-1], yRng[len(yRng)-1])
	rect2 := image.Rect(xRng[0], yRng[yIdx], xRng[len(xRng)-1], yRng[len(yRng)-1])

	return []image.Rectangle{rect1, rect2}, nil

}

// ReplaceColor finds all pixels in an image that matches a specific color,
// and replace them with another color
func ReplaceColor(img image.Image, find, replace color.Color) image.Image {
	size := img.Bounds().Max
	m := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			pxCol := img.At(x, y)
			if pxCol == find {
				m.Set(x, y, replace)
			} else {
				m.Set(x, y, pxCol)
			}
		}
	}

	return m
}

func DecodePNGFromPath(imgPath string) (image.Image, error) {
	imgFile, err := os.Open(imgPath)
	if err != nil {
		log.Error().Err(err).Msgf("Could not open img: %v", imgPath)
		return nil, err
	}
	defer imgFile.Close()
	img, err := png.Decode(imgFile)
	if err != nil {
		log.Error().Err(err).Msgf("Could not decode img: %v", imgPath)
		return nil, err
	}
	return img, nil
}

func EncodePNGToPath(imgPath string, img image.Image) error {
	out, err := os.Create(imgPath)
	if err != nil {
		log.Err(err).Msgf("Could not create output file: %v", imgPath)
		// log.Error().Msgf("Could not create output file: %v", imgPath)
		return err
	}
	defer out.Close()
	err = png.Encode(out, img)
	if err != nil {
		log.Err(err).Msg("Could not encode output image")
		// log.Error().Msg("Could not encode output image")
	}
	return err
}
