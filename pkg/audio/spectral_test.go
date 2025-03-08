package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowFunctions(t *testing.T) {
	// Create a spectral analyzer
	analyzer := NewSpectralAnalyzer()

	// Create a test frame (all ones)
	frame := make([]float64, analyzer.WindowSize)
	for i := range frame {
		frame[i] = 1.0
	}

	// Test Hamming window
	analyzer.WindowType = "hamming"
	hammingFrame := analyzer.ApplyWindow(frame)

	// Check that the window is applied correctly
	// Hamming window should have values between 0.08 and 1.0
	for i, val := range hammingFrame {
		if val < 0.08 || val > 1.0 {
			t.Errorf("Hamming window value out of range at index %d: %f", i, val)
		}
	}

	// Check that the window is symmetric
	for i := 0; i < analyzer.WindowSize/2; i++ {
		if math.Abs(hammingFrame[i]-hammingFrame[analyzer.WindowSize-1-i]) > 1e-10 {
			t.Errorf("Hamming window not symmetric at indices %d and %d: %f vs %f",
				i, analyzer.WindowSize-1-i, hammingFrame[i], hammingFrame[analyzer.WindowSize-1-i])
		}
	}

	// Test Hann window
	analyzer.WindowType = "hann"
	hannFrame := analyzer.ApplyWindow(frame)

	// Check that the window is applied correctly
	// Hann window should have values between 0.0 and 1.0
	for i, val := range hannFrame {
		if val < 0.0 || val > 1.0 {
			t.Errorf("Hann window value out of range at index %d: %f", i, val)
		}
	}
}

func TestFFT(t *testing.T) {
	// Create a spectral analyzer
	analyzer := NewSpectralAnalyzer()

	// Create a sine wave at 1000 Hz
	sampleRate := 44100
	frequency := 1000.0
	duration := 1.0
	numSamples := int(duration * float64(sampleRate))

	samples := make([]float64, numSamples)
	for i := range samples {
		time := float64(i) / float64(sampleRate)
		samples[i] = math.Sin(2 * math.Pi * frequency * time)
	}

	// Create audio data
	audioData := &AudioData{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   1,
		Duration:   duration,
	}

	// Compute spectrogram
	spectrogram, err := analyzer.ComputeSpectrogram(audioData, 1024, 512)
	if err != nil {
		t.Fatalf("Failed to compute spectrogram: %v", err)
	}

	// Check spectrogram dimensions
	expectedTimeBins := 1 + (numSamples-1024)/512
	if spectrogram.TimeBins != expectedTimeBins {
		t.Errorf("Expected %d time bins, got %d", expectedTimeBins, spectrogram.TimeBins)
	}

	if spectrogram.FreqBins != 513 { // 1024/2 + 1
		t.Errorf("Expected 513 frequency bins, got %d", spectrogram.FreqBins)
	}

	// Find the peak frequency bin
	peakBin := 0
	peakValue := 0.0

	// Use the middle time frame
	middleFrame := spectrogram.Data[spectrogram.TimeBins/2]

	for i, val := range middleFrame {
		if val > peakValue {
			peakValue = val
			peakBin = i
		}
	}

	// Calculate the frequency of the peak bin
	peakFreq := float64(peakBin) * float64(sampleRate) / 1024.0

	// Check that the peak frequency is close to 1000 Hz
	if math.Abs(peakFreq-frequency) > 50.0 {
		t.Errorf("Expected peak frequency around %f Hz, got %f Hz", frequency, peakFreq)
	}
}

func TestLogScale(t *testing.T) {
	// Create a spectral analyzer
	analyzer := NewSpectralAnalyzer()

	// Create a test spectrum
	spectrum := []float64{1.0, 10.0, 100.0, 1000.0}

	// Apply log scale (base 10)
	analyzer.LogScaleBase = 10.0
	logSpectrum := analyzer.ApplyLogScale(spectrum)

	// Check that the log scale is applied correctly
	expectedLogSpectrum := []float64{0.0, 1.0, 2.0, 3.0}
	for i, val := range logSpectrum {
		if math.Abs(val-expectedLogSpectrum[i]) > 1e-10 {
			t.Errorf("Expected log value %f at index %d, got %f", expectedLogSpectrum[i], i, val)
		}
	}
}

func TestNormalization(t *testing.T) {
	// Create a spectral analyzer
	analyzer := NewSpectralAnalyzer()

	// Create a test spectrum
	spectrum := []float64{1.0, 2.0, 4.0, 8.0}

	// Normalize the spectrum
	normalizedSpectrum := analyzer.NormalizeSpectrum(spectrum)

	// Check that the spectrum is normalized correctly
	expectedNormalizedSpectrum := []float64{0.125, 0.25, 0.5, 1.0}
	for i, val := range normalizedSpectrum {
		if math.Abs(val-expectedNormalizedSpectrum[i]) > 1e-10 {
			t.Errorf("Expected normalized value %f at index %d, got %f",
				expectedNormalizedSpectrum[i], i, val)
		}
	}
}

func TestSpectrogramImage(t *testing.T) {
	// Create a spectral analyzer
	analyzer := NewSpectralAnalyzer()

	// Create a test spectrogram
	timeBins := 100
	freqBins := 50
	spectrogram := &Spectrogram{
		Data:       make([][]float64, timeBins),
		FreqBins:   freqBins,
		TimeBins:   timeBins,
		TimePoints: make([]float64, timeBins),
		FreqPoints: make([]float64, freqBins),
	}

	// Fill the spectrogram with test data
	for i := 0; i < timeBins; i++ {
		spectrogram.Data[i] = make([]float64, freqBins)
		spectrogram.TimePoints[i] = float64(i) / 100.0

		for j := 0; j < freqBins; j++ {
			spectrogram.FreqPoints[j] = float64(j) * 100.0

			// Create a pattern: diagonal lines
			if (i+j)%10 < 5 {
				spectrogram.Data[i][j] = 0.8
			} else {
				spectrogram.Data[i][j] = 0.2
			}
		}
	}

	// Create a temporary directory for the test
	tempDir := filepath.Join(os.TempDir(), "spectrogram_test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save the spectrogram as an image
	imagePath := filepath.Join(tempDir, "spectrogram.png")
	err = analyzer.SaveSpectrogramImage(spectrogram, imagePath)
	if err != nil {
		t.Fatalf("Failed to save spectrogram image: %v", err)
	}

	// Check that the image file exists
	_, err = os.Stat(imagePath)
	if os.IsNotExist(err) {
		t.Errorf("Spectrogram image file was not created")
	}
}
