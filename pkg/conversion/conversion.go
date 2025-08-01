package conversion

import (
	"fmt"
	"math"
)

func IntToUint32(v int) (uint32, error) {
	if v >= 0 && v <= math.MaxUint32 {
		return uint32(v), nil //nolint:gosec
	}
	return 0, fmt.Errorf("overflow during typecast from int to uint32")
}

func IntToUint(v int) (uint, error) {
	if v >= 0 {
		return uint(v), nil //nolint:gosec
	}
	return 0, fmt.Errorf("overflow during typecast from int to uint")
}

func IntToInt32(v int) (int32, error) {
	if v >= math.MinInt32 && v <= math.MaxInt32 {
		return int32(v), nil //nolint:gosec
	}
	return 0, fmt.Errorf("overflow during typecast from int to uint")
}
