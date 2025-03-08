package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3"
)

// MP3Loader implements the Loader interface for MP3 files
type MP3Loader struct{}

// NewMP3Loader creates a new MP3 loader
func NewMP3Loader() *MP3Loader {
	return &MP3Loader{}
}

// Load reads and decodes an MP3 file into PCM samples
func (l *MP3Loader) Load(ctx context.Context, reader io.Reader, format AudioFormat) (*AudioData, error) {
	if format != MP3 {
		return nil, fmt.Errorf("unsupported format: %s, expected: %s", format, MP3)
	}

	// Read all data into memory to create a ReadSeeker
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %w", err)
	}

	// Create a new MP3 decoder
	decoder, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating MP3 decoder: %w", err)
	}

	// Get audio format information
	sampleRate := decoder.SampleRate()
	channels := 2 // MP3 is typically stereo

	// Calculate the number of samples
	numSamples := int(decoder.Length() / 4) // 4 bytes per sample (2 bytes per channel, 2 channels)

	// Create a buffer to hold the PCM data
	pcmData := make([]byte, decoder.Length())

	// Read all PCM data
	_, err = io.ReadFull(decoder, pcmData)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading PCM data: %w", err)
	}

	// Convert PCM data to float64 samples
	samples := make([]float64, numSamples*channels)
	for i := 0; i < numSamples; i++ {
		for ch := 0; ch < channels; ch++ {
			// Extract the 16-bit sample for this channel
			sampleIdx := i*4 + ch*2
			if sampleIdx+1 < len(pcmData) {
				// Convert from little-endian 16-bit to int16
				sample := int16(pcmData[sampleIdx]) | (int16(pcmData[sampleIdx+1]) << 8)
				// Normalize to [-1.0, 1.0]
				samples[i*channels+ch] = float64(sample) / 32768.0
			}
		}
	}

	// Calculate duration
	duration := float64(numSamples) / float64(sampleRate)

	return &AudioData{
		Samples:    samples,
		SampleRate: int(sampleRate),
		Channels:   channels,
		Duration:   duration,
	}, nil
}
