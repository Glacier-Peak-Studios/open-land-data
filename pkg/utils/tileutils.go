package utils

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type Tile struct {
	X int
	Y int
	Z int
}

func MakeTile(x int, y int, z int) Tile {
	return Tile{X: x, Y: y, Z: z}
}

func (t *Tile) GetPath() string {
	xStr := strconv.Itoa(t.X)
	yStr := strconv.Itoa(t.Y)
	zStr := strconv.Itoa(t.Z)
	return filepath.Join(zStr, xStr, yStr)
}

func (t *Tile) GetPathZX() string {
	xStr := strconv.Itoa(t.X)
	zStr := strconv.Itoa(t.Z)
	return filepath.Join(zStr, xStr)
}

func (t *Tile) GetPathXY() string {
	xStr := strconv.Itoa(t.X)
	yStr := strconv.Itoa(t.Y)
	return filepath.Join(xStr, yStr)
}

func (t *Tile) GetXYString() string {

	toJoin := []string{strconv.Itoa(t.X), strconv.Itoa(t.Y)}
	return strings.Join(toJoin, "")
}

func (t *Tile) GetXYInt() int {

	toJoin := []string{strconv.Itoa(t.X), strconv.Itoa(t.Y)}
	i, _ := strconv.Atoi(strings.Join(toJoin, ""))

	return i
}

func (t *Tile) leftTile() Tile {
	return Tile{X: t.X - 1, Y: t.Y, Z: t.Z}
}

func (t *Tile) rightTile() Tile {
	return Tile{X: t.X + 1, Y: t.Y, Z: t.Z}
}

func (t *Tile) upTile() Tile {
	return Tile{X: t.X, Y: t.Y - 1, Z: t.Z}
}

func (t *Tile) downTile() Tile {
	return Tile{X: t.X, Y: t.Y + 1, Z: t.Z}
}

func (t *Tile) overviewTile() Tile {
	return Tile{X: t.X / 2, Y: t.Y / 2, Z: t.Z - 1}
}

func (t *Tile) xyPoint() Point {
	return Point{X: t.X, Y: t.Y}
}

func (t *Tile) xyzPoint() Point3 {
	return Point3{X: t.X, Y: t.Y, Z: t.Z}
}

func PathToTile(path string) (Tile, string) {
	yStr := StripExt(filepath.Base(path))
	fdir := filepath.Dir(path)
	xStr := filepath.Base(fdir)
	fdir = filepath.Dir(fdir)
	zStr := filepath.Base(fdir)
	fdir = filepath.Dir(fdir)
	basepath := fdir

	x, _ := strconv.Atoi(xStr)
	y, _ := strconv.Atoi(yStr)
	z, err := strconv.Atoi(zStr)

	if err != nil {
		log.Debug().Err(err).Msg("")
	}

	retT := Tile{X: x, Y: y, Z: z}

	return retT, basepath
}

// Point represents a 2d point in the cartesian coordinate system.
type Point struct {
	X int
	Y int
}

func NewPoint(x string, y string) (Point, error) {
	xint, _ := strconv.Atoi(x)
	yint, err := strconv.Atoi(y)
	if err != nil {
		log.Error().Msg("Couldn't parse x/y string")
		return Point{X: 0, Y: 0}, err
	}
	return Point{X: xint, Y: yint}, nil
}

type Point3 struct {
	X int
	Y int
	Z int
}

// BBox represents a bounding box in the cartesian coordinate system.
type BBox struct {
	x0 int
	y0 int
	x1 int
	y1 int
}

// BBx creates a BBox from two diagonal Points: the top left (origin) and bottom right (extent)
func BBx(origin Point, extent Point) BBox {
	return BBox{x0: origin.X, y0: origin.Y, x1: extent.X, y1: extent.Y}
}

func (b *BBox) Origin() Point {
	return Point{X: b.x0, Y: b.y0}
}

func (b *BBox) Extent() Point {
	return Point{X: b.x1, Y: b.y1}
}

