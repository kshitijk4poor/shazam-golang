package audio

import (
	"context"
	"io"
)

type AudioFormat string

const (
	WAV  AudioFormat = "wav"
	MP3  AudioFormat = "mp3"
	FLAC AudioFormat = "flac"
)

// AudioData represents processed audio samples
type AudioData struct {
	Samples    []float64
	SampleRate int
	Channels   int
	Duration   float64
}

// Loader handles loading and decoding audio files
type Loader interface {
	// Load reads and decodes audio file into PCM samples
	Load(ctx context.Context, reader io.Reader, format AudioFormat) (*AudioData, error)
}

// Processor handles audio signal processing operations
type Processor interface {
	// Normalize adjusts audio amplitude to a standard level
	Normalize(data *AudioData) (*AudioData, error)

	// ConvertToMono converts stereo to mono by averaging channels
	ConvertToMono(data *AudioData) (*AudioData, error)

	// ResampleTo resamples audio to target sample rate
	ResampleTo(data *AudioData, targetSampleRate int) (*AudioData, error)
}

// Spectrogram represents the time-frequency representation of audio
type Spectrogram struct {
	Data       [][]float64 // Power spectrum over time
	FreqBins   int         // Number of frequency bins
	TimeBins   int         // Number of time bins
	TimePoints []float64   // Time points for each column
	FreqPoints []float64   // Frequency points for each row
}

// SpectralAnalyzer handles conversion of audio to spectral domain
type SpectralAnalyzer interface {
	// ComputeSpectrogram converts audio data to spectrogram
	ComputeSpectrogram(data *AudioData, windowSize, hopSize int) (*Spectrogram, error)
}
