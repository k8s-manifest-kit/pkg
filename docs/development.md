# Shared Packages Development

## Prerequisites and commands

Use Go 1.26.8. The Makefile defines the supported workflow:

```bash
make test
make fmt
make lint
make lint/fix
make check
```

Run focused package tests with the normal Go test tooling when diagnosing a
failure, but keep repository-wide validation on the Makefile targets.

## Testing conventions

Use `t.Context()` for context-aware tests and wrap errors with enough
operation context to make failures actionable. Cache tests should cover
custom key functions, TTL expiry, `Sync`, and clone isolation. Merge tests
should distinguish nested maps from slices and scalar replacement. Metrics
tests should verify both configured and no-op behavior.

Benchmarks belong beside the implementation they measure and should avoid
sharing mutable objects between iterations.

See [`design.md`](design.md) and [`../AGENTS.md`](../AGENTS.md).
