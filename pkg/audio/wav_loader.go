package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"

	"github.com/go-audio/wav"
)

// WAVLoader implements the Loader interface for WAV files
type WAVLoader struct{}

// NewWAVLoader creates a new WAV loader
func NewWAVLoader() *WAVLoader {
	return &WAVLoader{}
}

// Load reads and decodes a WAV file into PCM samples
func (l *WAVLoader) Load(ctx context.Context, reader io.Reader, format AudioFormat) (*AudioData, error) {
	if format != WAV {
		return nil, fmt.Errorf("unsupported format: %s, expected: %s", format, WAV)
	}

	// Read all data into memory to create a ReadSeeker
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %w", err)
	}

	// Create a new WAV decoder with a ReadSeeker
	decoder := wav.NewDecoder(bytes.NewReader(data))
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file")
	}

	// Get audio format information
	audioFormat := decoder.Format()
	sampleRate := int(audioFormat.SampleRate)
	channels := int(audioFormat.NumChannels)
	bitDepth := int(decoder.BitDepth)

	// Read all samples
	decoder.FwdToPCM()

	// Get all audio samples as integers
	samplesInt, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("error reading PCM data: %w", err)
	}

	// Calculate duration
	numFrames := len(samplesInt.Data) / channels
	duration := float64(numFrames) / float64(sampleRate)

	// Convert int samples to float64 samples (normalized to [-1.0, 1.0])
	maxValue := math.Pow(2, float64(bitDepth-1))
	samples := make([]float64, len(samplesInt.Data))
	for i, sample := range samplesInt.Data {
		samples[i] = float64(sample) / maxValue
	}

	return &AudioData{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
		Duration:   duration,
	}, nil
}
