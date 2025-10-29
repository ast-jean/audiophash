package features

import (
	"math"
	"sort"
)

// ExtractGlobalFeature computes a global feature vector from frame FFT magnitudes.
// Uses config.NumBins low-frequency bins and averages across frames.
func ExtractGlobalFeature(frameMags [][]float64, numBins int) []float64 {
	if len(frameMags) == 0 || numBins <= 0 {
		return nil
	}

	// Clamp numBins to the available bins in the first frame
	if numBins > len(frameMags[0]) {
		numBins = len(frameMags[0])
	}

	globalFeature := make([]float64, numBins)

	for bin := 0; bin < numBins; bin++ {
		sum := 0.0
		for _, f := range frameMags {
			sum += f[bin]
		}
		globalFeature[bin] = sum / float64(len(frameMags)) // mean per bin
	}

	return globalFeature
}

// Optional: apply log scaling for perceptual robustness
func LogScaleFeature(feature []float64) {
	for i := range feature {
		feature[i] = math.Log(1 + feature[i])
	}
}

// AggregateGlobalFeature aggregates per-frame magnitude spectra into a single global feature vector.
// Uses mean across frames per bin. Optionally clamp to NumBins.
func AggregateGlobalFeature(frameMags [][]float64, numBins int) []float64 {
	if len(frameMags) == 0 || numBins <= 0 {
		return nil
	}

	// Clamp numBins to available bins in first frame
	if numBins > len(frameMags[0]) {
		numBins = len(frameMags[0])
	}

	globalFeature := make([]float64, numBins)

	for bin := 0; bin < numBins; bin++ {
		sum := 0.0
		for _, f := range frameMags {
			sum += f[bin]
		}
		globalFeature[bin] = sum / float64(len(frameMags)) // mean
	}

	return globalFeature
}

// median aggregation for more robustness
func AggregateGlobalFeatureMedian(frameMags [][]float64, numBins int) []float64 {
	if len(frameMags) == 0 || numBins <= 0 {
		return nil
	}

	if numBins > len(frameMags[0]) {
		numBins = len(frameMags[0])
	}

	globalFeature := make([]float64, numBins)
	for bin := 0; bin < numBins; bin++ {
		values := make([]float64, len(frameMags))
		for i, f := range frameMags {
			values[i] = f[bin]
		}
		globalFeature[bin] = median(values)
	}

	return globalFeature
}

// median computes median of float64 slice
func median(arr []float64) float64 {
	n := len(arr)
	if n == 0 {
		return 0
	}

	sorted := make([]float64, n)
	copy(sorted, arr)
	sort.Float64s(sorted)

	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}
