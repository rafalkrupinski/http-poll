package uints64

const MIN = uint64(0)
const MAX = ^uint64(0)

func Max(a uint64, b uint64) uint64 {
	if Compare(a, b) > 1 {
		return a
	} else {
		return b
	}
}

func Min(a uint64, b uint64) uint64 {
	if Compare(a, b) < 1 {
		return a
	} else {
		return b
	}
}

func Compare(a uint64, b uint64) int {
	if a > b {
		return 1
	} else if a < b {
		return -1
	} else {
		return 0
	}
}
