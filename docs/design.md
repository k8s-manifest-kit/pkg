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

`maps.ValidateAcyclic` reports reference cycles among exactly the
`map[string]any` and `[]any` nodes those helpers traverse; every other type
is a leaf, so shared references that do not form a cycle pass, and retained
values of other types are outside its guarantee. Only nodes on one descent
path are compared, so a slice holding an acyclic prefix of itself remains
valid. The clone and merge helpers themselves still recurse without built-in
cycle validation, so callers who accept trees that may be cyclic validate
before copying them.

`k8s` contains helpers for Kubernetes objects and metadata, and `jq`
provides JQ-backed transformations.

`k8s.DecodeYAML` decodes multi-document YAML permissively: empty documents
and documents without nonempty string `kind` and `apiVersion` are skipped,
while syntax, decode, and conversion failures fail with the zero-based
index of the offending document. This is document metadata checking, not
Kubernetes schema validation; Helm strict mode concerns template evaluation
and is unrelated to this permissive contract.

## Metrics and errors

Metrics are optional and can be carried through a context with
`metrics.WithMetrics` and read with `metrics.FromContext`. The package exposes
render and renderer observations and supports no-op operation when metrics are
not configured. Shared error helpers preserve causes while adding operation
context.

Functional option helpers live in [`util/option`](../util/option)
(`Option[T]` and `FunctionalOption[T]`), and slice conversion helpers such as
`ToAnySlice` live in [`util/slice`](../util/slice). There is no mixed root
`util` Go package.

## HTTP clients

The factories live in [`util/httpclient`](../util/httpclient):
`httpclient.NewClient` and `httpclient.NewTransport` construct private HTTP
clients and transports. Every call returns a fresh instance with a fresh
`net.Dialer`; nothing is read from, cloned from, or shared with
`http.DefaultClient`, `http.DefaultTransport`, or any other package-level
mutable HTTP state. This is an explicit difference from Go's pooled globals:
instances are independent, so mutating one — or closing its idle
connections — cannot affect any other instance.

The transports are deliberately non-pooled: keep-alives are disabled
(`DisableKeepAlives: true`, `MaxIdleConnsPerHost: -1`), so every request
uses its own connection and no connection lingers idle after a response.
The remaining baseline fields follow the standard library defaults: the
proxy is resolved from the process environment, connection establishment is
bounded at 30 seconds with a 30 second TCP keep-alive interval, HTTP/2 is
attempted when the connection supports it, the idle connection bounds are
100 connections with a 90 second idle timeout, the TLS handshake is bounded
at 10 seconds, and the expect-continue timeout is 1 second. TLS
verification keeps its zero value: system roots, no custom shared TLS
configuration.

The client timeout stays zero deliberately. Callers that need a deadline
own the request context. The client also has no cookie jar and no custom
redirect policy. Callers own the lifecycle of everything they construct:
always close response bodies, and close idle connections on instances they
created when finished with them.

## Related documentation

- [`../AGENTS.md`](../AGENTS.md)
- [`development.md`](development.md)
- [`../../engine/docs/design.md`](../../engine/docs/design.md)
