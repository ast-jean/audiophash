// pkg/audiophash/audiophash.go
package audiophash

import (
	"errors"
	"fmt"
	"sort"

	"github.com/ast-jean/audiophash/pkg/audio"
	"github.com/ast-jean/audiophash/pkg/config"
	"github.com/ast-jean/audiophash/pkg/features"
	"github.com/ast-jean/audiophash/pkg/fft"
	"github.com/ast-jean/audiophash/pkg/hash"
)

// AudioPHashBytes is the canonical entry point for the perceptual hash.
// - b: raw audio bytes (PCM16/ WAV / MP3 bytes depending on fileformat).
// - cfg: optional pointer to config.Config. If nil, config.DefaultConfig(44100) is used.
// - fileformat: "pcm16", "pcm16le", "wav". (decoder must be implemented in audio pkg)
// Returns a 16-character hex string (64-bit hash) or an error.
//
// Debugging: set environment variable AUDIOPHASH_DEBUG=1 to enable verbose debug prints.
func AudioPHashBytes(b []byte, cfg *config.Config, fileformat string) (string, error) {
	debug := false

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
	if debug {
		fmt.Printf("[phash] start: bytes=%d format=%q sampleRate(cfg)=%d frameSize=%d hop=%d numBins=%d\n",
			len(b), fileformat, localCfg.SampleRate, localCfg.FrameSize, localCfg.Hop, localCfg.NumBins)
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

	if debug {
		fmt.Printf("[phash] decoded: samples=%d decoder_sr=%d\n", len(samples), sr)
		// show a tiny sample window
		if len(samples) > 0 {
			end := 8
			if len(samples) < end {
				end = len(samples)
			}
			fmt.Printf("[phash] first samples: %v\n", samples[:end])
		}
	}

	// ---------------------------
	// Resample if needed (decoder returns sr; raw PCM may return sr==0)
	// ---------------------------
	if sr != 0 && sr != localCfg.SampleRate {
		if debug {
			fmt.Printf("[phash] resampling: from=%d to=%d\n", sr, localCfg.SampleRate)
		}
		samples, err = audio.Resample(samples, sr, localCfg.SampleRate)
		if err != nil {
			return "", fmt.Errorf("resample: %w", err)
		}
		if debug {
			fmt.Printf("[phash] resampled: samples=%d\n", len(samples))
		}
	}

	// ---------------------------
	// Normalize amplitude
	// ---------------------------
	samples = audio.Normalize(samples)
	if debug {
		fmt.Printf("[phash] normalized: samples=%d\n", len(samples))
		// small stats
		minv, maxv, meanv := statsFloatSlice(samples)
		fmt.Printf("[phash] sample stats: min=%.6f max=%.6f mean=%.6f\n", minv, maxv, meanv)
	}

	// ---------------------------
	// Framing & windowing
	// ---------------------------
	frames := audio.Frame(samples, localCfg.FrameSize, localCfg.Hop)
	if len(frames) == 0 {
		return "", errors.New("no frames produced (audio too short?)")
	}
	if debug {
		fmt.Printf("[phash] framing: frames=%d frameSize=%d hop=%d\n", len(frames), localCfg.FrameSize, localCfg.Hop)
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
	if debug {
		fmt.Printf("[phash] fft: computed magnitude spectra for %d frames (bins per frame=%d)\n", len(frameMags), len(frameMags[0]))
		// print first frame few bins
		binsToShow := 8
		if len(frameMags[0]) < binsToShow {
			binsToShow = len(frameMags[0])
		}
		fmt.Printf("[phash] first frame magnitudes (first %d bins): %v\n", binsToShow, frameMags[0][:binsToShow])
	}

	// ---------------------------
	// Aggregate to global feature vector (use median aggregation for robustness)
	// ---------------------------
	globalFeature := features.AggregateGlobalFeatureMedian(frameMags, localCfg.NumBins)
	if len(globalFeature) == 0 {
		return "", errors.New("no global feature produced")
	}
	if debug {
		minv, maxv, meanv := statsFloatSlice(globalFeature)
		med := medianFloatSlice(globalFeature)
		fmt.Printf("[phash] aggregated feature: len=%d min=%.6f max=%.6f mean=%.6f median=%.6f\n", len(globalFeature), minv, maxv, meanv, med)
	}

	// optional log-scale
	features.LogScaleFeature(globalFeature)
	if debug {
		minv, maxv, meanv := statsFloatSlice(globalFeature)
		med := medianFloatSlice(globalFeature)
		fmt.Printf("[phash] log-scaled feature: len=%d min=%.6f max=%.6f mean=%.6f median=%.6f\n", len(globalFeature), minv, maxv, meanv, med)
	}

	// ---------------------------
	// PHash from feature -> 16-char hex
	// ---------------------------
	hashHex := hash.AudioPHashFromFeature(globalFeature)
	if hashHex == "" {
		return "", errors.New("failed to compute pHash")
	}

	if debug {
		u, _ := hash.HexToUint64(hashHex)
		fmt.Printf("[phash] result: hex=%s uint64=%016x\n", hashHex, u)
	}

	return hashHex, nil
}

// ---- small helpers for debug stats ----

func statsFloatSlice(s []float64) (minv, maxv, meanv float64) {
	if len(s) == 0 {
		return 0, 0, 0
	}
	minv = s[0]
	maxv = s[0]
	sum := 0.0
	for _, v := range s {
		if v < minv {
			minv = v
		}
		if v > maxv {
			maxv = v
		}
		sum += v
	}
	meanv = sum / float64(len(s))
	return minv, maxv, meanv
}

func medianFloatSlice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	tmp := make([]float64, len(s))
	copy(tmp, s)
	sort.Float64s(tmp)
	n := len(tmp)
	if n%2 == 0 {
		return (tmp[n/2-1] + tmp[n/2]) / 2
	}
	return tmp[n/2]
}
