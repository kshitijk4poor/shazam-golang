package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"testing"
)

// createTestWAVData creates a simple WAV file with a sine wave
func createTestWAVData(sampleRate, numSamples, channels int) []byte {
	// Create a buffer to write the WAV file
	buf := bytes.NewBuffer(nil)

	// Write the WAV header
	// RIFF header
	buf.WriteString("RIFF")
	// File size (will be filled in later)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	// WAVE header
	buf.WriteString("WAVE")
	// fmt chunk
	buf.WriteString("fmt ")
	// Chunk size
	binary.Write(buf, binary.LittleEndian, uint32(16))
	// Audio format (1 = PCM)
	binary.Write(buf, binary.LittleEndian, uint16(1))
	// Number of channels
	binary.Write(buf, binary.LittleEndian, uint16(channels))
	// Sample rate
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	// Byte rate
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*channels*2))
	// Block align
	binary.Write(buf, binary.LittleEndian, uint16(channels*2))
	// Bits per sample
	binary.Write(buf, binary.LittleEndian, uint16(16))
	// Data chunk
	buf.WriteString("data")
	// Data size
	binary.Write(buf, binary.LittleEndian, uint32(numSamples*channels*2))

	// Generate a sine wave
	for i := 0; i < numSamples; i++ {
		for c := 0; c < channels; c++ {
			// Generate a sine wave with frequency 440Hz
			t := float64(i) / float64(sampleRate)
			amplitude := 0.5 * math.Sin(2*math.Pi*440*t)

			// Convert to int16
			sample := int16(amplitude * 32767)

			// Write the sample
			binary.Write(buf, binary.LittleEndian, sample)
		}
	}

	// Update the file size
	data := buf.Bytes()
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))

	return data
}

func TestWAVLoader(t *testing.T) {
	// Create test WAV data
	sampleRate := 44100
	numSamples := 44100 // 1 second
	channels := 2
	wavData := createTestWAVData(sampleRate, numSamples, channels)

	// Create a WAV loader
	loader := NewWAVLoader()

	// Load the WAV data
	ctx := context.Background()
	audioData, err := loader.Load(ctx, bytes.NewReader(wavData), WAV)
	if err != nil {
		t.Fatalf("Failed to load WAV data: %v", err)
	}

	// Check the audio data
	if audioData.SampleRate != sampleRate {
		t.Errorf("Expected sample rate %d, got %d", sampleRate, audioData.SampleRate)
	}
	if audioData.Channels != channels {
		t.Errorf("Expected %d channels, got %d", channels, audioData.Channels)
	}
	if math.Abs(audioData.Duration-1.0) > 0.01 {
		t.Errorf("Expected duration 1.0, got %f", audioData.Duration)
	}
	if len(audioData.Samples) != numSamples*channels {
		t.Errorf("Expected %d samples, got %d", numSamples*channels, len(audioData.Samples))
	}
}

// TestMP3Loader tests the MP3 loader with a mock MP3 file
// Note: This is a basic test that checks if the loader can be created
// A full test would require a real MP3 file
func TestMP3Loader(t *testing.T) {
	// Create an MP3 loader
	loader := NewMP3Loader()

	// Check that the loader was created
	if loader == nil {
		t.Fatalf("Failed to create MP3 loader")
	}
}

// TestFLACLoader tests the FLAC loader with a mock FLAC file
// Note: This is a basic test that checks if the loader can be created
// A full test would require a real FLAC file
func TestFLACLoader(t *testing.T) {
	// Create a FLAC loader
	loader := NewFLACLoader()

	// Check that the loader was created
	if loader == nil {
		t.Fatalf("Failed to create FLAC loader")
	}
}

