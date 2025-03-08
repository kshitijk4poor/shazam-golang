package audio

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"os"

	"github.com/mjibson/go-dsp/fft"
)

// SpectralAnalyzerImpl implements the SpectralAnalyzer interface
type SpectralAnalyzerImpl struct {
	// Configuration parameters
	WindowSize    int     // Size of the window function
	HopSize       int     // Hop size between frames
	SampleRate    int     // Sample rate of the audio
	WindowType    string  // Type of window function (hamming, hann, etc.)
	MinFreq       float64 // Minimum frequency to consider (Hz)
	MaxFreq       float64 // Maximum frequency to consider (Hz)
	LogScaleBase  float64 // Base for logarithmic scaling (0 for linear scale)
	NormalizeSpec bool    // Whether to normalize the spectrogram
	MelScale      bool    // Whether to use mel scale for frequency bins
	NumMelBins    int     // Number of mel bins (if using mel scale)
}

// NewSpectralAnalyzer creates a new spectral analyzer with default settings
func NewSpectralAnalyzer() *SpectralAnalyzerImpl {
	return &SpectralAnalyzerImpl{
		WindowSize:    1024,
		HopSize:       512,
		SampleRate:    44100,
		WindowType:    "hamming",
		MinFreq:       0,
		MaxFreq:       22050, // Nyquist frequency for 44.1kHz
		LogScaleBase:  10.0,
		NormalizeSpec: true,
		MelScale:      false,
		NumMelBins:    128,
	}
}

// ApplyWindow applies a window function to a frame of audio samples
func (s *SpectralAnalyzerImpl) ApplyWindow(frame []float64) []float64 {
	if len(frame) != s.WindowSize {
		// Resize frame if necessary
		newFrame := make([]float64, s.WindowSize)
		copy(newFrame, frame)
		frame = newFrame
	}

	// Create a new windowed frame
	windowedFrame := make([]float64, len(frame))

	switch s.WindowType {
	case "hamming":
		// Hamming window: w(n) = 0.54 - 0.46 * cos(2π * n / (N-1))
		for i := 0; i < len(frame); i++ {
			windowCoeff := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(len(frame)-1))
			windowedFrame[i] = frame[i] * windowCoeff
		}
	case "hann":
		// Hann window: w(n) = 0.5 * (1 - cos(2π * n / (N-1)))
		for i := 0; i < len(frame); i++ {
			windowCoeff := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(frame)-1)))
			windowedFrame[i] = frame[i] * windowCoeff
		}
	case "blackman":
		// Blackman window: w(n) = 0.42 - 0.5 * cos(2π * n / (N-1)) + 0.08 * cos(4π * n / (N-1))
		for i := 0; i < len(frame); i++ {
			n := float64(i)
			N := float64(len(frame) - 1)
			windowCoeff := 0.42 - 0.5*math.Cos(2*math.Pi*n/N) + 0.08*math.Cos(4*math.Pi*n/N)
			windowedFrame[i] = frame[i] * windowCoeff
		}
	case "rectangular":
		// Rectangular window (no windowing)
		copy(windowedFrame, frame)
	default:
		// Default to Hamming window
		for i := 0; i < len(frame); i++ {
			windowCoeff := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(len(frame)-1))
			windowedFrame[i] = frame[i] * windowCoeff
		}
	}

	return windowedFrame
}

// ComputeFFT computes the Fast Fourier Transform of a windowed frame
func (s *SpectralAnalyzerImpl) ComputeFFT(windowedFrame []float64) []complex128 {
	// Convert float64 to complex128
	complexInput := make([]complex128, len(windowedFrame))
	for i, val := range windowedFrame {
		complexInput[i] = complex(val, 0)
	}

	// Compute FFT
	return fft.FFT(complexInput)
}

// ComputeMagnitudeSpectrum computes the magnitude spectrum from FFT results
func (s *SpectralAnalyzerImpl) ComputeMagnitudeSpectrum(fftResult []complex128) []float64 {
	// We only need the first half of the FFT result (up to Nyquist frequency)
	numBins := len(fftResult)/2 + 1
	magnitudeSpectrum := make([]float64, numBins)

	// Compute magnitude for each frequency bin
	for i := 0; i < numBins; i++ {
		// Compute magnitude (absolute value) of the complex FFT result
		magnitudeSpectrum[i] = cmplx.Abs(fftResult[i])
	}

	return magnitudeSpectrum
}

// ComputePowerSpectrum computes the power spectrum from FFT results
func (s *SpectralAnalyzerImpl) ComputePowerSpectrum(fftResult []complex128) []float64 {
	// We only need the first half of the FFT result (up to Nyquist frequency)
	numBins := len(fftResult)/2 + 1
	powerSpectrum := make([]float64, numBins)

	// Compute power for each frequency bin
	for i := 0; i < numBins; i++ {
		// Power is the square of the magnitude
		powerSpectrum[i] = math.Pow(cmplx.Abs(fftResult[i]), 2)
	}

	return powerSpectrum
}

