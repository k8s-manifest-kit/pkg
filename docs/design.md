# Design Document: pkg

## 1. Introduction

This package provides core utilities for the k8s-manifest-kit ecosystem. It contains shared functionality for caching, value merging, error handling, and Kubernetes object manipulation that is used across multiple k8s-manifest-kit components.

## 2. Package Structure

```
pkg/
├── util/
│   ├── cache/          # TTL-based caching with deep cloning
│   │   ├── cache.go
│   │   ├── cache_option.go
│   │   └── cache_test.go
│   ├── errors/         # Error handling utilities
│   │   └── errors.go
│   ├── jq/             # JQ expression utilities
│   │   ├── jq.go
│   │   └── jq_test.go
│   ├── k8s/            # Kubernetes object utilities
│   │   ├── k8s.go
│   │   └── k8s_test.go
│   ├── metrics/        # Metrics collection
│   │   ├── metrics.go
│   │   ├── metrics_test.go
│   │   ├── memory/     # In-memory metrics collector
│   │   └── noop/       # No-op metrics collector
│   ├── maps/           # Deep clone and merge for map[string]any
│   │   ├── clone.go
│   │   ├── clone_test.go
│   │   ├── merge.go
│   │   └── merge_test.go
│   └── option.go       # Functional options pattern support
```

## 3. Value Merging Strategy (pkg/util/maps)

The deep merge utility provides a robust way to merge nested map structures with well-defined semantics. This is used throughout k8s-manifest-kit for merging configuration values.

### 3.1. Merge Semantics

The deep merge follows these rules:

1. **Maps**: Recursively merged
   - Keys from both base and overlay are preserved
   - Overlapping keys use the overlay value
   - Nested maps are merged recursively

2. **Slices**: Completely replaced (NOT appended or merged)
   - Overlay slice entirely replaces base slice
   - No element-wise merging occurs

3. **Other Types**: Overlay value replaces base value
   - Scalars, structs, and other types are replaced

4. **Type Mismatches**: Overlay wins regardless of types
   - A map can be replaced by a string, or vice versa

5. **Nil Values**: Treated as empty
   - `nil` base returns cloned overlay
   - `nil` overlay returns cloned base
   - Both `nil` returns empty map

**Example - Nested Map Merge**:
```go
base := map[string]any{
    "config": map[string]any{
        "host":    "localhost",
        "port":    8080,
        "timeout": 30,
    },
}

overlay := map[string]any{
    "config": map[string]any{
        "port":    9090,  // Override existing key
        "retries": 3,     // Add new key
    },
}

result := maps.DeepMerge(base, overlay)
// result["config"] = {
//     "host":    "localhost",  // Preserved from base
//     "port":    9090,         // Overridden by overlay
//     "timeout": 30,           // Preserved from base
//     "retries": 3,            // Added by overlay
// }
```

**Example - Slice Replacement**:
```go
base := map[string]any{
    "tags": []string{"dev", "test"},
}

overlay := map[string]any{
    "tags": []string{"prod"},
}

result := maps.DeepMerge(base, overlay)
// result["tags"] = ["prod"]  // NOT ["dev", "test", "prod"]
```

**Example - Type Mismatch**:
```go
base := map[string]any{
    "service": map[string]any{
        "type": "ClusterIP",
        "port": 80,
    },
}

overlay := map[string]any{
    "service": "NodePort",  // String replacing a map
}

result := maps.DeepMerge(base, overlay)
// result["service"] = "NodePort"  // Map completely replaced by string
```

### 3.2. Performance Characteristics

The `DeepMerge` implementation is optimized for performance:

1. **Allocation Efficiency**: Preallocates result map with estimated capacity
2. **Selective Cloning**: Only clones values that won't be immediately replaced
3. **Type-Specific Optimization**: Uses fast type switches for common slice types
   - Fast path for `[]string`, `[]int`, `[]int64`, `[]float64`, `[]bool`
   - Reflection fallback for uncommon slice types
4. **No Shared Memory**: All values are deep cloned to prevent cache pollution

## 4. Caching Architecture (pkg/util/cache)

### 4.1. Overview

The package provides a custom caching implementation with TTL-based expiration. The cache was designed to:

* Provide TTL-based expiration with lazy cleanup
* Prevent cache pollution through automatic deep cloning
* Support generic types for flexibility
* Reduce external dependencies

### 4.2. Cache Interface

```go
// Generic cache interface
type Interface[T any] interface {
    Get(key string) (T, bool)
    Set(key string, value T)
    Sync()  // Triggers lazy expiration of TTL'd entries
}
```

### 4.3. Implementations

**Private `defaultCache[T]`**: Generic TTL-based cache

```go
type defaultCache[T any] struct {
    mu      sync.RWMutex
    entries map[string]entry[T]
    ttl     time.Duration
}

type entry[T any] struct {
    value     T
    expiresAt time.Time
}
```

**Private `renderCache`**: Wrapper for rendering with automatic deep cloning

```go
type renderCache struct {
    cache Interface[[]unstructured.Unstructured]
}

// Automatically clones on Get to prevent external modifications from affecting cache
func (c *renderCache) Get(key string) ([]unstructured.Unstructured, bool) {
    if value, found := c.cache.Get(key); found {
        result := make([]unstructured.Unstructured, len(value))
        for i, obj := range value {
            result[i] = *obj.DeepCopy()
        }
        return result, true
    }
    return nil, false
}

// Automatically clones on Set to prevent caller modifications from affecting cache
func (c *renderCache) Set(key string, value []unstructured.Unstructured) {
    cloned := make([]unstructured.Unstructured, len(value))
    for i, obj := range value {
        cloned[i] = *obj.DeepCopy()
    }
    c.cache.Set(key, cloned)
}
```

