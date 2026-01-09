# Go Coding Standards

This document outlines the coding standards for Go code in this project. It combines official Go best practices with project-specific conventions.

## Formatting

### Use `gofmt` and `goimports`

All code must be formatted with `gofmt`. Use `goimports` to automatically manage imports.

### No Empty Lines Within Function Bodies

Keep function bodies compact. Avoid empty lines between statements unless separating distinct logical phases of a complex algorithm.

```go
// Good
func processConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	validated := validate(cfg)
	return save(validated)
}

// Bad
func processConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	validated := validate(cfg)

	return save(validated)
}
```

### Line Length

Aim for lines under 100 characters. Break long lines at logical points.

## Naming

### Follow Go Conventions

- Use `MixedCaps` or `mixedCaps`, never underscores
- Acronyms should be all caps: `HTTPServer`, `userID`
- Interface names: single-method interfaces use method name + "er" suffix (`Reader`, `Writer`)
- Package names: short, lowercase, no underscores or mixed caps

### Be Concise But Clear

Variable names should be short in small scopes, longer in larger scopes.

```go
// Good - short scope
for i, v := range items {
	process(v)
}

// Good - longer scope
func (s *Server) handleUserRegistration(ctx context.Context, req *RegistrationRequest) error {
	validatedEmail := validateEmail(req.Email)
	// ...
}
```

### Receiver Names

Use one or two letter abbreviations, consistent across methods.

```go
// Good
func (c *Config) Copy() *Config
func (c *Config) Validate() error

// Bad
func (config *Config) Copy() *Config
func (this *Config) Validate() error
```

## Error Handling

### Check Errors Immediately

```go
// Good
f, err := os.Open(filename)
if err != nil {
	return err
}
defer f.Close()

// Bad
f, _ := os.Open(filename)
```

### Wrap Errors with Context

Use `fmt.Errorf` with `%w` to wrap errors.

```go
if err := db.Connect(); err != nil {
	return fmt.Errorf("connecting to database: %w", err)
}
```

### Error Messages

- Start with lowercase
- No trailing punctuation
- Describe what failed, not what succeeded

```go
// Good
return fmt.Errorf("parsing config file: %w", err)

// Bad
return fmt.Errorf("Failed to parse config file.: %w", err)
```

## Code Organization

### Package Structure

- One package per directory
- Package name matches directory name (except `main`)
- Keep packages focused on a single responsibility

### File Organization

Order declarations within a file:
1. Package documentation
2. Package declaration
3. Imports (grouped: stdlib, external, internal)
4. Constants
5. Variables
6. Types
7. Functions (constructors first, then methods grouped by receiver)

### Import Grouping

```go
import (
	"context"
	"fmt"

	"github.com/external/package"

	"github.com/bobcob7/sudo-gen/internal/config"
)
```

## Functions

### Keep Functions Small

Functions should do one thing. If a function needs a comment explaining what a section does, extract that section into a named function.

### Limit Parameters

Prefer fewer parameters. If a function needs many parameters, consider:
- Grouping related parameters into a struct
- Using functional options pattern
- Breaking the function into smaller pieces

```go
// Good
type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

func NewServer(cfg ServerConfig) *Server

// Avoid
func NewServer(host string, port int, timeout time.Duration, maxConns int, ...) *Server
```

### Return Early

Reduce nesting by handling errors and edge cases first.

```go
// Good
func process(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.Name == "" {
		return errors.New("name is required")
	}
	return doProcess(cfg)
}

// Bad
func process(cfg *Config) error {
	if cfg != nil {
		if cfg.Name != "" {
			return doProcess(cfg)
		} else {
			return errors.New("name is required")
		}
	} else {
		return errors.New("config is nil")
	}
}
```

## Interfaces

### Accept Interfaces, Return Structs

Functions should accept interface parameters and return concrete types.

```go
// Good
func ProcessReader(r io.Reader) (*Result, error)

// Avoid
func ProcessReader(r io.Reader) (ResultInterface, error)
```

### Keep Interfaces Small

Prefer many small interfaces over few large ones.

```go
// Good
type Reader interface {
	Read(p []byte) error
}

type Writer interface {
	Write(p []byte) error
}

type ReadWriter interface {
	Reader
	Writer
}
```

### Define Interfaces at Point of Use

Define interfaces in the package that uses them, not the package that implements them.

## Concurrency

### Don't Start Goroutines in Library Code

Let the caller control concurrency unless the API is explicitly concurrent.

### Always Handle Goroutine Lifecycle

