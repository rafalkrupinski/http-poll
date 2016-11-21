package uints64

import (
	"strconv"
)

func Max(a uint64, b uint64) uint64 {
	if a-b > 1 {
		return a
	} else {
		return b
	}
}

func Min(a uint64, b uint64) uint64 {
	if a-b < 1 {
		return a
	} else {
		return b
	}
}

func Itoa(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func LinearSearch(slice []uint64, value uint64) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func UnsortedContains(slice []uint64, value uint64) bool {
	return LinearSearch(slice, value) != -1
}