### 4.4. Public Constructors

```go
// Create a generic cache with TTL
func New[T any](opts ...Option) Interface[T]

// Create a render-specific cache with automatic deep cloning
func NewRenderCache(opts ...Option) Interface[[]unstructured.Unstructured]
```

### 4.5. Configuration

```go
// Configure TTL (defaults to 5 minutes if not specified or invalid)
cache.WithTTL(10 * time.Minute)

// Usage example
myCache := cache.New[string](cache.WithTTL(5 * time.Minute))
```

### 4.6. Cache Behavior

**TTL Expiration:**
* Entries are marked with expiration time on `Set()`
* Expiration is checked lazily on `Get()` - expired entries return as "not found"
* `Sync()` actively removes expired entries from storage

**Deep Cloning:**
* `renderCache` automatically clones on both `Get()` and `Set()`
* Prevents cache pollution from external modifications
* Caller can safely modify returned objects without affecting cache

**Nil Receiver Safety:**
* `renderCache` methods check for `nil` receiver and handle gracefully
* `Get()` returns `(nil, false)` for nil receiver
* `Set()` and `Sync()` are no-ops for nil receiver
* Defensive programming prevents panics in edge cases

### 4.7. Memory Management

The cache uses a **lazy expiration** strategy that is intentionally designed to balance performance and memory usage:

**Lazy Expiration Design:**

1. **On `Get()`**: Expired entries are detected and treated as "not found", but NOT deleted
   - Avoids write lock acquisition on every `Get()` (read-only operation)
   - Maintains high concurrency with `sync.RWMutex` read locks
   - Expired entries remain in memory until explicitly cleaned up

2. **On `Sync()`**: Expired entries are actively removed from the map
   - Acquires write lock to delete expired entries
   - Should be called periodically to prevent memory growth
   - Not called automatically - application controls cleanup timing

**When to call `Sync()`:**

```go
// Option 1: Periodic cleanup in background
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        cache.Sync()
    }
}()

// Option 2: After batch operations
for i := 0; i < 1000; i++ {
    cache.Set(fmt.Sprintf("key-%d", i), value)
}
cache.Sync()  // Clean up after bulk operations

// Option 3: Application-controlled (e.g., on low memory)
if memoryPressureDetected() {
    cache.Sync()
}
```

**Memory Characteristics:**

* **TTL duration**: Shorter TTL = more frequent expirations, lower memory usage
* **Default TTL**: 5 minutes (configurable via `cache.WithTTL()`)
* **Cleanup frequency**: Application-controlled via `Sync()` calls
* **Memory growth**: Bounded by (number of unique keys) × (entry size) × (time between Sync calls)

For typical workloads with reasonable TTL values (5-10 minutes) and periodic `Sync()` calls, memory growth is minimal and acceptable.

### 4.8. Benefits

1. **Reduced Dependencies**: No longer depends on `k8s.io/client-go/tools/cache`
2. **Type Safety**: Generic interface allows compile-time type checking
3. **Automatic Safety**: Deep cloning prevents accidental cache pollution
4. **Performance**: Lazy expiration avoids background goroutines
5. **Flexibility**: Works with any type via `Interface[T]`

## 5. Kubernetes Utilities (pkg/util/k8s)

Provides utilities for working with Kubernetes `unstructured.Unstructured` objects:

* **Deep Cloning**: Deep clone individual objects or slices of objects to prevent shared state
* **Object Manipulation**: Helper functions for common object operations

## 6. JQ Utilities (pkg/util/jq)

Provides utilities for working with JQ expressions:

* **Expression Compilation**: Compile JQ expressions with variables
* **Expression Execution**: Execute JQ queries against objects
* **Type-Safe Results**: Handle JQ query results with proper type checking

## 7. Error Utilities (pkg/util/errors)

Provides utilities for error handling:

* **Error Wrapping**: Helpers for wrapping errors with context
* **Error Chain Navigation**: Utilities for working with error chains

## 8. Functional Options Pattern (pkg/util/option.go)

The package provides a standard implementation of the functional options pattern used throughout k8s-manifest-kit.

### 8.1. Core Interface

```go
type Option[T any] interface {
    ApplyTo(target *T)
}
```

### 8.2. Function-Based Options

```go
type FunctionalOption[T any] func(*T)

func (f FunctionalOption[T]) ApplyTo(target *T) {
    f(target)
}
```

### 8.3. Usage Pattern

```go
// Define options type
type MyOptions struct {
    Value string
    Count int
}

// Implement ApplyTo for struct-based options
func (opts MyOptions) ApplyTo(target *MyOptions) {
    target.Value = opts.Value
    target.Count = opts.Count
}

// Create function-based option helpers
func WithValue(v string) util.Option[MyOptions] {
    return util.FunctionalOption[MyOptions](func(opts *MyOptions) {
        opts.Value = v
    })
}

func WithCount(c int) util.Option[MyOptions] {
    return util.FunctionalOption[MyOptions](func(opts *MyOptions) {
        opts.Count = c
    })
}
```

## 9. Design Principles

1. **Type Safety**: Leverage Go generics for compile-time type checking
2. **Performance**: Optimize hot paths (caching, merging, cloning)
3. **Simplicity**: Clear, straightforward implementations
4. **Extensibility**: Easy to extend and compose utilities
5. **Minimal Dependencies**: Reduce external dependencies where practical
6. **Defensive Programming**: Handle edge cases gracefully (nil receivers, empty inputs)

