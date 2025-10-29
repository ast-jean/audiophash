package audio

import (
	"errors"
	"math"
)

// Resample linearly resamples audio from `fromHz` to `toHz`.
// Input:
//
//	samples []float64 : original audio samples
//	fromHz int        : sample rate of input
//	toHz   int        : desired output sample rate
//
// Output:
//
//	[]float64 : resampled audio
//	error     : non-nil if input invalid
func Resample(samples []float64, fromHz, toHz int) ([]float64, error) {
	if fromHz <= 0 || toHz <= 0 {
		return nil, errors.New("invalid sample rate")
	}
	if len(samples) == 0 {
		return nil, errors.New("no samples to resample")
	}

	if fromHz == toHz {
		// No resampling needed
		out := make([]float64, len(samples))
		copy(out, samples)
		return out, nil
	}

	ratio := float64(toHz) / float64(fromHz)
	newLen := int(float64(len(samples)) * ratio)
	out := make([]float64, newLen)

	for i := 0; i < newLen; i++ {
		// Map output sample index -> input float index
		pos := float64(i) / ratio
		idx := int(pos)
		frac := pos - float64(idx)

		if idx+1 < len(samples) {
			// Linear interpolation
			out[i] = samples[idx]*(1-frac) + samples[idx+1]*frac
		} else {
			// Last sample
			out[i] = samples[len(samples)-1]
		}
	}

	return out, nil
}

// []float64 : normalized audio samples
func Normalize(samples []float64) []float64 {
	if len(samples) == 0 {
		return samples
	}

	// Find max absolute amplitude
	var maxAmp float64
	for _, s := range samples {
		if a := math.Abs(s); a > maxAmp {
			maxAmp = a
		}
	}

	// Avoid division by zero
	if maxAmp == 0 {
		return samples
	}

	// Scale all samples to [-1.0, +1.0]
	normalized := make([]float64, len(samples))
	scale := 1.0 / maxAmp
	for i, s := range samples {
		normalized[i] = s * scale
	}

	return normalized
}
