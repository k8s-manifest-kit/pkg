# Shared Packages Design

The `pkg` module contains support code shared by the engine and renderer
modules. Packages are intentionally small and dependency-focused.

## Cache

`cache.Interface[T]` exposes `Get(key any) (T, bool)`, `Set(key any, value T)`,
and `Sync()`. `cache.New` accepts options including `cache.WithTTL` and
`cache.WithKeyFunc`; the default TTL is five minutes. Expired entries are
ignored by `Get` and removed by `Sync`.

`cache.New` synchronizes access to its entries but performs no copying:
values passed to `Set` are stored as given and returned as given by `Get`.
Those values stay caller-owned — callers must not mutate them concurrently
with `Set` or `Get`. Keys are converted with the configured `KeyFunc`;
anything stronger, such as partitioning keys per render source, is the
caller's responsibility and not a generic-cache guarantee.

`NewRenderCache` stores render results safely for reuse. It deep-copies
`[]unstructured.Unstructured` at both the `Set` and `Get` boundaries, so
mutating a returned slice or its objects does not affect cached state, and
mutating the stored input after `Set` does not alter the cache.

## Data helpers

`maps.DeepMerge` recursively merges maps. Nested maps are merged, while slices
and other non-map values from the overlay replace the corresponding base
value.

`maps.DeepCloneMap` and `maps.DeepCloneValue` clone JSON-like trees: nested
`map[string]any` and `[]any` are cloned recursively, the common typed slices
(`[]string`, `[]int`, `[]int64`, `[]float64`, `[]bool`) are copied, and other
slice types are shallow-copied element by element. All other values —
primitives, pointers, structs — are retained as-is, so aliasing through
pointer or struct fields survives the clone. There is no cycle detection;
the helpers bound what they clone to the types above rather than isolating
arbitrary object graphs.

`k8s` contains helpers for Kubernetes objects and metadata, and `jq`
provides JQ-backed transformations.

## Metrics and errors

Metrics are optional and can be carried through a context with
`metrics.WithMetrics` and read with `metrics.FromContext`. The package exposes
render and renderer observations and supports no-op operation when metrics are
not configured. Shared error helpers preserve causes while adding operation
context.

Functional option helpers live in [`util/option.go`](../util/option.go)
(`Option[T]` and `FunctionalOption[T]`) in the `util` package; there is no
separate `util/option` directory.

## Related documentation

- [`../AGENTS.md`](../AGENTS.md)
- [`development.md`](development.md)
- [`../../engine/docs/design.md`](../../engine/docs/design.md)
