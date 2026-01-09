package copyobjects

import "time"

//go:generate go run ../../cmd/sudo-copy
// Config is the target type that merge functions produce.
// It contains various serializable field types to test different merge scenarios.
type Config struct {
	// Basic types
	Name       string  `json:"name,omitempty"`
	Port       int     `json:"port,omitempty"`
	MaxRetries int32   `json:"max_retries,omitempty"`
	Timeout    int64   `json:"timeout,omitempty"`
	Rate       float64 `json:"rate,omitempty"`
	Enabled    bool    `json:"enabled,omitempty"`
	EnabledPtr *bool   `json:"enabled_ptr,omitempty"`

	// String pointer (to distinguish unset from empty)
	Description *string `json:"description,omitempty"`

	// Slice types
	Hosts []string `json:"hosts,omitempty"`
	Ports []int    `json:"ports,omitempty"`
	Tags  []Tag    `json:"tags,omitempty"`

	// Map types
	Labels   map[string]string `json:"labels,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`

	// Nested struct
	Database    DatabaseConfig  `json:"database,omitempty"`
	DatabasePtr *DatabaseConfig `json:"database_ptr,omitempty"`

	// Time
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// DatabaseConfig represents a nested configuration struct.
type DatabaseConfig struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	SSLMode  string `json:"ssl_mode,omitempty"`
}

// Tag represents a simple key-value struct for slice testing.
type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// InputConfig is similar to Config but used as input for merge functions.
// It uses pointers for most fields to distinguish between "not set" and "zero value".
type InputConfig struct {
	// Basic types as pointers
	Name       *string  `json:"name,omitempty"`
	Port       *int     `json:"port,omitempty"`
	MaxRetries *int32   `json:"max_retries,omitempty"`
	Timeout    *int64   `json:"timeout,omitempty"`
	Rate       *float64 `json:"rate,omitempty"`
	Enabled    *bool    `json:"enabled,omitempty"`
	EnabledPtr *bool    `json:"enabled_ptr,omitempty"`

	// String pointer
	Description *string `json:"description,omitempty"`

	// Slice types
	Hosts []string   `json:"hosts,omitempty"`
	Ports []int      `json:"ports,omitempty"`
	Tags  []InputTag `json:"tags,omitempty"`

	// Map types
	Labels   map[string]string `json:"labels,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`

	// Nested struct as pointer
	Database    *InputDatabaseConfig `json:"database,omitempty"`
	DatabasePtr *InputDatabaseConfig `json:"database_ptr,omitempty"`

	// Time
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// InputDatabaseConfig is the input version of DatabaseConfig with pointer fields.
type InputDatabaseConfig struct {
	Host     *string `json:"host,omitempty"`
	Port     *int    `json:"port,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	SSLMode  *string `json:"ssl_mode,omitempty"`
}

// InputTag is the input version of Tag with pointer fields.
type InputTag struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// Helper functions to create pointers for testing
func Ptr[T any](v T) *T {
	return &v
}
