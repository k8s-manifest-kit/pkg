# Shared Packages

`pkg` contains reusable support packages for the manifest renderers and the
engine: caches, Kubernetes object helpers, JSONPath/JQ helpers, metrics,
error utilities, and functional options.

Notable packages include:

- `cache` for TTL-aware generic caches and render-object caches.
- `k8s` for object and metadata helpers.
- `maps` for recursive map merging.
- `jq` for JQ-based object transformations.
- `metrics` for optional render metrics.
- `errors` and `option.go` for shared error and option utilities.

See [`docs/design.md`](docs/design.md), [`docs/development.md`](docs/development.md),
and [`AGENTS.md`](AGENTS.md).

## License

Apache License 2.0. See [LICENSE](LICENSE).
