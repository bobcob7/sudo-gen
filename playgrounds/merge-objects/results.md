# Object Merging in Go: Analysis and Benchmarks

## Problem Overview

When building configurable applications, a common pattern is to merge multiple configuration sources together. For example:
- Default configuration from the application
- Configuration file on disk
- Environment variable overrides
- Command-line argument overrides

Each layer should override the previous one, but only for fields that are explicitly set. This creates a challenge: **how do we distinguish between "not set" and "zero value"?**

For instance, if a user explicitly sets `port: 0`, that's different from not specifying the port at all. The typical solution is to use pointer types in the input structs—a `nil` pointer means "not set", while a non-nil pointer (even pointing to a zero value) means "explicitly set".

This playground explores different approaches to merging two input configurations into a single output configuration, where:
- Input types use pointers to distinguish "not set" from "zero value"
- Output type uses concrete values (no pointers for basic types)
- Values from the second input override values from the first input

---

## Solution Overview

### 1. Manual Merge (`MergeManual`)

**Approach:** Explicit field-by-field assignment with nil checks.

```go
if src.Port != nil {
    dst.Port = *src.Port
}
```

**Pros:**
- Fastest performance
- Compile-time type safety
- No runtime reflection overhead
- Clear, readable code
- Easy to debug

**Cons:**
- Requires manual updates when types change
- Verbose for large structs
- Easy to miss fields when adding new ones
- Code duplication between similar merge operations

---

### 2. Type-Aware Reflection (`MergeReflection`)

**Approach:** Uses reflection but with type-specific handling functions. The main logic understands the relationship between `InputConfig` and `Config` types.

```go
func mergeInputToConfig(dst *Config, src *InputConfig) {
    if src.Name != nil {
        dst.Name = *src.Name
    }
    // ... other fields
}
```

**Pros:**
- Nearly as fast as manual merge
- Still type-aware at the top level
- Can be refactored to use reflection for specific complex sections

**Cons:**
- Same maintenance burden as manual merge
- Not truly generic

---

### 3. Generic Reflection (`MergeReflectionGeneric`)

**Approach:** Fully generic reflection-based merge that matches fields by JSON tags. Works with any struct types that have compatible JSON tags.

```go
func mergeStructByReflection(dst, src reflect.Value) {
    for i := 0; i < src.NumField(); i++ {
        // Find matching field in dst by JSON tag
        // Handle pointers, slices, maps, nested structs
    }
}
```

**Pros:**
- Truly generic—works with any compatible types
- No manual field mapping required
- Automatically handles new fields (if JSON tags match)

**Cons:**
- Significantly slower (~100x slower than manual)
- Complex implementation
- Runtime errors instead of compile-time errors
- Harder to debug
- More memory allocations

---

### 4. JSON Marshal/Unmarshal (`MergeJSON`)

**Approach:** Marshal input to JSON, then unmarshal into the output struct. Relies on `omitempty` tags to skip nil/zero fields.

```go
data, _ := json.Marshal(input1)
json.Unmarshal(data, &result)
data, _ = json.Marshal(input2)
json.Unmarshal(data, &result)
```

**Pros:**
- Simple, concise implementation
- Leverages well-tested standard library
- Automatically handles type conversion
- Easy to understand

**Cons:**
- ~20x slower than manual merge
- Higher memory usage (string allocations)
- Loses type information during serialization
- Cannot distinguish between explicit zero and omitted values after first unmarshal

---

### 5. JSON with Intermediate Map (`MergeJSONWithMap`)

**Approach:** Convert both inputs to `map[string]any`, deep merge the maps, then convert to the output type.

```go
var m1, m2 map[string]any
json.Unmarshal(json.Marshal(input1), &m1)
json.Unmarshal(json.Marshal(input2), &m2)
deepMergeMap(m1, m2)
json.Unmarshal(json.Marshal(m1), &result)
```

**Pros:**
- Provides deep merging for nested maps
- Full control over merge behavior
- Can handle arbitrary nesting

