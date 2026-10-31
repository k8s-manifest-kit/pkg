package cache_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/k8s-manifest-kit/pkg/util/cache"

	. "github.com/onsi/gomega"
)

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
				return "spec2:" + strconv.Itoa(k.ID)
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
