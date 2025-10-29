package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// DecodePCM16LEToFloat64 converts raw 16-bit PCM little-endian bytes to float64 samples in [-1.0, +1.0].
// Input:
//
//	b []byte       : raw PCM16LE bytes. Interleaved channels are not supported in this prototype.
//
// Output:
//
//	[]float64      : normalized mono samples
//	int            : sample rate (0 for raw PCM, since PCM16LE raw bytes do not include SR info)
//	error          : non-nil if decoding fails

func DecodePCM16LEToFloat64(b []byte) ([]float64, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("input byte slice is empty")
	}
	if len(b)%2 != 0 {
		return nil, 0, errors.New("byte length is not multiple of 2, invalid PCM16LE")
	}

	numSamples := len(b) / 2
	samples := make([]float64, numSamples)

	for i := 0; i < numSamples; i++ {
		// Each sample is 2 bytes, little-endian
		offset := i * 2
		raw := int16(binary.LittleEndian.Uint16(b[offset : offset+2]))
		// Normalize to [-1.0, +1.0] float64
		samples[i] = float64(raw) / 32768.0
	}

	// Raw PCM has no sample rate info; return 0.
	return samples, 0, nil
}

// DecodeWAVToFloat64 decodes a WAV file (PCM16) into float64 samples in [-1.0, +1.0].
// Input:
//
//	b []byte : WAV file bytes
//
// Output:
//
//	[]float64 : mono samples
//	int       : sample rate in Hz
//	error     : decoding error
func DecodeWAVToFloat64(b []byte) ([]float64, int, error) {
	if len(b) < 44 {
		return nil, 0, errors.New("WAV too short to contain header")
	}

	r := bytes.NewReader(b)

	// --- RIFF header ---
	var riff [4]byte
	if err := binary.Read(r, binary.LittleEndian, &riff); err != nil {
		return nil, 0, err
	}
	if string(riff[:]) != "RIFF" {
		return nil, 0, errors.New("not a RIFF file")
	}

	var _chunkSize uint32
	if err := binary.Read(r, binary.LittleEndian, &_chunkSize); err != nil {
		return nil, 0, err
	}

	var wave [4]byte
	if err := binary.Read(r, binary.LittleEndian, &wave); err != nil {
		return nil, 0, err
	}
	if string(wave[:]) != "WAVE" {
		return nil, 0, errors.New("not a WAVE file")
	}

	// --- fmt chunk ---
	var fmtHeader [4]byte
	if err := binary.Read(r, binary.LittleEndian, &fmtHeader); err != nil {
		return nil, 0, err
	}
	if string(fmtHeader[:]) != "fmt " {
		return nil, 0, errors.New("expected fmt chunk")
	}

	var fmtSize uint32
	if err := binary.Read(r, binary.LittleEndian, &fmtSize); err != nil {
		return nil, 0, err
	}

	var audioFormat uint16
	if err := binary.Read(r, binary.LittleEndian, &audioFormat); err != nil {
		return nil, 0, err
	}
	if audioFormat != 1 {
		return nil, 0, errors.New("only PCM format supported")
	}

	var numChannels uint16
	if err := binary.Read(r, binary.LittleEndian, &numChannels); err != nil {
		return nil, 0, err
	}

	var sampleRate uint32
	if err := binary.Read(r, binary.LittleEndian, &sampleRate); err != nil {
		return nil, 0, err
	}

	var _byteRate uint32
	if err := binary.Read(r, binary.LittleEndian, &_byteRate); err != nil {
		return nil, 0, err
	}

	var _blockAlign uint16
	if err := binary.Read(r, binary.LittleEndian, &_blockAlign); err != nil {
		return nil, 0, err
	}

	var bitsPerSample uint16
	if err := binary.Read(r, binary.LittleEndian, &bitsPerSample); err != nil {
		return nil, 0, err
	}
	if bitsPerSample != 16 {
		return nil, 0, errors.New("only 16-bit WAV supported")
	}

	// Skip any extra fmt bytes
	if fmtSize > 16 {
		if _, err := r.Seek(int64(fmtSize-16), 1); err != nil {
			return nil, 0, err
		}
	}

	// --- data chunk ---
	var dataHeader [4]byte
	var dataSize uint32
	for {
		if err := binary.Read(r, binary.LittleEndian, &dataHeader); err != nil {
			return nil, 0, err
		}
		if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
			return nil, 0, err
		}
		if string(dataHeader[:]) == "data" {
			break
		}
		// skip unknown chunk
		if _, err := r.Seek(int64(dataSize), 1); err != nil {
			return nil, 0, err
		}
	}

	numSamples := dataSize / uint32(bitsPerSample/8) / uint32(numChannels)
	samples := make([]float64, numSamples)

	for i := 0; i < int(numSamples); i++ {
		var sum float64
		for ch := 0; ch < int(numChannels); ch++ {
			var raw int16
			if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
				return nil, 0, err
			}
			sum += float64(raw) / 32768.0
		}
		// average channels to mono
		samples[i] = sum / float64(numChannels)
	}

	return samples, int(sampleRate), nil
}