// ApplyLogScale applies logarithmic scaling to a spectrum
func (s *SpectralAnalyzerImpl) ApplyLogScale(spectrum []float64) []float64 {
	if s.LogScaleBase <= 1.0 {
		// No log scaling
		return spectrum
	}

	logSpectrum := make([]float64, len(spectrum))
	for i, val := range spectrum {
		// Add a small value to avoid log(0)
		logSpectrum[i] = math.Log(val+1e-10) / math.Log(s.LogScaleBase)
	}

	return logSpectrum
}

// NormalizeSpectrum normalizes a spectrum to [0, 1] range
func (s *SpectralAnalyzerImpl) NormalizeSpectrum(spectrum []float64) []float64 {
	// Find the maximum value
	maxVal := 0.0
	for _, val := range spectrum {
		if val > maxVal {
			maxVal = val
		}
	}

	// Avoid division by zero
	if maxVal < 1e-10 {
		return spectrum
	}

	// Normalize
	normalizedSpectrum := make([]float64, len(spectrum))
	for i, val := range spectrum {
		normalizedSpectrum[i] = val / maxVal
	}

	return normalizedSpectrum
}

// ComputeSpectrogram converts audio data to a spectrogram
func (s *SpectralAnalyzerImpl) ComputeSpectrogram(data *AudioData, windowSize, hopSize int) (*Spectrogram, error) {
	// Update window and hop size if provided
	if windowSize > 0 {
		s.WindowSize = windowSize
	}
	if hopSize > 0 {
		s.HopSize = hopSize
	}

	// Ensure audio is mono
	if data.Channels != 1 {
		return nil, fmt.Errorf("spectrogram computation requires mono audio, got %d channels", data.Channels)
	}

	// Set sample rate from audio data
	s.SampleRate = data.SampleRate

	// Segment audio into frames
	processor := NewPCMProcessor()
	processor.FrameSize = s.WindowSize
	processor.HopSize = s.HopSize
	frames, err := processor.SegmentIntoFrames(data)
	if err != nil {
		return nil, fmt.Errorf("failed to segment audio into frames: %w", err)
	}

	// Compute spectrogram
	numFrames := len(frames)
	numBins := s.WindowSize/2 + 1
	spectrogramData := make([][]float64, numFrames)

	// Process each frame
	for i, frame := range frames {
		// Apply window function
		windowedFrame := s.ApplyWindow(frame)

		// Compute FFT
		fftResult := s.ComputeFFT(windowedFrame)

		// Compute power spectrum
		powerSpectrum := s.ComputePowerSpectrum(fftResult)

		// Apply log scaling if needed
		if s.LogScaleBase > 1.0 {
			powerSpectrum = s.ApplyLogScale(powerSpectrum)
		}

		// Normalize if needed
		if s.NormalizeSpec {
			powerSpectrum = s.NormalizeSpectrum(powerSpectrum)
		}

		// Store in spectrogram
		spectrogramData[i] = powerSpectrum
	}

	// Calculate time and frequency points
	timePoints := make([]float64, numFrames)
	for i := 0; i < numFrames; i++ {
		timePoints[i] = float64(i*s.HopSize) / float64(s.SampleRate)
	}

	freqPoints := make([]float64, numBins)
	for i := 0; i < numBins; i++ {
		freqPoints[i] = float64(i) * float64(s.SampleRate) / float64(s.WindowSize)
	}

	// Create and return spectrogram
	return &Spectrogram{
		Data:       spectrogramData,
		FreqBins:   numBins,
		TimeBins:   numFrames,
		TimePoints: timePoints,
		FreqPoints: freqPoints,
	}, nil
}

// SaveSpectrogramImage saves a spectrogram as an image
func (s *SpectralAnalyzerImpl) SaveSpectrogramImage(spectrogram *Spectrogram, filePath string) error {
	// Check if spectrogram is valid
	if spectrogram == nil || len(spectrogram.Data) == 0 || len(spectrogram.Data[0]) == 0 {
		return fmt.Errorf("invalid spectrogram data")
	}

	// Create a new image
	width := spectrogram.TimeBins
	height := spectrogram.FreqBins
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Define a color palette (from blue to red)
	getColor := func(value float64) color.RGBA {
		// Map value from [0,1] to a color
		// Blue (0,0,255) -> Cyan (0,255,255) -> Green (0,255,0) -> Yellow (255,255,0) -> Red (255,0,0)
		r, g, b := 0, 0, 0

		if value < 0.25 {
			// Blue to Cyan
			v := value * 4
			b = 255
			g = int(v * 255)
		} else if value < 0.5 {
			// Cyan to Green
			v := (value - 0.25) * 4
			g = 255
			b = 255 - int(v*255)
		} else if value < 0.75 {
			// Green to Yellow
			v := (value - 0.5) * 4
			g = 255
			r = int(v * 255)
		} else {
			// Yellow to Red
			v := (value - 0.75) * 4
			r = 255
			g = 255 - int(v*255)
		}

		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}

	// Fill the image with spectrogram data
	// Note: Frequency axis is typically displayed with low frequencies at the bottom
	for t := 0; t < width; t++ {
		for f := 0; f < height; f++ {
			// Get the spectrogram value (ensure it's in [0,1] range)
			value := spectrogram.Data[t][height-f-1] // Invert frequency axis
			if value < 0 {
				value = 0
			}
			if value > 1 {
				value = 1
			}

			// Set the pixel color
			img.Set(t, f, getColor(value))
		}
	}

	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	// Encode and save the image
	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}
