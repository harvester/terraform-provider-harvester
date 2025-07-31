package conversion

import "fmt"

const (
	MinInt32  = -(1 << 32)
	MaxInt32  = 1<<31 - 1
	MinUint   = 0
	MinUint32 = 0
	MaxUint32 = 1<<32 - 1
)

func IntToUint32(v int) (uint32, error) {
	if v >= MinUint32 && v <= MaxUint32 {
		return uint32(v), nil //nolint:gosec
	}
	return 0, fmt.Errorf("Overflow during typecast from int to uint32")
}

func IntToUint(v int) (uint, error) {
	if v >= MinUint {
		return uint(v), nil //nolint:gosec
	}
	return 0, fmt.Errorf("Overflow during typecast from int to uint")
}

func IntToInt32(v int) (int32, error) {
	if v >= MinInt32 && v <= MaxInt32 {
		return int32(v), nil //nolint:gosec
	}
	return 0, fmt.Errorf("Overflow during typecast from int to uint")
}
