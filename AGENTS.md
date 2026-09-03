# Agent Guide: pkg

`pkg` contains shared utilities used by the k8s-manifest-kit renderers and engine.

## Documentation

- [README](README.md) — module overview.
- [Design](docs/design.md) — utility behavior and design decisions.
- [Development](docs/development.md) — workflow and testing conventions.

## Packages and API

- `util/cache`: generic TTL cache via `cache.New[T]`, plus `NewRenderCache` for deep-cloned Kubernetes objects. Keys are arbitrary values converted by `WithKeyFunc`; the default TTL is five minutes.
- `util/maps`: `DeepMerge`, `DeepCloneMap`, and `DeepCloneValue`. Maps merge recursively; slices and non-map values from the overlay replace the base.
- `util/k8s`: YAML decoding, unstructured conversion, metadata helpers, and content hashing.
- `util/jq`: JQ engine with custom variables and functions.
- `util/metrics`: context-carried engine and renderer metrics, with memory and no-op implementations.
- `util/errors`: shared error helpers.
- `util/option.go`: generic `Option[T]` and `FunctionalOption[T]` definitions.

Use the module imports under `github.com/k8s-manifest-kit/pkg/util/...`. Do not reference the removed `util/option` directory.

## Development

Run commands from this directory:

```bash
make test
make fmt
make lint
make lint/fix
make check
```

Tests use vanilla Gomega and `t.Context()`. Keep reusable YAML/JSON test data at package scope, preserve error wrapping with `%w`, and add benchmarks for hot paths such as cache, clone, and merge operations.

