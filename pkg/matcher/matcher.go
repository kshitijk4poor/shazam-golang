package matcher

import (
	"context"

	"github.com/kshitijk4poor/shazam-golang/pkg/audio"
	"github.com/kshitijk4poor/shazam-golang/pkg/db"
)

// Match represents a confident match with a track
type Match struct {
	TrackID        string
	Confidence     float64
	TimeOffset     float64 // Matched position in reference track
	QueryTime      float64 // Position in query audio
	MatchedVectors int     // Number of matching vectors
}

// Engine handles audio identification
type Engine interface {
	// Identify processes query audio and returns matches
	Identify(ctx context.Context, data *audio.AudioData) ([]Match, error)

	// AddTrack processes and adds a reference track
	AddTrack(ctx context.Context, data *audio.AudioData, metadata *db.TrackMetadata) error
}

// Config holds matcher configuration
type Config struct {
	MinConfidence     float64 // Minimum confidence for valid match
	MinMatchedVectors int     // Minimum number of matching vectors
	SearchNeighbors   int     // Number of neighbors to search for each vector
	TimeAlignWindow   float64 // Time window for alignment verification
	MaxTimeDeviation  float64 // Maximum allowed time offset deviation
}

// TimeAlignment handles verification of temporal consistency
type TimeAlignment interface {
	// VerifyAlignment checks if matched vectors have consistent time offsets
	VerifyAlignment(matches []db.SearchResult) ([]Match, error)
}

// Scorer handles match scoring and ranking
type Scorer interface {
	// Score computes confidence scores for potential matches
	Score(matches []db.SearchResult) ([]Match, error)
}
