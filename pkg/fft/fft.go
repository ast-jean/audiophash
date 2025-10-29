package fft

import (
	"math"

	"gonum.org/v1/gonum/dsp/fourier"
)

// ComputeMagnitude computes the FFT of a single frame and returns the magnitude spectrum.
// Input:
//
//	frame []float64 : time-domain samples (length N, typically power of 2)
//
// Output:
//
//	[]float64      : magnitudes of bins 0..N/2 (real, non-negative)
func ComputeMagnitude(frame []float64) []float64 {
	N := len(frame)
	if N == 0 {
		return nil
	}

	fft := fourier.NewFFT(N)
	complexResult := fft.Coefficients(nil, frame)

	// Only need first N/2 bins (positive frequencies)
	mags := make([]float64, N/2)
	for i := 0; i < N/2; i++ {
		mags[i] = cmplxAbs(complexResult[i])
	}

	return mags
}

// cmplxAbs returns the magnitude of a complex number.
func cmplxAbs(c complex128) float64 {
	return math.Hypot(real(c), imag(c))
}
