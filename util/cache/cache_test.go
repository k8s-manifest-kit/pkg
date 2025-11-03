package cache_test

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/k8s-manifest-kit/pkg/util/cache"

	. "github.com/onsi/gomega"
)

func TestCache(t *testing.T) {

	t.Run("should cache and retrieve results", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(5 * time.Minute))

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

	t.Run("should NOT clone cached results", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(5 * time.Minute))

		key := "clone-test"
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

		// Get again - should be affected by previous modification since no deep clone
		cached2, found2 := c.Get(key)
		g.Expect(found2).To(BeTrue())
		g.Expect(cached2[0].GetName()).To(Equal("modified"))
	})

	t.Run("should handle different keys separately", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(5 * time.Minute))

		key1 := "key1"
		key2 := "key2"

		result1 := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "deployment",
				},
			}},
		}

		result2 := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Service",
				"metadata": map[string]any{
					"name": "service",
				},
			}},
		}

		c.Set(key1, result1)
		c.Set(key2, result2)

		cached1, found1 := c.Get(key1)
		g.Expect(found1).To(BeTrue())
		g.Expect(cached1[0].GetKind()).To(Equal("Deployment"))

		cached2, found2 := c.Get(key2)
		g.Expect(found2).To(BeTrue())
		g.Expect(cached2[0].GetKind()).To(Equal("Service"))
	})

	t.Run("should expire entries after TTL", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(100 * time.Millisecond))

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

		// Should be expired now
		_, found = c.Get(key)
		g.Expect(found).To(BeFalse())
	})

	t.Run("should handle empty values", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(5 * time.Minute))

		key := "empty-key"
		result := make([]unstructured.Unstructured, 0)

		c.Set(key, result)

		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached).To(BeEmpty())
	})

	t.Run("should handle nil values", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(5 * time.Minute))

		key := "nil-key"
		var result []unstructured.Unstructured

		c.Set(key, result)

		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached).To(BeNil())
	})

	t.Run("should use default TTL if invalid", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(0))
		g.Expect(c).ToNot(BeNil())

		c = cache.New[[]unstructured.Unstructured](cache.WithTTL(-10 * time.Second))
		g.Expect(c).ToNot(BeNil())
	})

	t.Run("should update existing entry", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[[]unstructured.Unstructured](cache.WithTTL(5 * time.Minute))

		key := "update-key"

		result1 := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "v1",
				},
			}},
		}

		result2 := []unstructured.Unstructured{
			{Object: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "v2",
				},
			}},
		}

		c.Set(key, result1)
		c.Set(key, result2)

		// Should have the updated value
		cached, found := c.Get(key)
		g.Expect(found).To(BeTrue())
		g.Expect(cached[0].GetName()).To(Equal("v2"))
	})
}

func TestCacheKeyFunc(t *testing.T) {

	t.Run("should work with string keys", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[string](cache.WithTTL(5 * time.Minute))

		c.Set("key1", "value1")
		c.Set("key2", "value2")

		val1, found1 := c.Get("key1")
		g.Expect(found1).To(BeTrue())
		g.Expect(val1).To(Equal("value1"))

		val2, found2 := c.Get("key2")
		g.Expect(found2).To(BeTrue())
		g.Expect(val2).To(Equal("value2"))
	})

	t.Run("should work with int keys", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[string](cache.WithTTL(5 * time.Minute))

		c.Set(1, "value1")
		c.Set(2, "value2")

		val1, found1 := c.Get(1)
		g.Expect(found1).To(BeTrue())
		g.Expect(val1).To(Equal("value1"))

		val2, found2 := c.Get(2)
		g.Expect(found2).To(BeTrue())
		g.Expect(val2).To(Equal("value2"))
	})

	t.Run("should work with int64 keys", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[string](cache.WithTTL(5 * time.Minute))

		c.Set(int64(100), "value100")
		c.Set(int64(200), "value200")

		val1, found1 := c.Get(int64(100))
		g.Expect(found1).To(BeTrue())
		g.Expect(val1).To(Equal("value100"))

		val2, found2 := c.Get(int64(200))
		g.Expect(found2).To(BeTrue())
		g.Expect(val2).To(Equal("value200"))
	})

	t.Run("should work with struct keys using default KeyFunc", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[string](cache.WithTTL(5 * time.Minute))

		type testSpec struct {
			Name    string
			Version int
		}

		spec1 := testSpec{Name: "app1", Version: 1}
		spec2 := testSpec{Name: "app2", Version: 2}

		c.Set(spec1, "result1")
		c.Set(spec2, "result2")

		val1, found1 := c.Get(spec1)
		g.Expect(found1).To(BeTrue())
		g.Expect(val1).To(Equal("result1"))

		val2, found2 := c.Get(spec2)
		g.Expect(found2).To(BeTrue())
		g.Expect(val2).To(Equal("result2"))
	})

	t.Run("should use custom KeyFunc when provided", func(t *testing.T) {
		g := NewWithT(t)

		type customSpec struct {
			Path    string
			Version string
		}

		// Custom KeyFunc that only uses Path
		customKeyFunc := func(key any) string {
			if spec, ok := key.(customSpec); ok {
				return spec.Path
			}

			return ""
		}

		c := cache.New[string](
			cache.WithTTL(5*time.Minute),
			cache.WithKeyFunc(customKeyFunc),
		)

		spec1 := customSpec{Path: "/path1", Version: "v1"}
		spec2 := customSpec{Path: "/path1", Version: "v2"} // Same path, different version

		c.Set(spec1, "result1")
		c.Set(spec2, "result2") // Should overwrite result1 because KeyFunc ignores Version

		// Both specs should return the same result because they have the same Path
		val1, found1 := c.Get(spec1)
		g.Expect(found1).To(BeTrue())
		g.Expect(val1).To(Equal("result2"))

		val2, found2 := c.Get(spec2)
		g.Expect(found2).To(BeTrue())
		g.Expect(val2).To(Equal("result2"))
	})

	t.Run("should handle custom KeyFunc with multiple types", func(t *testing.T) {
		g := NewWithT(t)

		type spec1 struct {
			Name string
		}
		type spec2 struct {
			ID int
		}

		customKeyFunc := func(key any) string {
			switch k := key.(type) {
			case spec1:
				return "spec1:" + k.Name
			case spec2:
				return "spec2:" + string(rune(k.ID))
			case string:
				return "string:" + k
			default:
				return "unknown"
			}
		}

		c := cache.New[string](
			cache.WithTTL(5*time.Minute),
			cache.WithKeyFunc(customKeyFunc),
		)

		c.Set(spec1{Name: "test"}, "value1")
		c.Set(spec2{ID: 42}, "value2")
		c.Set("direct", "value3")

		val1, found1 := c.Get(spec1{Name: "test"})
		g.Expect(found1).To(BeTrue())
		g.Expect(val1).To(Equal("value1"))

		val2, found2 := c.Get(spec2{ID: 42})
		g.Expect(found2).To(BeTrue())
		g.Expect(val2).To(Equal("value2"))

		val3, found3 := c.Get("direct")
		g.Expect(found3).To(BeTrue())
		g.Expect(val3).To(Equal("value3"))
	})

	t.Run("should use default KeyFunc when nil is provided", func(t *testing.T) {
		g := NewWithT(t)
		c := cache.New[string](
			cache.WithTTL(5*time.Minute),
			cache.WithKeyFunc(nil), // Explicitly pass nil
		)

		// Should still work with default behavior
		c.Set("test", "value")
		val, found := c.Get("test")
		g.Expect(found).To(BeTrue())
		g.Expect(val).To(Equal("value"))
	})
}

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
