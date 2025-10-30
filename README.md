# Audio Perceptual Hashing in Golang

This project implements a **16-character perceptual hash (pHash) for audio files** using Go. It is designed to produce a compact fingerprint representing the perceptual content of an audio file, robust to minor distortions, volume changes, compression, and truncation.

## Project Overview

The system is divided into modular components for readability, maintainability, and extensibility:

```
audiophash/
├── cmd/                  # Main entry points / CLI
├── pkg/                  # Core library
│   ├── audio/            # Audio reading, normalization, framing
│   ├── fft/              # Frequency domain operations
│   ├── features/         # Feature extraction (spectral, MFCC, etc.)
│   ├── hash/             # Hash generation and comparison
│   └── utils/            # Misc utilities
├── test/                 # Test data and scripts
├── go.mod
└── README.md
```

## Technical Approach

### 1. Audio Input & Preprocessing

* Accepts raw PCM bytes or WAV files.
* Converts stereo to mono.
* Normalizes amplitude to a fixed range (-1.0 to 1.0).
* Splits audio into overlapping frames (2048 samples, 50% overlap).
* Applies a Hann window to reduce spectral leakage.

### 2. Frequency Domain Conversion

* Performs **Fast Fourier Transform (FFT)** on each frame.
* Optionally converts magnitudes to the Mel scale for perceptual relevance.
* Extracts low-frequency bins (first 32–64) for hashing.

### 3. Feature Aggregation

* Aggregates frame-level features to a global feature vector.
* Uses the average or median across frames for robustness.

### 4. Hash Generation

* Computes the median of the feature vector.
* Binarizes each feature relative to the median:

  * Feature > median → 1
  * Feature ≤ median → 0
* Combines binary features into a 64-bit hash.
* Converts binary hash to a **16-character hexadecimal string**.

### 5. Hash Comparison

* Computes **Hamming distance** between two hashes.
* Measures perceptual similarity between audio files.

## Usage Examples

### CLI

```
audiophash hash file.wav
# Outputs: 16-character hex hash

audiophash compare file1.wav file2.wav
# Outputs: Hamming distance
```

### Go API

```go
import "audiophash/pkg/hash"

audioBytes, _ := os.ReadFile("file.wav")
h := hash.AudioPHash(audioBytes, 44100)
dist := hash.HammingDistance(h1, h2)
```

## Key Features

* **Robustness:** Small distortions, volume changes, and truncation minimally affect the hash.
* **Compact:** Produces a 16-character hex string (64 bits).
* **Modular Architecture:** Easy to extend with new feature extraction methods (MFCC, spectral contrast).
* **Testable:** Unit and integration tests for all modules.

## References

* [pHash.org – Perceptual Hashing](https://www.phash.org/)
* Zauner, C. (2010). Implementation and Benchmarking of Perceptual Image Hashing Algorithms.
* Haytham Fayek. Speech Processing for Machine Learning: MFCC feature extraction.
* IEEE Audio Fingerprinting Research.

## Possible contribution enhancements

* Streaming audio processing.
    - allow incremental processing from a stream or audio device (microphone, live input).
* Parallel FFT and feature extraction for large datasets.
    - process frames concurrently (Go goroutines, channels, worker pool).
* Use MFCCs instead of raw FFT for more perceptual accuracy.
    - use Mel-frequency cepstral coefficients (MFCCs) for features
