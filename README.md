# Shared Packages

`pkg` contains reusable support packages for the manifest renderers and the
engine: caches, Kubernetes object helpers, JSONPath/JQ helpers, metrics,
error utilities, functional options, and HTTP client factories.

Notable packages include:

- `cache` for TTL-aware generic caches and render-object caches.
- `httpclient` for private, non-pooled HTTP clients and transports.
- `k8s` for object and metadata helpers.
- `maps` for recursive map merging.
- `jq` for JQ-based object transformations.
- `metrics` for optional render metrics.
- `errors`, `option`, and `slice` for shared error, functional-option, and slice utilities.

See [`docs/design.md`](docs/design.md), [`docs/development.md`](docs/development.md),
and [`AGENTS.md`](AGENTS.md).

## License

Apache License 2.0. See [LICENSE](LICENSE).
