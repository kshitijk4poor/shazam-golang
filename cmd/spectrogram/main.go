package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kshitijk4poor/shazam-golang/pkg/audio"
)

func main() {
	// Parse command-line arguments
	windowSize := flag.Int("window", 1024, "Window size for FFT")
	hopSize := flag.Int("hop", 512, "Hop size between frames")
	windowType := flag.String("window-type", "hamming", "Window function type (hamming, hann, blackman, rectangular)")
	logScale := flag.Bool("log", true, "Apply logarithmic scaling")
	normalize := flag.Bool("normalize", true, "Normalize spectrogram values")
	outputDir := flag.String("output", ".", "Output directory for spectrogram images")
	targetSampleRate := flag.Int("samplerate", 44100, "Target sample rate for resampling")
	convertToMono := flag.Bool("mono", true, "Convert audio to mono")
	flag.Parse()

	// Check if a file path was provided
	if flag.NArg() < 1 {
		fmt.Println("Usage: spectrogram [options] <audio-file>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Get the file path
	filePath := flag.Arg(0)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File '%s' does not exist\n", filePath)
		os.Exit(1)
	}

	// Create an audio utils instance
	utils := audio.NewAudioUtils()

	// Load and preprocess the audio file
	fmt.Printf("Loading audio file: %s\n", filePath)
	audioData, err := utils.LoadAndPreprocess(filePath, *targetSampleRate, *convertToMono)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display audio information
	fmt.Println("\nAudio Information:")
	fmt.Printf("File:        %s\n", filepath.Base(filePath))
	fmt.Printf("Format:      %s\n", filepath.Ext(filePath)[1:])
	fmt.Printf("Channels:    %d\n", audioData.Channels)
	fmt.Printf("Sample Rate: %d Hz\n", audioData.SampleRate)
	fmt.Printf("Duration:    %.2f seconds\n", audioData.Duration)
	fmt.Printf("Samples:     %d\n", len(audioData.Samples))

	// Create a spectral analyzer
	analyzer := audio.NewSpectralAnalyzer()
	analyzer.WindowSize = *windowSize
	analyzer.HopSize = *hopSize
	analyzer.WindowType = *windowType
	analyzer.LogScaleBase = 0.0
	if *logScale {
		analyzer.LogScaleBase = 10.0
	}
	analyzer.NormalizeSpec = *normalize

	// Compute spectrogram
	fmt.Println("\nComputing spectrogram...")
	spectrogram, err := analyzer.ComputeSpectrogram(audioData, *windowSize, *hopSize)
	if err != nil {
		fmt.Printf("Error computing spectrogram: %v\n", err)
		os.Exit(1)
	}

	// Display spectrogram information
	fmt.Println("Spectrogram Information:")
	fmt.Printf("Time Bins:  %d\n", spectrogram.TimeBins)
	fmt.Printf("Freq Bins:  %d\n", spectrogram.FreqBins)
	fmt.Printf("Time Range: %.2f - %.2f seconds\n", spectrogram.TimePoints[0], spectrogram.TimePoints[len(spectrogram.TimePoints)-1])
	fmt.Printf("Freq Range: %.2f - %.2f Hz\n", spectrogram.FreqPoints[0], spectrogram.FreqPoints[len(spectrogram.FreqPoints)-1])

	// Create output directory if it doesn't exist
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(*outputDir, 0755)
		if err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate output file path
	baseFileName := filepath.Base(filePath)
	baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))]
	outputPath := filepath.Join(*outputDir, fmt.Sprintf("%s_spectrogram.png", baseFileName))

	// Save spectrogram as an image
	fmt.Printf("\nSaving spectrogram to: %s\n", outputPath)
	err = analyzer.SaveSpectrogramImage(spectrogram, outputPath)
	if err != nil {
		fmt.Printf("Error saving spectrogram image: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Spectrogram generation completed successfully.")
}
