package cache_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/k8s-manifest-kit/pkg/util/cache"

	. "github.com/onsi/gomega"
)

func TestCacheConcurrency(t *testing.T) {

	t.Run("should support overlapping Get, Set, and Sync on a shared generic cache", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[string](cache.WithTTL(5 * time.Minute))

		const workers = 8
		const iterations = 50

		// Every worker owns its keys and writes immutable string values, so the
		// final content of each key is deterministic despite the overlapping calls.
		runConcurrently(workers, func(workerIndex int) {
			key := fmt.Sprintf("worker-%d", workerIndex)
			value := strconv.Itoa(workerIndex)

			for i := range iterations {
				iterationKey := fmt.Sprintf("%s-iteration-%d", key, i)

				c.Set(key, value)
				c.Set(iterationKey, value)
				_, _ = c.Get(key)
				_, _ = c.Get(iterationKey)
				c.Sync()
			}
		})

		for workerIndex := range workers {
			key := fmt.Sprintf("worker-%d", workerIndex)
			value := strconv.Itoa(workerIndex)

			cached, found := c.Get(key)
			g.Expect(found).To(BeTrue())
			g.Expect(cached).To(Equal(value))

			lastIterationKey := fmt.Sprintf("%s-iteration-%d", key, iterations-1)
			cachedIteration, foundIteration := c.Get(lastIterationKey)
			g.Expect(foundIteration).To(BeTrue())
			g.Expect(cachedIteration).To(Equal(value))
		}
	})
}

func TestRenderCacheConcurrency(t *testing.T) {

	t.Run("should support overlapping Get, Set, and Sync on a shared render cache", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		const workers = 8
		const iterations = 25

		// Every worker owns its key and sets freshly built objects, so Set never
		// clones an object that another goroutine mutates. Mutating a result
		// returned by Get is safe because each Get returns an independent clone.
		runConcurrently(workers, func(workerIndex int) {
			key := fmt.Sprintf("render-worker-%d", workerIndex)
			name := fmt.Sprintf("object-%d", workerIndex)

			for range iterations {
				c.Set(key, []unstructured.Unstructured{newCacheTestObject("Deployment", name)})

				cached, found := c.Get(key)
				if found {
					cached[0].SetName("mutated-by-worker")
				}

				c.Sync()
			}
		})

		for workerIndex := range workers {
			key := fmt.Sprintf("render-worker-%d", workerIndex)
			name := fmt.Sprintf("object-%d", workerIndex)

			cached, found := c.Get(key)
			g.Expect(found).To(BeTrue())
			g.Expect(cached).To(HaveLen(1))
			g.Expect(cached[0].GetKind()).To(Equal("Deployment"))
			g.Expect(cached[0].GetName()).To(Equal(name))
		}
	})

	t.Run("should return independently mutable results after concurrent use", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		const workers = 8
		const iterations = 25
		key := "shared-render-key"
		name := "shared-object"

		runConcurrently(workers, func(_ int) {
			for range iterations {
				// All workers store the same deterministic content under the
				// shared key, so the stored snapshot is known once they join.
				c.Set(key, []unstructured.Unstructured{newCacheTestObject("Deployment", name)})
				_, _ = c.Get(key)
				c.Sync()
			}
		})

		first, foundFirst := c.Get(key)
		g.Expect(foundFirst).To(BeTrue())
		g.Expect(first).To(HaveLen(1))

		second, foundSecond := c.Get(key)
		g.Expect(foundSecond).To(BeTrue())
		g.Expect(second).To(HaveLen(1))

		first[0].SetName("mutated-first")
		first[0].SetKind("MutatedKind")

		g.Expect(second[0].GetName()).To(Equal(name))
		g.Expect(second[0].GetKind()).To(Equal("Deployment"))

		stored, foundStored := c.Get(key)
		g.Expect(foundStored).To(BeTrue())
		g.Expect(stored[0].GetName()).To(Equal(name))
		g.Expect(stored[0].GetKind()).To(Equal("Deployment"))
	})
}