```go
func (s *Server) Start(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			s.shutdown()
		}
	}()
}
```

### Use Channels for Communication

Prefer channels over shared memory with mutexes when possible.

## Testing

### Test File Naming

- `foo_test.go` for tests of `foo.go`
- Use `_test` package suffix for black-box testing when appropriate

### Table-Driven Tests

Use table-driven tests for testing multiple cases.

```go
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "valid config",
			input:   &Config{Name: "test"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

### Test Helper Functions

Use `t.Helper()` in test helper functions.

```go
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

## Refactoring for Testability

### Extract Dependencies

Move external dependencies (database, HTTP clients, file system) behind interfaces.

```go
// Before: hard to test
func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("config.json")
	// ...
}

// After: testable
type FileReader interface {
	ReadFile(name string) ([]byte, error)
}

func LoadConfig(reader FileReader) (*Config, error) {
	data, err := reader.ReadFile("config.json")
	// ...
}
```

### Separate Pure Logic from Side Effects

Extract pure functions that can be tested without mocks.

```go
// Before: mixed concerns
func ProcessAndSave(input string) error {
	// parsing logic
	// validation logic
	// transformation logic
	return db.Save(result)
}

// After: separated concerns
func Parse(input string) (*Data, error) {
	// pure parsing logic - easily testable
}

func Validate(d *Data) error {
	// pure validation - easily testable
}

func Transform(d *Data) *Result {
	// pure transformation - easily testable
}

func ProcessAndSave(input string, db Database) error {
	data, err := Parse(input)
	if err != nil {
		return err
	}
	if err := Validate(data); err != nil {
		return err
	}
	result := Transform(data)
	return db.Save(result)
}
```

### Use Functional Options for Complex Configuration

```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.timeout = d
	}
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
```

### Avoid Global State

Global state makes testing difficult. Pass dependencies explicitly.

```go
// Bad
var db *Database

func GetUser(id int) (*User, error) {
	return db.FindUser(id)
}

// Good
type UserService struct {
	db Database
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.db.FindUser(id)
}
```

### Break Up Large Functions

Large functions are hard to test comprehensively. Break them into smaller, focused functions.

```go
// Before: 100+ line function
func ProcessOrder(order *Order) error {
	// validate
	// calculate totals
	// apply discounts
	// check inventory
	// reserve items
	// charge payment
	// send confirmation
	// ...
}

// After: composable, testable pieces
func (s *OrderService) ProcessOrder(order *Order) error {
	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	totals := s.calculateTotals(order)
	totals = s.applyDiscounts(totals, order.Coupons)
	if err := s.checkInventory(order.Items); err != nil {
		return fmt.Errorf("inventory: %w", err)
	}
	if err := s.reserveItems(order.Items); err != nil {
		return fmt.Errorf("reservation: %w", err)
	}
	if err := s.chargePayment(order.Payment, totals.Total); err != nil {
		return fmt.Errorf("payment: %w", err)
	}
	return s.sendConfirmation(order)
}
```

## Performance

### Preallocate Slices and Maps

When the size is known, preallocate to avoid repeated allocations.

```go
// Good
result := make([]string, 0, len(input))
for _, v := range input {
	result = append(result, transform(v))
}

// Less efficient
var result []string
for _, v := range input {
	result = append(result, transform(v))
}
```

### Avoid Unnecessary Allocations

```go
// Good - reuse buffer
var buf bytes.Buffer
for _, item := range items {
	buf.Reset()
	buf.WriteString(item)
	process(buf.Bytes())
}

// Bad - allocates each iteration
for _, item := range items {
	var buf bytes.Buffer
	buf.WriteString(item)
	process(buf.Bytes())
}
```

### Use `strings.Builder` for String Concatenation

```go
// Good
var b strings.Builder
for _, s := range parts {
	b.WriteString(s)
}
result := b.String()

// Bad
var result string
for _, s := range parts {
	result += s
}
```

### Profile Before Optimizing

Don't optimize prematurely. Use `go test -bench` and `pprof` to identify actual bottlenecks.

## Documentation

### Package Documentation

Every package should have a package comment in one file (usually `doc.go` or the primary file).

```go
// Package config provides configuration loading and validation
// for the application.
package config
```

### Exported Identifiers

All exported identifiers should have a doc comment starting with the identifier name.

```go
// Config represents the application configuration.
// It is safe for concurrent read access.
type Config struct {
	// ...
}

// Load reads configuration from the given path.
// It returns an error if the file cannot be read or parsed.
func Load(path string) (*Config, error) {
	// ...
}
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
