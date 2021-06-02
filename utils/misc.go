package utils

import (
	"os"
	"path/filepath"
)

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Filter2(vs []string, filt []string, f func(string, []string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v,  filt) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterFI(vs []os.FileInfo, f func(os.FileInfo) bool) []os.FileInfo {
	vsf := make([]os.FileInfo, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Map(vs []string, f func(string) string) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		vF := f(v)
		vsf = append(vsf, vF)
	}
	return vsf
}

func MapFI(vs []os.FileInfo, f func(os.FileInfo) string) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		vF := f(v)
		vsf = append(vsf, vF)
	}
	return vsf
}

func SetMap(slc []string, f func(string) string) []string {
	vsf := make([]string, 0)
	for _, v := range slc {
		vF := f(v)
		if !stringInSlice(vF, vsf) {
			vsf = append(vsf, vF)
		}
	}
	return vsf
}

func AppendSetT(set []Tile, tile Tile) []Tile {
	if !tileInSlice(tile, set) {
		return append(set, tile)
	}
	return set
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func tileInSlice(a Tile, list []Tile) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func OverviewRoot(path string) string {
	tile, basepath := PathToTile(path)
	up := tile.overviewTile()
	root := MakeTile(up.X*2, up.Y*2, up.Z+1)
	return filepath.Join(basepath, root.GetPath())

}

// func Filter[T any](vs []T, f func(T) bool) []T {
// 	vsf := make([]T, 0)
// 	for _, v := range vs {
// 		if f(v) {
// 			vsf = append(vsf, v)
// 		}
// 	}
// 	return vsf
// }

func IntRange(min, max int) []int {
	inc := 1
	if max < min {
		inc = -1
	}
	a := make([]int, AbsInt(max-min)+1)
	for i := range a {
		a[i] = min + (inc * i)
	}
	return a
}

func AbsInt(val int) int {
	if val < 0 {
		val = val * -1
	}
	return val

}
