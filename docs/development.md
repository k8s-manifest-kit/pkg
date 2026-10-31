# Shared Packages Development

## Prerequisites and commands

Build with the Go version declared in [`go.mod`](../go.mod). The Makefile
defines the supported workflow:

```bash
make test
make test/race
make bench
make fmt
make lint
make lint/fix
make check
```

`make test` runs the unit tests and `make test/race` repeats them with the
race detector. `make check` runs lint and vulncheck only; it is not a
substitute for either test target. Formatting is applied separately through
`make fmt`, and CI fails when committed code is not formatted.

Run focused package tests with the normal Go test tooling when diagnosing a
failure, but keep repository-wide validation on the Makefile targets. CI runs
the unit tests with the race detector through `make test/race`; run it locally
before committing concurrency-sensitive changes.

## Testing conventions

Tests use Gomega and live in external `_test` packages that exercise the
public API; use `t.Context()` for context-aware tests. Wrap errors with
operation context using `%w` where a cause exists, check sentinel errors with
`errors.Is`, and word validation messages so they name the path, key, or type
needed for diagnosis rather than echoing input contents.

Cache tests should cover custom key functions, TTL expiry, `Sync`, and clone
isolation. Keep the two cache flavors distinct: `cache.New` stores values as
given, so tests must hand it values that no other goroutine mutates;
`NewRenderCache` deep-copies unstructured objects at the `Get` and `Set`
boundaries. Prove returned-object independence by actually mutating a
returned object — the specific nested field under test — and then checking
that the cached source and other `Get` results are unchanged. Use observable
evidence such as found flags and returned values, not call counts.

Clone tests should cover the documented behavior of `DeepCloneMap` and
`DeepCloneValue` type by type: nested maps and `[]any`, the common typed
slices, the shallow copy applied to other slice types, and values retained
as-is such as pointers and structs. Do not write tests that imply cycle
handling or isolation of arbitrary object graphs.

Merge tests should distinguish nested maps from slices and scalar
replacement. Metrics tests should verify both configured and no-op behavior.

Concurrency tests start workers behind a barrier, join them, and assert only
from the test goroutine, using deterministic keys and per-worker owned values
instead of sleeps to manufacture contention or expiry. Anything a caller
supplies to a shared cache — including key functions passed with
`WithKeyFunc` — runs under overlapping calls and must be safe for concurrent
use.

Benchmarks belong beside the implementation they measure and should avoid
sharing mutable objects between iterations.

## Code organization and Go style

Declare each parameter type explicitly. Do not group parameters that share a
type: write `apiVersion string, kind string`, not `apiVersion, kind string`.

Put helpers shared across a package's implementation in `_support.go` and
test-only helpers in `_support_test.go` (see
`util/cache/cache_support_test.go`); keep cohesive feature files intact rather
than extracting every small helper. Import grouping and formatting follow
`.golangci.yml`; run `make fmt` rather than hand-formatting.

### Multiline strings

Declare every multiline string — fixtures, manifests, templates, and expected
output — as a single descriptive, standalone, package-level `const name =`
declaration with a raw string literal.

Never embed multiline string literals in function or method bodies, and never
declare them as function-local constants. Never group multiline constants in
a `const (...)` block; related single-line scalar constants may still be
grouped, and single-line inline strings are unaffected.

Keep large literal content separate from execution logic for readability.
Preserve meaningful content indentation; the declaration layout must not
introduce unintended bytes. A function may format the named template into a
local variable — the rule concerns where the literal is declared, not where
its string value is used.

```go
const fixtureYAML = `apiVersion: v1
kind: ConfigMap
`
```

See [`design.md`](design.md) and [`../AGENTS.md`](../AGENTS.md).
