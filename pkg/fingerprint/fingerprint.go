package fingerprint

import (
	"github.com/kshitijk4poor/shazam-golang/pkg/audio"
)

// Peak represents a spectral peak in time-frequency domain
type Peak struct {
	Frequency float64
	Time      float64
	Amplitude float64
}

// Vector represents a fingerprint vector in high-dimensional space
type Vector struct {
	Data    []float32 // Vector components
	TimeRef float64   // Reference time in audio
	TrackID string    // Associated track identifier
}

// Generator handles creation of fingerprint vectors from audio
type Generator interface {
	// ExtractPeaks finds significant peaks in spectrogram
	ExtractPeaks(spec *audio.Spectrogram) ([]Peak, error)

	// GenerateVector creates fingerprint vector from peaks
	GenerateVector(peaks []Peak) (*Vector, error)

	// Process handles complete fingerprint generation from audio
	Process(data *audio.AudioData) ([]*Vector, error)
}

// Config holds fingerprint generation parameters
type Config struct {
	VectorDim     int     // Dimensionality of fingerprint vectors
	PeakThreshold float64 // Minimum amplitude for peak detection
	NeighborSize  int     // Size of neighborhood for peak finding
	VectorsPerSec float64 // Number of vectors to generate per second
}
