package maps_test

import (
	"testing"

	"github.com/k8s-manifest-kit/pkg/util/maps"

	. "github.com/onsi/gomega"
)

const valueMutated = "mutated"

func TestDeepCloneMap(t *testing.T) {
	t.Run("should return nil for nil input", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(maps.DeepCloneMap(nil)).To(BeNil())
	})

	t.Run("should return empty map for empty input", func(t *testing.T) {
		g := NewWithT(t)

		result := maps.DeepCloneMap(map[string]any{})

		g.Expect(result).ToNot(BeNil())
		g.Expect(result).To(BeEmpty())
	})

	t.Run("should deeply copy nested maps", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{
			"top": "value",
			"nested": map[string]any{
				"level2": map[string]any{
					"level3": "deep",
				},
			},
		}

		clone := maps.DeepCloneMap(original)

		nestedClone := clone["nested"].(map[string]any)
		level2Clone := nestedClone["level2"].(map[string]any)
		level2Clone["level3"] = valueMutated

		nestedOrig := original["nested"].(map[string]any)
		level2Orig := nestedOrig["level2"].(map[string]any)
		g.Expect(level2Orig["level3"]).To(Equal("deep"))
	})

	t.Run("should deeply copy []any slices", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{
			"items": []any{
				map[string]any{"name": "a"},
				map[string]any{"name": "b"},
			},
		}

		clone := maps.DeepCloneMap(original)

		clonedItems := clone["items"].([]any)
		clonedItems[0].(map[string]any)["name"] = valueMutated

		originalItems := original["items"].([]any)
		g.Expect(originalItems[0].(map[string]any)["name"]).To(Equal("a"))
	})

	t.Run("should deeply copy typed slices", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{
			"strings": []string{"a", "b", "c"},
			"ints":    []int{1, 2, 3},
			"floats":  []float64{1.1, 2.2},
			"bools":   []bool{true, false},
		}

		clone := maps.DeepCloneMap(original)

		clone["strings"].([]string)[0] = valueMutated
		clone["ints"].([]int)[0] = 999
		clone["floats"].([]float64)[0] = 9.9
		clone["bools"].([]bool)[0] = false

		g.Expect(original["strings"].([]string)[0]).To(Equal("a"))
		g.Expect(original["ints"].([]int)[0]).To(Equal(1))
		g.Expect(original["floats"].([]float64)[0]).To(Equal(1.1))
		g.Expect(original["bools"].([]bool)[0]).To(BeTrue())
	})

	t.Run("should copy []int64 slices", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{
			"timestamps": []int64{100, 200, 300},
		}

		clone := maps.DeepCloneMap(original)

		clone["timestamps"].([]int64)[0] = 999

		g.Expect(original["timestamps"].([]int64)[0]).To(Equal(int64(100)))
	})

	t.Run("should copy uncommon slice types via reflection", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{
			"bytes":   []uint8{1, 2, 3},
			"float32": []float32{1.1, 2.2},
		}

		clone := maps.DeepCloneMap(original)

		clone["bytes"].([]uint8)[0] = 255
		clone["float32"].([]float32)[0] = 9.9

		g.Expect(original["bytes"].([]uint8)[0]).To(Equal(uint8(1)))
		g.Expect(original["float32"].([]float32)[0]).To(Equal(float32(1.1)))
	})

	t.Run("should return primitives as-is", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{
			"string":  "hello",
			"int":     42,
			"float":   3.14,
			"bool":    true,
			"nothing": nil,
		}

		clone := maps.DeepCloneMap(original)

		g.Expect(clone).To(Equal(original))
	})
}

func TestDeepCloneValue(t *testing.T) {
	t.Run("should return nil for nil input", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(maps.DeepCloneValue(nil)).To(BeNil())
	})

	t.Run("should return primitive values as-is", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(maps.DeepCloneValue("hello")).To(Equal("hello"))
		g.Expect(maps.DeepCloneValue(42)).To(Equal(42))
		g.Expect(maps.DeepCloneValue(3.14)).To(Equal(3.14))
		g.Expect(maps.DeepCloneValue(true)).To(BeTrue())
	})

	t.Run("should deep copy map[string]any", func(t *testing.T) {
		g := NewWithT(t)

		original := map[string]any{"key": "value"}
		clone := maps.DeepCloneValue(original)

		cloneMap := clone.(map[string]any)
		cloneMap["key"] = valueMutated

		g.Expect(original["key"]).To(Equal("value"))
	})

	t.Run("should deep copy []any", func(t *testing.T) {
		g := NewWithT(t)

		original := []any{"a", "b"}
		clone := maps.DeepCloneValue(original)

		cloneSlice := clone.([]any)
		cloneSlice[0] = valueMutated

		g.Expect(original[0]).To(Equal("a"))
	})
}
