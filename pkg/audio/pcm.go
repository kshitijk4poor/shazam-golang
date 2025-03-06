package audio

import (
	"fmt"
	"math"
)

// PCMProcessor handles PCM audio processing operations
type PCMProcessor struct {
	// Configuration parameters
	TargetSampleRate int
	FrameSize        int
	HopSize          int
}

// NewPCMProcessor creates a new PCM processor with default settings
func NewPCMProcessor() *PCMProcessor {
	return &PCMProcessor{
		TargetSampleRate: 44100, // Default target sample rate
		FrameSize:        1024,  // Default frame size
		HopSize:          512,   // Default hop size (50% overlap)
	}
}

// ConvertToMono converts stereo audio to mono by averaging channels
func (p *PCMProcessor) ConvertToMono(data *AudioData) (*AudioData, error) {
	if data.Channels == 1 {
		// Already mono
		return data, nil
	}

	if data.Channels != 2 {
		return nil, fmt.Errorf("unsupported number of channels: %d, only mono and stereo are supported", data.Channels)
	}

	// Create a new mono audio data
	monoSamples := make([]float64, len(data.Samples)/2)
	for i := 0; i < len(monoSamples); i++ {
		// Average the left and right channels
		monoSamples[i] = (data.Samples[i*2] + data.Samples[i*2+1]) / 2.0
	}

	return &AudioData{
		Samples:    monoSamples,
		SampleRate: data.SampleRate,
		Channels:   1,
		Duration:   data.Duration,
	}, nil
}

// Normalize adjusts audio amplitude to a standard level
func (p *PCMProcessor) Normalize(data *AudioData) (*AudioData, error) {
	if len(data.Samples) == 0 {
		return data, nil
	}

	// Find the maximum absolute value
	maxAbs := 0.0
	for _, sample := range data.Samples {
		abs := math.Abs(sample)
		if abs > maxAbs {
			maxAbs = abs
		}
	}

	// If the maximum is already close to 1.0 or is zero, no need to normalize
	if maxAbs < 0.001 || math.Abs(maxAbs-1.0) < 0.001 {
		return data, nil
	}

	// Create a new normalized audio data
	normalizedSamples := make([]float64, len(data.Samples))
	for i, sample := range data.Samples {
		normalizedSamples[i] = sample / maxAbs
	}

	return &AudioData{
		Samples:    normalizedSamples,
		SampleRate: data.SampleRate,
		Channels:   data.Channels,
		Duration:   data.Duration,
	}, nil
}

// ResampleTo resamples audio to target sample rate using linear interpolation
func (p *PCMProcessor) ResampleTo(data *AudioData, targetSampleRate int) (*AudioData, error) {
	if data.SampleRate == targetSampleRate {
		// Already at target sample rate
		return data, nil
	}

	// Calculate the ratio between the original and target sample rates
	ratio := float64(targetSampleRate) / float64(data.SampleRate)

	// Calculate the number of frames in the original and new audio
	origFrames := len(data.Samples) / data.Channels
	newFrames := int(float64(origFrames) * ratio)

	// Create a new resampled audio data
	resampledSamples := make([]float64, newFrames*data.Channels)

	// Resample each channel separately
	for ch := 0; ch < data.Channels; ch++ {
		for i := 0; i < newFrames; i++ {
			// Calculate the position in the original samples
			origPos := float64(i) / ratio

			// Get the indices of the two nearest samples
			idx1 := int(math.Floor(origPos))
			idx2 := idx1 + 1

			// Calculate the fractional part for interpolation
			frac := origPos - float64(idx1)

			// Handle boundary conditions
			if idx1 >= origFrames {
				idx1 = origFrames - 1
			}
			if idx2 >= origFrames {
				idx2 = origFrames - 1
			}

			// Get the original samples for this channel
			sample1 := data.Samples[idx1*data.Channels+ch]
			sample2 := data.Samples[idx2*data.Channels+ch]

			// Linear interpolation
			resampledSamples[i*data.Channels+ch] = sample1*(1-frac) + sample2*frac
		}
	}

	// Calculate new duration
	newDuration := float64(newFrames) / float64(targetSampleRate)

	return &AudioData{
		Samples:    resampledSamples,
		SampleRate: targetSampleRate,
		Channels:   data.Channels,
		Duration:   newDuration,
	}, nil
}

// SegmentIntoFrames divides audio into overlapping frames
func (p *PCMProcessor) SegmentIntoFrames(data *AudioData) ([][]float64, error) {
	if len(data.Samples) < p.FrameSize {
		return nil, fmt.Errorf("audio data too short for frame size %d", p.FrameSize)
	}

	// Calculate the number of frames
	numFrames := 1 + (len(data.Samples)-p.FrameSize)/p.HopSize

	// Create frames array
	frames := make([][]float64, numFrames)

	// Extract frames
	for i := 0; i < numFrames; i++ {
		startIdx := i * p.HopSize
		endIdx := startIdx + p.FrameSize

		// Handle the last frame if it would go beyond the sample array
		if endIdx > len(data.Samples) {
			endIdx = len(data.Samples)
		}

		// Create a new frame
		frame := make([]float64, p.FrameSize)

		// Copy samples to the frame
		copy(frame, data.Samples[startIdx:endIdx])

		// Zero-pad if necessary
		if endIdx-startIdx < p.FrameSize {
			for j := endIdx - startIdx; j < p.FrameSize; j++ {
				frame[j] = 0.0
			}
		}

		frames[i] = frame
	}

	return frames, nil
}
