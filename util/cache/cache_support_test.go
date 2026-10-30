package cache_test

import (
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// newCacheTestObject builds a freshly owned unstructured object for cache tests.
// Every call returns an independent object, so workers can hand it to Set or
// mutate a returned clone without sharing state with another goroutine.
func newCacheTestObject(kind string, name string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       kind,
			"metadata": map[string]any{
				"name": name,
			},
		},
	}
}

// runConcurrently starts workerCount workers behind a start barrier and joins
// them before returning. The worker receives its own index and must not assert;
// assertions belong to the calling test goroutine after the workers are joined.
func runConcurrently(workerCount int, worker func(workerIndex int)) {
	start := make(chan struct{})

	var wg sync.WaitGroup
	for workerIndex := range workerCount {
		wg.Go(func() {
			<-start
			worker(workerIndex)
		})
	}

	close(start)
	wg.Wait()
}
