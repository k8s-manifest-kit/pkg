# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The `pkg` module provides core utilities for the k8s-manifest-kit ecosystem. It contains shared functionality for caching, value merging, error handling, JQ expressions, and Kubernetes object manipulation that is used across multiple k8s-manifest-kit components.

**Part of the [k8s-manifest-kit](https://github.com/k8s-manifest-kit) organization.**

## Documentation

- **[README.md](README.md)** - Module overview and quick start
- **[docs/design.md](docs/design.md)** - Architecture, design decisions, and implementation details
- **[docs/development.md](docs/development.md)** - Coding conventions, testing guidelines, and contribution guide

## Core Utilities

### Cache (util/cache)

TTL-based caching with automatic deep cloning:
- Generic `Interface[T]` for any type
- `NewRenderCache()` for Kubernetes objects with automatic cloning
- Lazy expiration with manual `Sync()` cleanup
- Configurable TTL via `WithTTL()`

See [docs/design.md#4-caching-architecture](docs/design.md#4-caching-architecture) for complete documentation.

### Maps Utilities (util/maps)

Deep cloning and merging of nested `map[string]any` structures:
- `DeepCloneMap` / `DeepCloneValue` for fully independent copies of JSON-like trees
- `DeepMerge` for recursive map merging (preserves keys from both sides)
- Slice replacement (not appending)
- Type-safe handling of mismatches
- Performance-optimized with selective cloning and fast typed-slice paths

```go
clone := maps.DeepCloneMap(original)
result := maps.DeepMerge(base, overlay)
```

See [docs/design.md#3-value-merging-strategy](docs/design.md#3-value-merging-strategy) for semantics and examples.

### Kubernetes Utilities (util/k8s)

Helpers for working with `unstructured.Unstructured` objects:
- Deep cloning of objects and slices
- Object manipulation utilities

### JQ Utilities (util/jq)

JQ expression compilation and execution:
- Expression compilation with variables
- Type-safe query results

### Error Utilities (util/errors)

Error handling helpers:
- Error wrapping with context
- Error chain navigation

### Functional Options (util/option.go)

Standard functional options pattern implementation:
```go
type Option[T any] interface {
    ApplyTo(target *T)
}

type FunctionalOption[T any] func(*T)
```

Used consistently across k8s-manifest-kit for flexible configuration.

See [docs/design.md#8-functional-options-pattern](docs/design.md#8-functional-options-pattern) for usage patterns.

## Development

**Run tests:**
```bash
make test
```

**Format and lint:**
```bash
make fmt
make lint
```

**Run benchmarks:**
```bash
go test -v ./util/... -run=^$ -bench=.
```

For detailed development information:
- **Build commands**: See [docs/development.md#setup-and-build](docs/development.md#setup-and-build)
- **Coding conventions**: See [docs/development.md#coding-conventions](docs/development.md#coding-conventions)
- **Testing guidelines**: See [docs/development.md#testing-guidelines](docs/development.md#testing-guidelines)
- **Code review guidelines**: See [docs/development.md#code-review-guidelines](docs/development.md#code-review-guidelines)

## Testing Conventions

- Use vanilla Gomega (dot import)
- All test data as package-level constants
- Benchmark naming: `Benchmark<Utility><TestName>`
- Use `t.Context()` instead of `context.Background()`

See [docs/development.md#testing-guidelines](docs/development.md#testing-guidelines) for complete testing practices.

## Key Design Principles

1. **Type Safety**: Leverage Go generics for compile-time checking
2. **Performance**: Optimize hot paths (caching, merging, cloning)
3. **Simplicity**: Clear, straightforward implementations
4. **Composability**: Utilities compose well together
5. **Minimal Dependencies**: Reduce external dependencies where practical
6. **Defensive Programming**: Handle edge cases gracefully (nil receivers, empty inputs)

See [docs/design.md#9-design-principles](docs/design.md#9-design-principles) for more details.

