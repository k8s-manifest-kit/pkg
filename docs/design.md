# Shared Packages Design

The `pkg` module contains support code shared by the engine and renderer
modules. Packages are intentionally small and dependency-focused.

## Cache

`cache.Interface[T]` exposes `Get(key any) (T, bool)`, `Set(key any, value T)`,
and `Sync()`. `cache.New` accepts options including `cache.WithTTL` and
`cache.WithKeyFunc`; the default TTL is five minutes. Expired entries are
ignored by `Get` and removed by `Sync`.

`NewRenderCache` stores render results safely for reuse. Render objects are
deep-copied at cache boundaries so callers cannot mutate cached state.

## Data helpers

`maps.DeepMerge` recursively merges maps. Nested maps are merged, while slices
and other non-map values from the overlay replace the corresponding base
value. `k8s` contains helpers for Kubernetes objects and metadata, and `jq`
provides JQ-backed transformations.

## Metrics and errors

Metrics are optional and can be carried through a context with
`metrics.WithMetrics` and read with `metrics.FromContext`. The package exposes
render and renderer observations and supports no-op operation when metrics are
not configured. Shared error helpers preserve causes while adding operation
context.

Functional option helpers live in `pkg/option.go`; there is no `util/option`
directory.

## Related documentation

- [`../AGENTS.md`](../AGENTS.md)
- [`development.md`](development.md)
- [`../../engine/docs/design.md`](../../engine/docs/design.md)
