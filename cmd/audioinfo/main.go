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
	targetSampleRate := flag.Int("samplerate", 44100, "Target sample rate for resampling")
	convertToMono := flag.Bool("mono", false, "Convert audio to mono")
	flag.Parse()

	// Check if a file path was provided
	if flag.NArg() < 1 {
		fmt.Println("Usage: audioinfo [options] <audio-file>")
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

	// Calculate and display audio statistics
	fmt.Println("\nAudio Statistics:")
	fmt.Printf("RMS:                  %.6f\n", utils.CalculateRMS(audioData.Samples))
	fmt.Printf("Energy:               %.6f\n", utils.CalculateEnergy(audioData.Samples))
	fmt.Printf("Zero Crossing Rate:   %.6f\n", utils.CalculateZeroCrossingRate(audioData.Samples))

	// Create a spectral analyzer
	analyzer := audio.NewSpectralAnalyzer()

	// Compute spectrogram
	fmt.Println("\nComputing spectrogram...")
	spectrogram, err := analyzer.ComputeSpectrogram(audioData, 1024, 512)
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

	fmt.Println("\nAudio processing completed successfully.")
}
