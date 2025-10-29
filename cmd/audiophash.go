// pkg/audiophash/audiophash.go
package audiophash

import (
	"errors"
	"fmt"

	"pkg/audio"
	"pkg/config"
	"pkg/features"
	"pkg/fft"
	"pkg/hash"
)

// AudioPHashBytes is the canonical entry point for the perceptual hash.
// - b: raw audio bytes (currently we expect PCM16LE raw bytes or WAV if you implement the WAV decoder).
// - cfg: optional pointer to config.Config. If nil, config.DefaultConfig(44100) is used.
// Returns a 16-character hex string (64-bit hash) or an error.
//
// This file wires the pipeline; each step delegates to its package. Many steps are placeholders
func AudioPHashBytes(b []byte, cfg *config.Config, fileformat string) (string, error) {
	// ---------------------------
	// Defaults & validation
	// ---------------------------
	var localCfg config.Config
	if cfg == nil {
		localCfg = config.DefaultConfig(44100)
	} else {
		localCfg = *cfg
	}
	if err := localCfg.ValidateAndFill(); err != nil {
		return "", err
	}
	if len(b) == 0 {
		return "", errors.New("input bytes empty")
	}

	// ---------------------------
	// Decode -> []float64 samples (mono)
	// ---------------------------
	// Expect audio package to provide DecodePCM16LEToFloat64 or WAV decoder.
	// decode should return samples and sampleRate (sr). For raw PCM16LE raw bytes the SR
	// is not encoded in the bytes — decoder may accept an expected SR parameter or return 0.
	switch fileformat {
	case "pcm16", "pcm16le":
		samples, sr, err = decode.DecodePCM16LEToFloat64(b)
		if err != nil {
			return "", fmt.Errorf("decode PCM16LE: %w", err)
		}

	case "wav":
		samples, sr, err = decode.DecodeWAVToFloat64(b)
		if err != nil {
			return "", fmt.Errorf("decode WAV: %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported audio format: %s", format)
	}
	// Resample according to Config
	if sr != 0 && sr != localCfg.SampleRate {
		samples, err = audio.Resample(samples, sr, localCfg.SampleRate)
		if err != nil {
			return "", fmt.Errorf("resample: %w", err)
		}
	}
	// ---------------------------
	// Normalize amplitude
	// ---------------------------
	// Normalize to [-1.0, +1.0] using peak normalization (or RMS if preferred).
	samples = audio.Normalize(samples)

	// ---------------------------
	// Framing & windowing
	// ---------------------------
	// Use configured frame size & hop. audio.Frame returns windowed frames (each length FrameSize).
	frames := audio.Frame(samples, localCfg.FrameSize, localCfg.Hop)
	if len(frames) == 0 {
		return "", errors.New("no frames produced (audio too short?)")
	}

	// ---------------------------
	// FFT per frame -> magnitude spectra
	// ---------------------------
	// fft package should expose ComputeMagnitude(frame []float64) []float64
	// which computes an N-point FFT (N==len(frame)) and returns magnitudes for bins 0..N/2.
	frameMags := make([][]float64, len(frames))
	for i, f := range frames {
		frameMags[i] = fft.ComputeMagnitude(f)
	}
	// ---------------------------
	// Feature extraction (per-frame)
	// ---------------------------
	// features.Extract should take spectra and produce per-frame feature vectors.
	// frameMags [][]float64 obtained from FFT
	globalFeature := features.ExtractGlobalFeature(frameMags, localCfg.NumBins)
	features.LogScaleFeature(globalFeature)

	// ---------------------------
	// Aggregate to global feature vector
	// ---------------------------
	// Aggregate per-bin over frames using median.
	globalFeature := features.AggregateGlobalFeatureMedian(frameMags, localCfg.NumBins)

	// ---------------------------
	// PHash from feature
	// ---------------------------
	globalFeature := features.AggregateGlobalFeature(frameMags, localCfg.NumBins)
	features.LogScaleFeature(globalFeature)

	hashHex := hash.AudioPHashFromFeature(globalFeature)

	return hashHex, nil
}