**Cons:**
- Slowest implementation (~40x slower than manual)
- Highest memory usage
- Multiple serialization/deserialization passes
- Type information completely lost in intermediate state

---

## Benchmark Results

```
goos: darwin
goarch: arm64
pkg: merge-config/playgrounds/merge-objects
cpu: Apple M4

BenchmarkMergeImplementations/MergeManual-10                    14458065    282.1 ns/op     384 B/op       3 allocs/op
BenchmarkMergeImplementations/MergeReflection-10                13424118    306.1 ns/op     384 B/op       3 allocs/op
BenchmarkMergeImplementations/MergeReflectionGeneric-10           122442  26811.0 ns/op     920 B/op      15 allocs/op
BenchmarkMergeImplementations/MergeJSON-10                        748659   5408.0 ns/op    1969 B/op      48 allocs/op
BenchmarkMergeImplementations/MergeJSONWithMap-10                 315933  10911.0 ns/op    6507 B/op     148 allocs/op
```

### Performance Comparison

| Implementation | Time (ns/op) | Relative Speed | Memory (B/op) | Allocations |
|----------------|--------------|----------------|---------------|-------------|
| MergeManual | 282 | 1.0x (baseline) | 384 | 3 |
| MergeReflection | 306 | 1.1x slower | 384 | 3 |
| MergeJSON | 5,408 | 19.2x slower | 1,969 | 48 |
| MergeJSONWithMap | 10,911 | 38.7x slower | 6,507 | 148 |
| MergeReflectionGeneric | 26,811 | 95.1x slower | 920 | 15 |

### Analysis

**MergeManual & MergeReflection** are essentially equivalent in performance. Both achieve ~280-306 ns/op with only 3 allocations. The "reflection" version isn't truly using reflection for the core logic—it's just organized differently. These are the clear winners for performance.

**MergeJSON** is surprisingly fast for a serialization-based approach at ~5.4μs. The 48 allocations come from string/byte slice allocations during JSON encoding/decoding. This is a reasonable trade-off for simplicity in non-hot-path code.

**MergeJSONWithMap** doubles the time of MergeJSON because it does an additional serialization round-trip through the intermediate map. The 148 allocations reflect all the map entries and interface boxing.

**MergeReflectionGeneric** is the slowest despite not using JSON. The overhead comes from:
- Repeated calls to `reflect.Value.Field()` and `reflect.Type.Field()`
- JSON tag parsing on every field access
- Dynamic type checking and conversion
- Creating new reflect.Value objects for nested operations

---

## Recommendation

### Best Solution: **MergeManual** (or MergeReflection)

For most use cases, **MergeManual** is the recommended approach because:

1. **Performance**: It's the fastest option at 282 ns/op with minimal allocations. For configuration merging that happens at startup, this might not matter—but if you're merging frequently (e.g., per-request overrides), it absolutely does.

2. **Type Safety**: Compile-time checks catch errors before runtime. If you add a field to `InputConfig` but forget to handle it in the merge function, the compiler won't catch it—but neither will any other approach.

3. **Debuggability**: When something goes wrong, you can step through explicit code rather than reflect machinery or JSON parsing.

4. **Maintainability**: Despite being verbose, the code is straightforward. Any Go developer can understand and modify it.

### When to Choose Alternatives

| Situation | Recommended Approach |
|-----------|---------------------|
| Hot path, performance critical | MergeManual |
| Startup-only config, simplicity preferred | MergeJSON |
| Need deep map merging | MergeJSONWithMap |
| Many different type pairs to merge | MergeReflectionGeneric (with caching) |
| Protobuf or other non-JSON types | MergeManual or MergeReflection |

### Future Improvements

If using the generic reflection approach, consider:
- **Caching struct metadata**: Parse JSON tags once and cache the field mappings
- **Code generation**: Use `go generate` to create type-specific merge functions
- **Third-party libraries**: Consider [mergo](https://github.com/darccio/mergo) or [copier](https://github.com/jinzhu/copier) which have optimized implementations

For this playground's use case (merging configuration objects), **MergeManual** provides the best balance of performance, safety, and clarity.
