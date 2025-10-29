package audio

import "math"

// Frame splits audio samples into overlapping frames and applies a Hann window.
// Inputs:
//
//	samples   []float64 : mono audio samples
//	frameSize int       : number of samples per frame (N), power of 2 recommended for FFT
//	hop       int       : hop size (H), number of samples to advance for each frame
//
// Output:
//
//	[][]float64 : 2D slice, each inner slice is one windowed frame
func Frame(samples []float64, frameSize, hop int) [][]float64 {
	if frameSize <= 0 || hop <= 0 || hop > frameSize {
		return nil // caller must validate config
	}

	numFrames := 1 + (len(samples)-frameSize)/hop
	if numFrames < 1 {
		numFrames = 0
	}
	frames := make([][]float64, 0, numFrames)

	// Precompute Hann window
	window := make([]float64, frameSize)
	for i := 0; i < frameSize; i++ {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(frameSize-1)))
	}

	for start := 0; start+frameSize <= len(samples); start += hop {
		frame := make([]float64, frameSize)
		for i := 0; i < frameSize; i++ {
			frame[i] = samples[start+i] * window[i]
		}
		frames = append(frames, frame)
	}

	return frames
}
