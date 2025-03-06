package db

import (
	"context"

	"github.com/kshitijk4poor/shazam-golang/pkg/fingerprint"
)

// TrackMetadata contains information about an audio track
type TrackMetadata struct {
	ID       string
	Title    string
	Artist   string
	Duration float64
	Added    int64 // Unix timestamp
}

// SearchResult represents a match from the vector database
type SearchResult struct {
	TrackID       string
	Score         float64
	TimeOffset    float64
	MatchedVector *fingerprint.Vector
}

// VectorDB defines interface for vector database operations
type VectorDB interface {
	// Add inserts vectors and metadata for a track
	Add(ctx context.Context, metadata *TrackMetadata, vectors []*fingerprint.Vector) error

	// Search finds nearest neighbors for query vectors
	Search(ctx context.Context, query []*fingerprint.Vector, k int) ([]SearchResult, error)

	// Delete removes a track and its vectors
	Delete(ctx context.Context, trackID string) error

	// Get retrieves track metadata
	Get(ctx context.Context, trackID string) (*TrackMetadata, error)

	// List returns all track metadata
	List(ctx context.Context) ([]*TrackMetadata, error)

	// Save persists the database to disk
	Save(ctx context.Context, path string) error

	// Load restores the database from disk
	Load(ctx context.Context, path string) error
}

// Config holds database configuration
type Config struct {
	M              int // Number of connections in HNSW graph
	EfConstruction int // Size of dynamic candidate list for construction
	EfSearch       int // Size of dynamic candidate list for search
	Dim            int // Vector dimensionality
	MaxElements    int // Maximum number of vectors to store
}