func TestPCMProcessor(t *testing.T) {
	// Create a PCM processor
	processor := NewPCMProcessor()

	// Create test audio data
	audioData := &AudioData{
		Samples:    make([]float64, 44100*2), // 1 second of stereo audio
		SampleRate: 44100,
		Channels:   2,
		Duration:   1.0,
	}

	// Generate a sine wave
	for i := 0; i < 44100; i++ {
		t := float64(i) / 44100.0
		audioData.Samples[i*2] = 0.5 * math.Sin(2*math.Pi*440*t)   // Left channel
		audioData.Samples[i*2+1] = 0.5 * math.Sin(2*math.Pi*880*t) // Right channel
	}

	// Test ConvertToMono
	monoData, err := processor.ConvertToMono(audioData)
	if err != nil {
		t.Fatalf("Failed to convert to mono: %v", err)
	}
	if monoData.Channels != 1 {
		t.Errorf("Expected 1 channel, got %d", monoData.Channels)
	}
	if len(monoData.Samples) != 44100 {
		t.Errorf("Expected 44100 samples, got %d", len(monoData.Samples))
	}

	// Test ResampleTo
	resampledData, err := processor.ResampleTo(audioData, 22050)
	if err != nil {
		t.Fatalf("Failed to resample: %v", err)
	}
	if resampledData.SampleRate != 22050 {
		t.Errorf("Expected sample rate 22050, got %d", resampledData.SampleRate)
	}
	if math.Abs(resampledData.Duration-1.0) > 0.01 {
		t.Errorf("Expected duration 1.0, got %f", resampledData.Duration)
	}

	// Test Normalize
	// First, create data with low amplitude
	lowAmpData := &AudioData{
		Samples:    make([]float64, 44100),
		SampleRate: 44100,
		Channels:   1,
		Duration:   1.0,
	}
	for i := 0; i < 44100; i++ {
		t := float64(i) / 44100.0
		lowAmpData.Samples[i] = 0.1 * math.Sin(2*math.Pi*440*t)
	}

	normalizedData, err := processor.Normalize(lowAmpData)
	if err != nil {
		t.Fatalf("Failed to normalize: %v", err)
	}

	// Find the maximum amplitude
	maxAmp := 0.0
	for _, sample := range normalizedData.Samples {
		if math.Abs(sample) > maxAmp {
			maxAmp = math.Abs(sample)
		}
	}
	if math.Abs(maxAmp-1.0) > 0.01 {
		t.Errorf("Expected maximum amplitude 1.0, got %f", maxAmp)
	}
}

func TestAudioUtils(t *testing.T) {
	// Create an AudioUtils instance
	utils := NewAudioUtils()

	// Test CalculateRMS
	samples := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	rms := utils.CalculateRMS(samples)
	expectedRMS := math.Sqrt((0.1*0.1 + 0.2*0.2 + 0.3*0.3 + 0.4*0.4 + 0.5*0.5) / 5)
	if math.Abs(rms-expectedRMS) > 0.0001 {
		t.Errorf("Expected RMS %f, got %f", expectedRMS, rms)
	}

	// Test CalculateEnergy
	energy := utils.CalculateEnergy(samples)
	expectedEnergy := 0.1*0.1 + 0.2*0.2 + 0.3*0.3 + 0.4*0.4 + 0.5*0.5
	if math.Abs(energy-expectedEnergy) > 0.0001 {
		t.Errorf("Expected energy %f, got %f", expectedEnergy, energy)
	}

	// Test CalculateZeroCrossingRate
	crossingSamples := []float64{0.1, 0.2, -0.3, -0.4, 0.5}
	zcr := utils.CalculateZeroCrossingRate(crossingSamples)
	expectedZCR := 2.0 / 4.0 // 2 crossings in 4 transitions
	if math.Abs(zcr-expectedZCR) > 0.0001 {
		t.Errorf("Expected ZCR %f, got %f", expectedZCR, zcr)
	}

	// Test that all loaders are initialized
	if len(utils.Loaders) != 3 {
		t.Errorf("Expected 3 loaders, got %d", len(utils.Loaders))
	}

	// Check each loader type
	if _, ok := utils.Loaders[WAV]; !ok {
		t.Errorf("WAV loader not found")
	}
	if _, ok := utils.Loaders[MP3]; !ok {
		t.Errorf("MP3 loader not found")
	}
	if _, ok := utils.Loaders[FLAC]; !ok {
		t.Errorf("FLAC loader not found")
	}
}
