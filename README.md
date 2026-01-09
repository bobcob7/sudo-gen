# sudo-gen

A collection of Go code generators for struct boilerplate. Generate deep copy methods, partial types for merging, equality comparisons, and thread-safe configuration brokers - all without reflection at runtime.

## Overview

sudo-gen provides four code generators that eliminate common struct boilerplate:

| Generator | What it generates |
|-----------|-------------------|
| `copy` | Type-safe deep copy methods |
| `merge` | Partial types and `ApplyPartial` methods for config merging |
| `equals` | Type-safe equality comparison methods |
| `layerbroker` | Thread-safe config broker with ordered layers and field subscriptions |

## Installation

```bash
go install github.com/bobcob7/sudo-gen@latest
```

Or add it as a tool dependency in your project:

```bash
go get github.com/bobcob7/sudo-gen
```

## Usage

Add a `go:generate` directive above your struct definition:

```go
package config

//go:generate sudo-gen copy
type Config struct {
    Name     string
    Database *DatabaseConfig
}

type DatabaseConfig struct {
    Host string
    Port int
}
```

Run code generation:

```bash
go generate ./...
```

Each generator produces specific output files. See [Generators](#generators) below for details.

## Generators

### copy

Generates deep copy methods for structs.

```go
//go:generate sudo-gen copy
```

**Output:** `*_copy.go`

### merge

Generates partial types with pointer fields and `ApplyPartial` methods for merging configs.

```go
//go:generate sudo-gen merge
```

**Output:** `*_partial.go`, `*_merge.go`

### equals

Generates type-safe equality comparison methods.

```go
//go:generate sudo-gen equals
```

**Output:** `*_equals.go`

### layerbroker

Generates a thread-safe configuration broker with ordered layers and per-field subscriptions. Includes merge and copy output.

```go
//go:generate sudo-gen layerbroker
```

**Output:** `*_layerbroker.go`, `*_partial.go`, `*_merge.go`, `*_copy.go`

---

Run `sudo-gen -help` for all flags and advanced usage.

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
├── main.go                # Code generation tool entrypoint
├── internal/
│   └── codegen/           # Shared code generation logic
│       ├── merge/         # Merge-specific templates
│       ├── copy/          # Copy-specific templates
│       ├── equals/        # Equals-specific templates
│       └── layerbroker/   # LayerBroker templates
├── examples/
│   └── basic/             # Example usage with generated code
```

## Requirements

- Go 1.21+
