package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
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
		offset := i * 2
		raw := int16(binary.LittleEndian.Uint16(b[offset : offset+2]))
		samples[i] = float64(raw) / 32768.0
	}

	return samples, 0, nil
}

// DecodeWAVToFloat64 decodes a WAV file (16, 24, or 32-bit PCM) into float64 samples in [-1.0, +1.0].
// Mono output is returned by averaging all channels.
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

	// --- scan for "fmt " chunk ---
	var audioFormat uint16
	var numChannels uint16
	var sampleRate uint32
	var bitsPerSample uint16

	for {
		var chunkHeader [4]byte
		var chunkSize uint32

		if err := binary.Read(r, binary.LittleEndian, &chunkHeader); err != nil {
			return nil, 0, err
		}
		if err := binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
			return nil, 0, err
		}

		switch string(chunkHeader[:]) {
		case "fmt ":
			// read fmt chunk
			if err := binary.Read(r, binary.LittleEndian, &audioFormat); err != nil {
				return nil, 0, err
			}
			if err := binary.Read(r, binary.LittleEndian, &numChannels); err != nil {
				return nil, 0, err
			}
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
			if err := binary.Read(r, binary.LittleEndian, &bitsPerSample); err != nil {
				return nil, 0, err
			}
			if audioFormat != 1 {
				return nil, 0, errors.New("only PCM format supported")
			}
			if bitsPerSample != 16 && bitsPerSample != 24 && bitsPerSample != 32 {
				return nil, 0, errors.New("only 16, 24, or 32-bit WAV supported")
			}
			// skip extra fmt bytes
			if extra := int64(chunkSize) - 16; extra > 0 {
				if _, err := r.Seek(extra, io.SeekCurrent); err != nil {
					return nil, 0, err
				}
			}
			goto foundFmt
		default:
			// skip unknown chunk
			if _, err := r.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, 0, err
			}
		}
	}
foundFmt:

	// --- scan for "data" chunk ---
	var dataSize uint32
	for {
		var chunkHeader [4]byte
		if err := binary.Read(r, binary.LittleEndian, &chunkHeader); err != nil {
			return nil, 0, err
		}
		if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
			return nil, 0, err
		}
		if string(chunkHeader[:]) == "data" {
			break
		}
		if _, err := r.Seek(int64(dataSize), io.SeekCurrent); err != nil {
			return nil, 0, err
		}
	}

	numSamples := dataSize / uint32(bitsPerSample/8) / uint32(numChannels)
	samples := make([]float64, numSamples)

	for i := 0; i < int(numSamples); i++ {
		var sum float64
		for ch := 0; ch < int(numChannels); ch++ {
			var val float64
			switch bitsPerSample {
			case 16:
				var raw int16
				if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
					return nil, 0, err
				}
				val = float64(raw) / 32768.0
			case 24:
				buf := make([]byte, 3)
				if _, err := r.Read(buf); err != nil {
					return nil, 0, err
				}
				raw := int32(buf[0]) | int32(buf[1])<<8 | int32(buf[2])<<16
				if raw&0x800000 != 0 {
					raw |= ^0xffffff
				}
				val = float64(raw) / 8388608.0
			case 32:
				var raw int32
				if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
					return nil, 0, err
				}
				val = float64(raw) / 2147483648.0
			}
			sum += val
		}
		samples[i] = sum / float64(numChannels)
	}

	return samples, int(sampleRate), nil
}
