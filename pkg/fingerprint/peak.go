package fingerprint

import (
	"fmt"
	"sort"

	"github.com/kshitijk4poor/shazam-golang/pkg/audio"
)

// Peak represents a spectral peak in the time-frequency domain
type Peak struct {
	TimeIndex     int     // Index of the time bin
	FreqIndex     int     // Index of the frequency bin
	Time          float64 // Time position in seconds
	Frequency     float64 // Frequency in Hz
	Amplitude     float64 // Amplitude/energy of the peak
	IsLocalMaxima bool    // Whether this is a local maxima
}

// PeakExtractor extracts spectral peaks from a spectrogram
type PeakExtractor struct {
	// Configuration parameters
	NeighborhoodSize  int     // Size of the neighborhood for local maxima detection
	AbsoluteThreshold float64 // Absolute amplitude threshold
	RelativeThreshold float64 // Relative amplitude threshold (fraction of max amplitude)
	MaxPeaksPerFrame  int     // Maximum number of peaks to extract per time frame
	MinFrequency      float64 // Minimum frequency to consider (Hz)
	MaxFrequency      float64 // Maximum frequency to consider (Hz)
}

// NewPeakExtractor creates a new peak extractor with default settings
func NewPeakExtractor() *PeakExtractor {
	return &PeakExtractor{
		NeighborhoodSize:  3,     // 3x3 neighborhood
		AbsoluteThreshold: 0.01,  // Minimum amplitude
		RelativeThreshold: 0.1,   // 10% of maximum amplitude
		MaxPeaksPerFrame:  5,     // Maximum 5 peaks per time frame
		MinFrequency:      100.0, // Minimum frequency 100 Hz
		MaxFrequency:      4000.0, // Maximum frequency 4000 Hz
	}
}

// ExtractPeaks extracts spectral peaks from a spectrogram
func (p *PeakExtractor) ExtractPeaks(spectrogram *audio.Spectrogram) ([]Peak, error) {
	if spectrogram == nil || len(spectrogram.Data) == 0 || len(spectrogram.Data[0]) == 0 {
		return nil, fmt.Errorf("invalid spectrogram data")
	}

	// Find the global maximum amplitude for relative thresholding
	maxAmplitude := 0.0
	for _, frame := range spectrogram.Data {
		for _, amplitude := range frame {
			if amplitude > maxAmplitude {
				maxAmplitude = amplitude
			}
		}
	}

	// Calculate the effective threshold (maximum of absolute and relative)
	effectiveThreshold := p.AbsoluteThreshold
	relativeThresholdValue := maxAmplitude * p.RelativeThreshold
	if relativeThresholdValue > effectiveThreshold {
		effectiveThreshold = relativeThresholdValue
	}

	// Find the frequency bin indices corresponding to min/max frequencies
	minFreqIndex := 0
	maxFreqIndex := len(spectrogram.FreqPoints) - 1

	for i, freq := range spectrogram.FreqPoints {
		if freq >= p.MinFrequency && minFreqIndex == 0 {
			minFreqIndex = i
		}
		if freq > p.MaxFrequency {
			maxFreqIndex = i - 1
			break
		}
	}

	// Extract peaks
	var allPeaks []Peak

	// Process each time frame
	for t := 0; t < spectrogram.TimeBins; t++ {
		// Find peaks in this time frame
		var framePeaks []Peak

		// Check each frequency bin within the specified range
		for f := minFreqIndex; f <= maxFreqIndex; f++ {
			amplitude := spectrogram.Data[t][f]

			// Skip if below threshold
			if amplitude < effectiveThreshold {
				continue
			}

			// Check if this is a local maximum in its neighborhood
			isLocalMax := true
			for dt := -p.NeighborhoodSize; dt <= p.NeighborhoodSize; dt++ {
				for df := -p.NeighborhoodSize; df <= p.NeighborhoodSize; df++ {
					// Skip the point itself
					if dt == 0 && df == 0 {
						continue
					}

					// Check if the neighbor is within bounds
					nt := t + dt
					nf := f + df
					if nt >= 0 && nt < spectrogram.TimeBins && nf >= 0 && nf < spectrogram.FreqBins {
						// If any neighbor has higher amplitude, this is not a local maximum
						if spectrogram.Data[nt][nf] > amplitude {
							isLocalMax = false
							break
						}
					}
				}
				if !isLocalMax {
					break
				}
			}

			// If this is a local maximum, add it to the peaks
			if isLocalMax {
				peak := Peak{
					TimeIndex:     t,
					FreqIndex:     f,
					Time:          spectrogram.TimePoints[t],
					Frequency:     spectrogram.FreqPoints[f],
					Amplitude:     amplitude,
					IsLocalMaxima: true,
				}
				framePeaks = append(framePeaks, peak)
			}
		}

		// Sort peaks by amplitude (descending) and take the top N
		if len(framePeaks) > 0 {
			sort.Slice(framePeaks, func(i, j int) bool {
				return framePeaks[i].Amplitude > framePeaks[j].Amplitude
			})

			// Limit the number of peaks per frame
			if len(framePeaks) > p.MaxPeaksPerFrame {
				framePeaks = framePeaks[:p.MaxPeaksPerFrame]
			}

			// Add to all peaks
			allPeaks = append(allPeaks, framePeaks...)
		}
	}

	return allPeaks, nil
}

// FilterPeaks filters peaks based on various criteria
func (p *PeakExtractor) FilterPeaks(peaks []Peak, options map[string]interface{}) []Peak {
	if len(peaks) == 0 {
		return peaks
	}

	filtered := make([]Peak, 0, len(peaks))

	// Apply frequency range filtering
	minFreq, hasMinFreq := options["min_frequency"].(float64)
	maxFreq, hasMaxFreq := options["max_frequency"].(float64)
	
	// Apply amplitude threshold filtering
	ampThreshold, hasAmpThreshold := options["amplitude_threshold"].(float64)
	
	// Apply time range filtering
	minTime, hasMinTime := options["min_time"].(float64)
	maxTime, hasMaxTime := options["max_time"].(float64)

	// Filter peaks
	for _, peak := range peaks {
		// Check frequency range
		if hasMinFreq && peak.Frequency < minFreq {
			continue
		}
		if hasMaxFreq && peak.Frequency > maxFreq {
			continue
		}
		
		// Check amplitude threshold
		if hasAmpThreshold && peak.Amplitude < ampThreshold {
			continue
		}
		
		// Check time range
		if hasMinTime && peak.Time < minTime {
			continue
		}
		if hasMaxTime && peak.Time > maxTime {
			continue
		}
		
		// If all checks pass, add to filtered peaks
		filtered = append(filtered, peak)
	}

	return filtered
}

// VisualizePeaks creates a visualization of peaks on a spectrogram
func (p *PeakExtractor) VisualizePeaks(spectrogram *audio.Spectrogram, peaks []Peak, filePath string) error {
	// Import required packages
	import (
		"image"
		"image/color"
		"image/draw"
		"image/png"
		"os"
	)

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

	// Draw peaks on the image
	peakColor := color.RGBA{255, 255, 255, 255} // White
	for _, peak := range peaks {
		// Calculate the position in the image
		x := peak.TimeIndex
		y := height - peak.FreqIndex - 1 // Invert frequency axis
		
		// Draw a small circle around the peak
		for dx := -2; dx <= 2; dx++ {
			for dy := -2; dy <= 2; dy++ {
				// Skip corners to make it more circular
				if dx*dx + dy*dy > 5 {
					continue
				}
				
				// Check if the point is within the image bounds
				nx := x + dx
				ny := y + dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					img.Set(nx, ny, peakColor)
				}
			}
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