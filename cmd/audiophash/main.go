// pkg/audiophash/audiophash.go
package audiophash

import (
	"errors"
	"fmt"

	"github.com/ast-jean/audiophash/pkg/audio"
	"github.com/ast-jean/audiophash/pkg/config"
	"github.com/ast-jean/audiophash/pkg/features"
	"github.com/ast-jean/audiophash/pkg/fft"
	"github.com/ast-jean/audiophash/pkg/hash"
)

// AudioPHashBytes is the canonical entry point for the perceptual hash.
// - b: raw audio bytes (PCM16/ WAV / MP3 bytes depending on fileformat).
// - cfg: optional pointer to config.Config. If nil, config.DefaultConfig(44100) is used.
// - fileformat: "pcm16", "pcm16le", "wav", "mp3", etc. (decoder must be implemented in audio pkg)
// Returns a 16-character hex string (64-bit hash) or an error.
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
	var (
		samples []float64
		sr      int
		err     error
	)

	switch fileformat {
	case "pcm16", "pcm16le":
		samples, sr, err = audio.DecodePCM16LEToFloat64(b)
		if err != nil {
			return "", fmt.Errorf("decode PCM16LE: %w", err)
		}

	case "wav":
		samples, sr, err = audio.DecodeWAVToFloat64(b)
		if err != nil {
			return "", fmt.Errorf("decode WAV: %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported audio format: %s", fileformat)
	}

	// ---------------------------
	// Resample if needed (decoder returns sr; raw PCM may return sr==0)
	// ---------------------------
	if sr != 0 && sr != localCfg.SampleRate {
		samples, err = audio.Resample(samples, sr, localCfg.SampleRate)
		if err != nil {
			return "", fmt.Errorf("resample: %w", err)
		}
	}

	// ---------------------------
	// Normalize amplitude
	// ---------------------------
	samples = audio.Normalize(samples)

	// ---------------------------
	// Framing & windowing
	// ---------------------------
	frames := audio.Frame(samples, localCfg.FrameSize, localCfg.Hop)
	if len(frames) == 0 {
		return "", errors.New("no frames produced (audio too short?)")
	}

	// ---------------------------
	// FFT per frame -> magnitude spectra
	// ---------------------------
	frameMags := make([][]float64, len(frames))
	for i, f := range frames {
		frameMags[i] = fft.ComputeMagnitude(f)
		if frameMags[i] == nil {
			return "", errors.New("fft compute magnitude returned nil (ensure fft.ComputeMagnitude is implemented)")
		}
	}

	// ---------------------------
	// Aggregate to global feature vector (use median aggregation for robustness)
	// ---------------------------
	globalFeature := features.AggregateGlobalFeatureMedian(frameMags, localCfg.NumBins)
	if len(globalFeature) == 0 {
		return "", errors.New("no global feature produced")
	}
	features.LogScaleFeature(globalFeature)

	// ---------------------------
	// PHash from feature -> 16-char hex
	// ---------------------------
	hashHex := hash.AudioPHashFromFeature(globalFeature)
	if hashHex == "" {
		return "", errors.New("failed to compute pHash")
	}

	return hashHex, nil
}
