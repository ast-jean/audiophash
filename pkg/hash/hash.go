package hash

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

// AudioPHashFromFeature converts a global feature vector to 64-bit hex pHash.
func AudioPHashFromFeature(globalFeature []float64) string {
	if len(globalFeature) == 0 {
		return ""
	}

	// Ensure length is 64 for 64-bit hash
	feature := make([]float64, 64)
	copy(feature, globalFeature)
	for i := len(globalFeature); i < 64; i++ {
		feature[i] = 0
	}

	// Compute median
	medianVal := median(feature)

	var hash uint64
	for i, val := range feature {
		if val > medianVal {
			hash |= 1 << uint(63-i) // MSB first
		}
	}

	return fmt.Sprintf("%016x", hash)
}

// median computes median of a slice
func median(arr []float64) float64 {
	n := len(arr)
	sorted := make([]float64, n)
	copy(sorted, arr)
	sort.Float64s(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// HexToUint64 decodes 16-char hex (64-bit) to uint64
func HexToUint64(hexStr string) (uint64, error) {
	if len(hexStr) != 16 {
		// also allow leading 0s omitted? require 16 for strictness
		return 0, errors.New("hex must be 16 chars")
	}
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, err
	}
	var v uint64
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(b[i])
	}
	return v, nil
}
