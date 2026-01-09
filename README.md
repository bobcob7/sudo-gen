# Merge Config

A Go toolkit for type-safe configuration merging and deep copying. This project provides code generation tools that create partial types and merge/copy methods for your structs.

## Overview

This project was inspired by an issue I encountered at work. I needed a framework for managing flexible configs that could merge and persist from multiple sources.

The solution includes:
- **sudo-gen**: A unified code generation tool with subcommands for generating merge and copy functionality
- **Playgrounds**: Experiments comparing different approaches to object merging in Go

## Quick Start

```go
package config

//go:generate go run merge-config/cmd/sudo-gen merge
//go:generate go run merge-config/cmd/sudo-gen copy
type Config struct {
    Name     string            `json:"name"`
    Port     int               `json:"port"`
    Database *DatabaseConfig   `json:"database"`
    Labels   map[string]string `json:"labels"`
}

type DatabaseConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}
```

Run `go generate ./...` to generate:
- `config_partial.go` - Partial types with pointer fields
- `config_merge.go` - `ApplyPartial` methods for merging
- `config_copy.go` - `Copy` methods for deep copying

## Tools

### [sudo-gen](cmd/sudo-gen/README.md)

Unified code generation tool with two subcommands:

| Subcommand | Description | Generated Files |
|------------|-------------|-----------------|
| `merge` | Generate partial types and ApplyPartial methods | `*_partial.go`, `*_merge.go` |
| `copy` | Generate deep copy methods | `*_copy.go` |

See the [sudo-gen README](cmd/sudo-gen/README.md) for detailed usage and examples.

## Use Cases

### Configuration Merging

When loading configuration from multiple sources (defaults, files, environment variables, CLI flags), you need to distinguish between "not set" and "zero value":

```go
// Start with defaults
cfg := &Config{Name: "default", Port: 8080}

// Apply file config (only set fields override)
cfg.ApplyPartial(fileConfig)

// Apply environment overrides
cfg.ApplyPartial(envConfig)
```

### Deep Copying

When you need independent copies of complex nested structures:

```go
original := &Config{Database: &DatabaseConfig{Host: "localhost"}}
copied := original.Copy()
copied.Database.Host = "remote" // doesn't affect original
```

## Project Structure

```
.
├── cmd/
│   └── sudo-gen/          # Unified code generation tool
├── internal/
│   └── codegen/           # Shared code generation logic
│       ├── merge/         # Merge-specific templates
│       └── copy/          # Copy-specific templates
├── examples/
│   └── basic/             # Example usage with generated code
└── playgrounds/
    └── merge-objects/     # Benchmarks comparing merge strategies
```

## Benchmarks

The `playgrounds/merge-objects` directory contains benchmarks comparing different merge strategies:

| Method | Performance | Use Case |
|--------|-------------|----------|
| Manual | Fastest | Production code |
| Reflection | Moderate | Generic utilities |
| JSON | Slowest | Simple prototyping |

See [playgrounds/merge-objects/results.md](playgrounds/merge-objects/results.md) for detailed benchmark results.

## Requirements

- Go 1.21+
