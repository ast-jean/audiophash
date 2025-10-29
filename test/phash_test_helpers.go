package tests

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/bits"
	"testing"
)

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

// HammingDistance between two uint64 hashes
func HammingDistance(h1, h2 uint64) int {
	return bits.OnesCount64(h1 ^ h2)
}

// HammingPercent (0..100)
func HammingPercent(h1, h2 uint64) float64 {
	return float64(HammingDistance(h1, h2)) / 64.0 * 100.0
}

// loadFile reads file bytes (helper)
func loadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return b
}
