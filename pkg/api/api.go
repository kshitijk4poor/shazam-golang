package api

import (
	"github.com/kshitijk4poor/shazam-golang/pkg/db"
	"github.com/kshitijk4poor/shazam-golang/pkg/matcher"
)

// IdentifyRequest represents an audio identification request
type IdentifyRequest struct {
	AudioData []byte `json:"-"`      // Raw audio data
	Format    string `json:"format"` // Audio format (wav, mp3, etc)
}

// IdentifyResponse represents the response to an identification request
type IdentifyResponse struct {
	Matches []matcher.Match `json:"matches"`
	Error   string          `json:"error,omitempty"`
}

// AddTrackRequest represents a request to add a reference track
type AddTrackRequest struct {
	AudioData []byte           `json:"-"`
	Format    string           `json:"format"`
	Metadata  db.TrackMetadata `json:"metadata"`
}

// AddTrackResponse represents the response to an add track request
type AddTrackResponse struct {
	TrackID string `json:"track_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListTracksResponse represents the response to a list tracks request
type ListTracksResponse struct {
	Tracks []db.TrackMetadata `json:"tracks"`
	Error  string             `json:"error,omitempty"`
}

// Config holds API service configuration
type Config struct {
	Port            int    `json:"port"`
	Host            string `json:"host"`
	MaxRequestSize  int64  `json:"max_request_size"` // Maximum size of request body in bytes
	ReadTimeout     int    `json:"read_timeout"`     // Read timeout in seconds
	WriteTimeout    int    `json:"write_timeout"`    // Write timeout in seconds
	ShutdownTimeout int    `json:"shutdown_timeout"` // Graceful shutdown timeout in seconds
}
