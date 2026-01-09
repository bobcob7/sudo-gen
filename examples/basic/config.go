package basic

import "time"

//go:generate go run ../../cmd/sudo-gen merge
//go:generate go run ../../cmd/sudo-gen copy
//go:generate go run ../../cmd/sudo-gen manager
type Config struct {
	// Basic types
	Name        string  `json:"name,omitempty"`
	Port        int     `json:"port,omitempty"`
	MaxRetries  int32   `json:"max_retries,omitempty"`
	Timeout     int64   `json:"timeout,omitempty"`
	Rate        float64 `json:"rate,omitempty"`
	Enabled     bool    `json:"enabled,omitempty"`
	Description *string `json:"description,omitempty"`

	// Slice types
	Hosts []string `json:"hosts,omitempty"`
	Tags  []Tag    `json:"tags,omitempty"`

	// Map types
	Labels   map[string]string `json:"labels,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`

	// Nested struct
	Database *DatabaseConfig `json:"database,omitempty"`

	// Time
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// DatabaseConfig represents database connection settings.
type DatabaseConfig struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	SSLMode  string `json:"ssl_mode,omitempty"`
}

// Tag represents a key-value tag.
type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