// PointInBBox checks if a 2d Point is with the bounding box
func (b *BBox) PointInBBox(p Point) bool {
	inXRange := p.X > b.x0 && p.X < b.x1
	inYRange := p.Y > b.y0 && p.Y < b.y1
	return inXRange && inYRange
}

func (b *BBox) ExpandBy(amount int) {
	b.x0 -= amount
	b.x1 += amount
	b.y0 -= amount
	b.y1 += amount
}

func (b *BBox) ShrinkBy(amount int) {
	b.x0 += amount
	b.x1 -= amount
	b.y0 += amount
	b.y1 -= amount
}

// func (b *BBox) ChangeBy(side int, amount int) {
// 	switch side {
// 	case 0:
// 		b.ChangeLeft(amount)
// 	}
// }

func (b *BBox) ChangeLeft(amount int) {
	b.x0 += amount
}

func (b *BBox) ChangeRight(amount int) {
	b.x1 += amount
}

func (b *BBox) ChangeTop(amount int) {
	b.y0 += amount
}

func (b *BBox) ChangeBottom(amount int) {
	b.y1 += amount
}

func (b *BBox) ChangeSide(side string, amount int) {
	switch side {
	case "left":
		b.ChangeLeft(-amount)
	case "right":
		b.ChangeRight(amount)
	case "top":
		b.ChangeTop(-amount)
	case "bottom":
		b.ChangeBottom(amount)
	}
}

func SideToNum(side string) int {
	switch side {
	case "left":
		return 0
	case "right":
		return 1
	case "top":
		return 3
	case "bottom":
		return 2
	}
	return 0
}

func (b *BBox) getSideLine(side string) BBox {
	switch side {
	case "left":
		return BBox{x0: b.x0, y0: b.y0, x1: b.x0, y1: b.y1}
	case "right":
		return BBox{x0: b.x1, y0: b.y0, x1: b.x1, y1: b.y1}
	case "top":
		return BBox{x0: b.x0, y0: b.y0, x1: b.x1, y1: b.y0}
	case "bottom":
		return BBox{x0: b.x0, y0: b.y1, x1: b.x1, y1: b.y1}
	}
	return ZeroBBox()
}

func (b *BBox) isBBoxWhite(basepath string, z int) bool {
	for ix := b.Origin().X; ix <= b.Extent().X; ix++ {
		for iy := b.Origin().Y; iy <= b.Extent().Y; iy++ {
			tile := Tile{X: ix, Y: iy, Z: z}
			imgFile := filepath.Join(basepath, tile.GetPath()+".png")
			if !isImgWhiteOrTransparent(imgFile) {
				return false
			}
		}
	}
	return true
}

func ZeroBBox() BBox {
	return BBox{x0: 0, y0: 0, x1: 0, y1: 0}
}

// GetBBoxIntersect returns the BBox formed when two other BBoxes intersect,
// aka the union of those BBoxes.
// If the passed BBoxes don't intersect, an error will be returned.
func GetBBoxIntersect(b1 BBox, b2 BBox) (BBox, error) {
	var err error = nil
	var intersect BBox
	orig := Point{X: Max(b1.x0, b2.x0), Y: Max(b1.y0, b2.y0)}
	extent := Point{X: Min(b1.x1, b2.x1), Y: Min(b1.y1, b2.y1)}
	if orig.X > extent.X || orig.Y > extent.Y {
		err = errors.New("invalid BBox: no intersection")
	} else {
		intersect = BBx(orig, extent)
	}
	return intersect, err
}

func GetBBoxMerge(b1 BBox, b2 BBox) (BBox, error) {
	var err error = nil
	var outersect BBox
	orig := Point{X: Min(b1.x0, b2.x0), Y: Min(b1.y0, b2.y0)}
	extent := Point{X: Max(b1.x1, b2.x1), Y: Max(b1.y1, b2.y1)}
	if orig.X > extent.X || orig.Y > extent.Y {
		err = errors.New("invalid BBox: Invalid Merge")
	} else {
		outersect = BBx(orig, extent)
	}
	return outersect, err
}
