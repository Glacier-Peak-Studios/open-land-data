package utils

import (
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

func FilterByList[T any](vs []T, filt []T, f func(T, []T) bool) []T {
	vsf := make([]T, 0)
	for _, v := range vs {
		if f(v, filt) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Filter[T any](vs []T, f func(T) bool) []T {
	vsf := make([]T, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Map[T, V any](vs []T, f func(T) V) []V {
	vsf := make([]V, 0)
	for _, v := range vs {
		vF := f(v)
		vsf = append(vsf, vF)
	}
	return vsf
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
