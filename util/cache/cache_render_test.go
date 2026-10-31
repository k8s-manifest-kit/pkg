package cache_test

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/k8s-manifest-kit/pkg/util/cache"

	. "github.com/onsi/gomega"
)

func TestRenderCache(t *testing.T) {

	t.Run("should cache and retrieve results", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		key := "test-key"
		result := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "test",
				},
			}},
		}

		// Initially empty
		_, found := c.Get(key)
		g.Expect(found).To(BeFalse())

		// Set value
		c.Set(key, result)

		// Should find it now
		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached).To(HaveLen(1))
		g.Expect(cached[0].GetKind()).To(Equal("Deployment"))
	})

	t.Run("should automatically clone on Get", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		key := "clone-get-test"
		result := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Service",
				"metadata": map[string]any{
					"name": "test",
				},
			}},
		}

		c.Set(key, result)

		// Get cached result
		cached1, found1 := c.Get(key)
		g.Expect(found1).To(BeTrue())

		// Modify the cached result
		cached1[0].SetName("modified")

		// Get again - should NOT be affected by previous modification due to automatic cloning
		cached2, found2 := c.Get(key)
		g.Expect(found2).To(BeTrue())
		g.Expect(cached2[0].GetName()).To(Equal("test"))
	})

	t.Run("should automatically clone on Set", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		key := "clone-set-test"
		result := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Pod",
				"metadata": map[string]any{
					"name": "original",
				},
			}},
		}

		// Set value
		c.Set(key, result)

		// Modify the original
		result[0].SetName("modified")

		// Get from cache - should have original value due to cloning on Set
		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached[0].GetName()).To(Equal("original"))
	})

	t.Run("should handle empty values", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		key := "empty-key"
		result := make([]unstructured.Unstructured, 0)

		c.Set(key, result)

		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached).To(BeEmpty())
	})

	t.Run("should handle nil values", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(5 * time.Minute))

		key := "nil-key"
		var result []unstructured.Unstructured

		c.Set(key, result)

		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached).To(BeNil())
	})

	t.Run("should expire entries after TTL", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.NewRenderCache(cache.WithTTL(100 * time.Millisecond))

		key := "expiring-key"
		result := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Pod",
				"metadata": map[string]any{
					"name": "pod",
				},
			}},
		}

		c.Set(key, result)

		// Should be cached immediately
		_, found := c.Get(key)
		g.Expect(found).To(BeTrue())

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Sync to trigger expiration
		c.Sync()

		// Should be expired now
		_, found = c.Get(key)
		g.Expect(found).To(BeFalse())
	})
}
