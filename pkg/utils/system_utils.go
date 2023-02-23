package utils

import (
	"math"
	"runtime"
)

func GetDefaultWorkers() int {
	maxNum := runtime.NumCPU()

	ep := math.Round(float64(maxNum) * 0.8)

	return int(ep)
}
