# sudo-gen

A unified code generation tool for Go structs. Generate partial types, merge methods, and deep copy methods with a single command.

## Installation

```bash
go install merge-config/cmd/sudo-gen@latest
```

Or use directly with `go run`:

```go
//go:generate go run github.com/yourorg/merge-config/cmd/sudo-gen merge
```

## Subcommands

### merge

Generates partial types and `ApplyPartial` methods for type-safe config merging.

**Generated files:**
- `{source}_partial.go` - Partial version of the struct with pointer fields
- `{source}_merge.go` - `ApplyPartial` method for merging partials into the original type

**Use case:** Merging configuration from multiple sources (defaults, file, environment, CLI flags) where you need to distinguish between "not set" and "zero value".

### copy

Generates deep copy methods for structs.

**Generated files:**
- `{type}_copy.go` - `Copy()` method that performs a deep copy

**Use case:** Creating independent copies of complex nested structures without shared references.

## Usage

### Basic Usage

Place the `go:generate` directive directly above your struct definition:

```go
//go:generate go run merge-config/cmd/sudo-gen merge
//go:generate go run merge-config/cmd/sudo-gen copy
type Config struct {
    Name     string            `json:"name"`
    Port     int               `json:"port"`
    Database *DatabaseConfig   `json:"database"`
    Labels   map[string]string `json:"labels"`
}
```

Then run:

```bash
go generate ./...
```

### With Explicit Type

If the directive is not directly above the struct, specify the type explicitly:

```go
//go:generate go run merge-config/cmd/sudo-gen merge -type=Config
//go:generate go run merge-config/cmd/sudo-gen copy -type=Config
```

### Custom Method Name (copy only)

```go
//go:generate go run merge-config/cmd/sudo-gen copy -method=Clone
```

This generates a `Clone()` method instead of `Copy()`.

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-type` | Name of the struct type | Auto-detected from directive position |
| `-output` | Output directory for generated files | Same as source |
| `-package` | Package name for generated files | Same as source |
| `-method` | Name of the generated copy method (copy only) | `Copy` |
| `-help` | Show help message | - |

## Generated Code Examples

### merge subcommand

Given this input:

```go
type Config struct {
    Name     string          `json:"name"`
    Port     int             `json:"port"`
    Database *DatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}
```

**Generated `config_partial.go`:**

```go
type ConfigPartial struct {
    Name     *string                `json:"name,omitempty"`
    Port     *int                   `json:"port,omitempty"`
    Database *DatabaseConfigPartial `json:"database,omitempty"`
}

type DatabaseConfigPartial struct {
    Host *string `json:"host,omitempty"`
    Port *int    `json:"port,omitempty"`
}
```

**Generated `config_merge.go`:**

```go
func (c *Config) ApplyPartial(p *ConfigPartial) {
    if c == nil || p == nil {
        return
    }
    if p.Name != nil {
        c.Name = *p.Name
    }
    if p.Port != nil {
        c.Port = *p.Port
    }
    if p.Database != nil {
        if c.Database == nil {
            c.Database = &DatabaseConfig{}
        }
        c.Database.ApplyPartial(p.Database)
    }
}
```

**Usage:**

```go
// Start with defaults
cfg := &Config{
    Name: "default",
    Port: 8080,
}

// Apply file config (only overrides what's set)
filePartial := &ConfigPartial{
    Port: ptr(9090), // helper: func ptr[T any](v T) *T { return &v }
}
cfg.ApplyPartial(filePartial)

// Apply env config
envPartial := &ConfigPartial{
    Name: ptr("from-env"),
}
cfg.ApplyPartial(envPartial)

// Result: Name="from-env", Port=9090
```

### copy subcommand

**Generated `config_copy.go`:**

```go
func (c *Config) Copy() *Config {
    if c == nil {
        return nil
    }
    dst := &Config{}
    dst.Name = c.Name
    dst.Port = c.Port
    if c.Database != nil {
        dst.Database = c.Database.Copy()
    }
    return dst
}

func (c *DatabaseConfig) Copy() *DatabaseConfig {
    if c == nil {
        return nil
    }
    dst := &DatabaseConfig{}
    dst.Host = c.Host
    dst.Port = c.Port
    return dst
}
```

**Usage:**

```go
original := &Config{
    Name: "test",
    Database: &DatabaseConfig{Host: "localhost"},
}

copied := original.Copy()
copied.Database.Host = "remote" // doesn't affect original
```

## Supported Field Types

Both subcommands support:

- Basic types (`string`, `int`, `bool`, `float64`, etc.)
- Pointer types (`*string`, `*int`, etc.)
- Slices (`[]string`, `[]MyStruct`)
- Maps (`map[string]string`, `map[string]any`)
- Nested structs (local package only)
- External types (`time.Time`)
- Pointer to structs (`*DatabaseConfig`)

## Type Detection

The tool automatically detects the target type when the `go:generate` directive is placed directly above the struct:

```go
//go:generate go run merge-config/cmd/sudo-gen merge
type Config struct { // <-- This type is detected automatically
    ...
}
```

Detection works by:
1. Looking for a struct declaration immediately following the `go:generate` directive
2. Falling back to `GOLINE` environment variable to find the struct after that line

## Nested Struct Handling

When a struct contains references to other structs in the same package, the tool automatically generates partial types and methods for those nested types as well.

```go
type Config struct {
    Database *DatabaseConfig // Generates DatabaseConfigPartial and DatabaseConfig.ApplyPartial
    Tags     []Tag           // Generates TagPartial and Tag.ApplyPartial
}
```

## Best Practices

1. **Place directives directly above the type** for automatic type detection
2. **Use both merge and copy together** when you need config merging and safe copying
3. **Run `go generate` after modifying structs** to regenerate the code
4. **Don't edit generated files** - they will be overwritten

## Architecture

```
cmd/sudo-gen/
├── main.go           # CLI entrypoint with subcommand routing
└── README.md         # This file

internal/codegen/
├── types.go          # Shared types (StructInfo, FieldInfo, etc.)
├── parser.go         # AST parsing utilities
├── generator.go      # Template generation utilities
├── merge/
│   └── merge.go      # Merge-specific templates and logic
└── copy/
    └── copy.go       # Copy-specific templates and logic
```
