package config

import (
	"errors"
	"fmt"
)

// Config holds framing and sample parameters.
type Config struct {
	SampleRate int // sample rate in Hz (required)
	FrameSize  int // N: samples per frame (if 0 -> default 2048)
	Hop        int // H: hop size in samples (if 0 -> default FrameSize/2)
	NumBins    int // number of FFT bins to use per frame for pHash (default 32)
}

// DefaultConfig returns common defaults.
func DefaultConfig(sr int) Config {
	const defaultFrame = 2048
	const defaultBins = 64
	if sr <= 0 {
		sr = 44100
	}
	return Config{
		SampleRate: sr,
		FrameSize:  defaultFrame,
		Hop:        defaultFrame / 2,
		NumBins:    defaultBins,
	}
}

// ValidateAndFill normalizes zero values and checks constraints.
func (c *Config) ValidateAndFill() error {
	if c.SampleRate <= 0 {
		return errors.New("sample rate must be > 0")
	}
	if c.FrameSize <= 0 {
		c.FrameSize = 2048
	}
	if c.Hop <= 0 {
		c.Hop = c.FrameSize / 2
	}
	if c.Hop <= 0 || c.Hop > c.FrameSize {
		return errors.New("invalid hop: must be 1..FrameSize")
	}
	if !isPowerOfTwo(c.FrameSize) {
		return fmt.Errorf("frameSize must be a power of two (got %d)", c.FrameSize)
	}
	return nil
}

// isPowerOfTwo returns true if x is power-of-two.
func isPowerOfTwo(x int) bool {
	return x > 0 && (x&(x-1)) == 0
}
