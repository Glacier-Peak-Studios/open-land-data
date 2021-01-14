package utils

import (
	"os"
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

func FilterFI(vs []os.FileInfo, f func(os.FileInfo) bool) []os.FileInfo {
	vsf := make([]os.FileInfo, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
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
