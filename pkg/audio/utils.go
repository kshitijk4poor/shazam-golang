package audio

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// AudioUtils provides utility functions for audio processing
type AudioUtils struct {
	Loader    Loader
	Processor Processor
}

// NewAudioUtils creates a new AudioUtils instance
func NewAudioUtils() *AudioUtils {
	return &AudioUtils{
		Loader:    NewWAVLoader(),
		Processor: NewPCMProcessor(),
	}
}

// LoadAndPreprocess loads an audio file and applies preprocessing steps
func (u *AudioUtils) LoadAndPreprocess(filePath string, targetSampleRate int, convertToMono bool) (*AudioData, error) {
	// Determine the audio format from the file extension
	format, err := getAudioFormatFromPath(filePath)
	if err != nil {
		return nil, err
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Load the audio data
	ctx := context.Background()
	audioData, err := u.Loader.Load(ctx, file, format)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio: %w", err)
	}

	// Convert to mono if requested
	if convertToMono && audioData.Channels > 1 {
		audioData, err = u.Processor.ConvertToMono(audioData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to mono: %w", err)
		}
	}

	// Resample if needed
	if audioData.SampleRate != targetSampleRate {
		audioData, err = u.Processor.ResampleTo(audioData, targetSampleRate)
		if err != nil {
			return nil, fmt.Errorf("failed to resample audio: %w", err)
		}
	}

	// Normalize the audio
	audioData, err = u.Processor.Normalize(audioData)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize audio: %w", err)
	}

	return audioData, nil
}

// CalculateRMS calculates the Root Mean Square (RMS) of audio samples
func (u *AudioUtils) CalculateRMS(samples []float64) float64 {
	if len(samples) == 0 {
		return 0.0
	}

	sumSquares := 0.0
	for _, sample := range samples {
		sumSquares += sample * sample
	}

	return math.Sqrt(sumSquares / float64(len(samples)))
}

// CalculateEnergy calculates the energy of audio samples
func (u *AudioUtils) CalculateEnergy(samples []float64) float64 {
	if len(samples) == 0 {
		return 0.0
	}

	energy := 0.0
	for _, sample := range samples {
		energy += sample * sample
	}

	return energy
}

// CalculateZeroCrossingRate calculates the zero-crossing rate of audio samples
func (u *AudioUtils) CalculateZeroCrossingRate(samples []float64) float64 {
	if len(samples) <= 1 {
		return 0.0
	}

	crossings := 0
	for i := 1; i < len(samples); i++ {
		if (samples[i-1] >= 0 && samples[i] < 0) || (samples[i-1] < 0 && samples[i] >= 0) {
			crossings++
		}
	}

	return float64(crossings) / float64(len(samples)-1)
}

// getAudioFormatFromPath determines the audio format from the file extension
func getAudioFormatFromPath(path string) (AudioFormat, error) {
	ext := filepath.Ext(path)
	if ext == "" {
		return "", fmt.Errorf("file has no extension")
	}

	// Remove the leading dot and convert to lowercase
	ext = ext[1:]
	switch ext {
	case "wav":
		return WAV, nil
	case "mp3":
		return MP3, nil
	case "flac":
		return FLAC, nil
	default:
		return "", fmt.Errorf("unsupported audio format: %s", ext)
	}
}
