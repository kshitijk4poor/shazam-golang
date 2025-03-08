package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"

	"github.com/mewkiz/flac"
)

// FLACLoader implements the Loader interface for FLAC files
type FLACLoader struct{}

// NewFLACLoader creates a new FLAC loader
func NewFLACLoader() *FLACLoader {
	return &FLACLoader{}
}

// Load reads and decodes a FLAC file into PCM samples
func (l *FLACLoader) Load(ctx context.Context, reader io.Reader, format AudioFormat) (*AudioData, error) {
	if format != FLAC {
		return nil, fmt.Errorf("unsupported format: %s, expected: %s", format, FLAC)
	}

	// Read all data into memory to create a ReadSeeker
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %w", err)
	}

	// Create a new FLAC decoder
	stream, err := flac.New(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating FLAC decoder: %w", err)
	}
	defer stream.Close()

	// Get audio format information
	info := stream.Info
	sampleRate := int(info.SampleRate)
	channels := int(info.NChannels)
	bitsPerSample := int(info.BitsPerSample)

	// Calculate the maximum value for normalization
	maxValue := math.Pow(2, float64(bitsPerSample-1)) - 1

	// Estimate the number of samples
	numSamples := int(info.NSamples)
	samples := make([]float64, numSamples*channels)

	// Read all audio frames
	sampleIndex := 0
	for {
		frame, err := stream.ParseNext()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing FLAC frame: %w", err)
		}

		// Process each channel
		for i := 0; i < len(frame.Subframes); i++ {
			subframe := frame.Subframes[i]
			for j := 0; j < len(subframe.Samples); j++ {
				if sampleIndex+j*channels+i < len(samples) {
					// Normalize to [-1.0, 1.0]
					samples[sampleIndex+j*channels+i] = float64(subframe.Samples[j]) / maxValue
				}
			}
		}

		// Update sample index
		sampleIndex += len(frame.Subframes[0].Samples) * channels
	}

	// Calculate duration
	duration := float64(numSamples) / float64(sampleRate)

	return &AudioData{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
		Duration:   duration,
	}, nil
}
