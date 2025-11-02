# Development Guide: pkg

This document provides coding conventions, testing guidelines, and contribution practices for developing the pkg utilities module.

For architectural information and design decisions, see [design.md](design.md).

## Table of Contents

1. [Setup and Build](#setup-and-build)
2. [Coding Conventions](#coding-conventions)
3. [Testing Guidelines](#testing-guidelines)
4. [Code Review Guidelines](#code-review-guidelines)

## Setup and Build

### Build Commands

```bash
# Run all tests
make test

# Format code
make fmt

# Run linter
make lint

# Fix linting issues automatically
make lint/fix

# Clean build artifacts and test cache
make clean

# Update dependencies
make deps
```

### Test Commands

```bash
# Run all tests with verbose output
go test -v ./...

# Run tests in a specific package
go test -v ./util/cache

# Run a specific test
go test -v ./util/cache -run TestCacheTTL

# Run benchmarks
go test -v ./util/... -run=^$ -bench=.
```

## Coding Conventions

### Functional Options Pattern

All struct initialization uses the functional options pattern for flexible, extensible configuration.

**Define Options as Interfaces:**
```go
type Option[T any] interface {
    ApplyTo(target *T)
}
```

**Provide Both Function-Based and Struct-Based Options:**
```go
// Function-based option
func WithTTL(ttl time.Duration) Option {
    return util.FunctionalOption[cacheOptions](func(opts *cacheOptions) {
        opts.ttl = ttl
    })
}

// Struct-based option for bulk configuration
type CacheOptions struct {
    TTL time.Duration
}

func (opts CacheOptions) ApplyTo(target *cacheOptions) {
    target.ttl = opts.TTL
}
```

**Guidelines:**
- For slice/map fields in struct-based options, use the type directly (not pointers)
- Place all options and related methods in `*_option.go` files
- Provide both patterns to support different use cases

**Usage:**
```go
// Function-based (flexible, composable)
cache.New[string](
    cache.WithTTL(5 * time.Minute),
)

// Struct-based (bulk configuration via literals)
cache.New[string](&cache.CacheOptions{
    TTL: 5 * time.Minute,
})
```

### Error Handling Conventions

* Errors are wrapped using `fmt.Errorf` with `%w` for proper error chain propagation
* Context is passed through all functions for cancellation support
* First error encountered stops processing and is returned immediately
* All constructors validate inputs and return errors
* Use `errors.As()` to extract typed errors from error chains
* Use `errors.Is()` to check for specific underlying errors

### Package Organization

```
pkg/
├── util/           # Utility functions
│   ├── cache/      # Caching utilities
│   ├── errors/     # Error handling
│   ├── jq/         # JQ expression handling
│   ├── k8s/        # Kubernetes object utilities
│   ├── metrics/    # Metrics collection
│   ├── merge.go    # Value merging
│   └── option.go   # Functional options
```

Each utility follows the pattern:
- `utility.go` - Main implementation
- `utility_option.go` - Functional options (if needed)
- `utility_test.go` - Tests

## Testing Guidelines

### Test Framework

- Use vanilla Gomega (not Ginkgo)
- Use dot imports for Gomega: `import . "github.com/onsi/gomega"`
- Prefer `Should` over `To`
- For error validation: `Should(HaveOccurred())` / `ShouldNot(HaveOccurred())`
- Use subtests (`t.Run`) for organizing related test cases
- Use `t.Context()` instead of `context.Background()` or `context.TODO()` (Go 1.24+)

**Example:**
```go
func TestCache(t *testing.T) {
    g := NewWithT(t)
    ctx := t.Context()

    t.Run("should cache values", func(t *testing.T) {
        c := cache.New[string]()
        c.Set("key", "value")
        
        value, found := c.Get("key")
        g.Expect(found).Should(BeTrue())
        g.Expect(value).Should(Equal("value"))
    })
}
```

### Test Data Organization

**CRITICAL**: All test data must be defined as package-level constants, never inline within test methods.

**Good:**
```go
const testYAML = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`

func TestSomething(t *testing.T) {
    // Use testYAML constant
}
```

**Bad:**
```go
func TestSomething(t *testing.T) {
    yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`  // WRONG: inline test data
}
```

**Rules:**
- ALL test data (YAML, JSON, strings, etc.) must be package-level constants
- Define constants at the top of test files, grouped by test scenario
- Use descriptive names that indicate purpose (e.g., `testConfigMapYAML`, `testMergeBase`)
- Add comments to group related constants (e.g., `// Test constants for merge tests`)
- This makes tests more readable and data reusable across tests

### Benchmark Naming

- Include utility name in benchmark tests
- Format: `Benchmark<Utility><TestName>`
- Examples: `BenchmarkCacheGet`, `BenchmarkDeepMerge`, `BenchmarkCloneUnstructured`

### Test Strategy

**Unit Tests**: Test each component in isolation
- Cache: Test TTL expiration, Get/Set behavior, Sync cleanup
- Merge: Test nested maps, slices, type mismatches, nil handling
- Clone: Test deep cloning preserves structure and prevents shared state
- JQ: Test expression compilation and execution

**Benchmark Tests**: Performance testing
- Named with utility prefix: `BenchmarkCacheGet`, `BenchmarkDeepMerge`
- Test hot paths (Get/Set, merge, clone operations)
- Measure allocation overhead

**Test Patterns**:
- Use vanilla Gomega (no Ginkgo)
- Subtests via `t.Run()`
- Use `t.Context()` instead of `context.Background()`
- Test edge cases (nil, empty, invalid inputs)

## Code Review Guidelines

### Linter Rules

All code must pass `make lint` before submission. Key linter rules:

- **goconst**: Extract repeated string literals to constants
- **gosec**: No hardcoded secrets (use `//nolint:gosec` only for test data with comment explaining why)
- **staticcheck**: Follow all suggestions
- **Comment formatting**: All comments must end with periods

### Git Commit Conventions

**Commit Message Format:**
```
<type>: <subject>

<body>

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring (no functional changes)
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `build`: Build system or dependency changes
- `chore`: Maintenance tasks

**Subject:**
- Use imperative mood ("add feature" not "added feature")
- Don't capitalize first letter
- No period at the end
- Max 72 characters

**Body:**
- Explain what and why (not how)
- Separate from subject with blank line
- Wrap at 72 characters
- Use bullet points for multiple items

**Example:**
```
feat: add TTL-based cache with deep cloning

This commit introduces a generic TTL-based cache implementation:

- Generic interface supporting any type via Go generics
- Lazy expiration for performance (cleanup on Sync())
- Automatic deep cloning for render cache to prevent pollution
- Configurable TTL via functional options

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Pull Request Checklist

Before submitting a PR:
- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code formatted (`make fmt`)
- [ ] New tests added for new features
- [ ] Documentation updated (design.md or development.md as needed)
- [ ] All test data extracted to package-level constants
- [ ] Benchmark tests follow naming convention
- [ ] Error handling follows conventions
- [ ] Functional options pattern used for configuration

### Code Style

- **Function signatures**: Each parameter must have its own type declaration (never group parameters with same type)
- **Comments**: Explain *why*, not *what*. Focus on non-obvious behavior, edge cases, and relationships
- **Error wrapping**: Always use `fmt.Errorf` with `%w` for error chains
- **Context propagation**: Pass context through all layers for cancellation support
- **Zero values**: Leverage zero value semantics instead of pointers where appropriate
- **Generics**: Use type parameters for reusable utilities that work with multiple types

## Extensibility

### Adding New Utilities

1. Create a new file in `pkg/util/<utility>/` (or `pkg/util/<utility>.go` for simple utilities)
2. Define clear interfaces for public API
3. Implement functionality with proper error handling
4. Add comprehensive tests covering edge cases
5. Add benchmarks for performance-critical code
6. Document design decisions in `docs/design.md`
7. Update this guide with usage patterns if needed

**Example Structure:**
```go
// pkg/util/myutil/myutil.go
package myutil

import "context"

// Public interface
type Interface interface {
    Process(ctx context.Context, input string) (string, error)
}

// Implementation
type processor struct {
    opts *Options
}

func New(opts ...Option) Interface {
    // Initialize with options
}

func (p *processor) Process(ctx context.Context, input string) (string, error) {
    // Implementation
}
```

### Best Practices

1. **Single Responsibility**: Each utility should do one thing well
2. **Composability**: Utilities should compose well together
3. **Performance**: Optimize hot paths, measure with benchmarks
4. **Safety**: Handle nil receivers and empty inputs gracefully
5. **Documentation**: Document non-obvious behavior and design decisions
6. **Testing**: High test coverage with edge cases

