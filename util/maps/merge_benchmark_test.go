package maps_test

import (
	"fmt"
	"testing"

	"github.com/k8s-manifest-kit/pkg/util/maps"
)

// Benchmarks

func BenchmarkDeepMerge_SmallMaps(b *testing.B) {
	base := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": map[string]any{
			"nested1": "value",
		},
	}
	overlay := map[string]any{
		"key2": "override",
		"key4": "new",
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_LargeMaps(b *testing.B) {
	base := make(map[string]any, 100)
	for i := range 100 {
		base[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	overlay := make(map[string]any, 50)
	for i := range 50 {
		overlay[fmt.Sprintf("key%d", i*2)] = fmt.Sprintf("override%d", i)
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_DeepNesting(b *testing.B) {
	base := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"level5": "value",
					},
				},
			},
		},
	}
	overlay := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"level5": "override",
						"new":    "value",
					},
				},
			},
		},
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_WithSlices(b *testing.B) {
	base := map[string]any{
		"anySlice":    []any{"a", "b", "c"},
		"stringSlice": []string{"x", "y", "z"},
		"intSlice":    []int{1, 2, 3, 4, 5},
		"config": map[string]any{
			"nested": []any{
				map[string]any{"key": "value"},
			},
		},
	}
	overlay := map[string]any{
		"anySlice":    []any{"d", "e"},
		"stringSlice": []string{"p", "q"},
		"intSlice":    []int{6, 7, 8},
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_ComplexNested(b *testing.B) {
	base := map[string]any{
		"replicaCount": 1,
		"image": map[string]any{
			"repository": "nginx",
			"tag":        "v1.0",
			"pullPolicy": "IfNotPresent",
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 80,
			"annotations": map[string]any{
				"key1": "value1",
			},
		},
		"resources": map[string]any{
			"limits": map[string]any{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
	}
	overlay := map[string]any{
		"replicaCount": 3,
		"image": map[string]any{
			"tag": "v2.0",
		},
		"service": map[string]any{
			"type": "LoadBalancer",
			"annotations": map[string]any{
				"key2": "value2",
			},
		},
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_TypedSlices(b *testing.B) {
	b.Run("string", func(b *testing.B) {
		base := map[string]any{
			"tags": []string{"dev", "test", "staging"},
		}
		overlay := map[string]any{
			"tags": []string{"prod"},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(base, overlay)
		}
	})

	b.Run("int", func(b *testing.B) {
		base := map[string]any{
			"ports": []int{8080, 9090, 3000},
		}
		overlay := map[string]any{
			"ports": []int{443},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(base, overlay)
		}
	})

	b.Run("int64", func(b *testing.B) {
		base := map[string]any{
			"timestamps": []int64{1234567890, 9876543210},
		}
		overlay := map[string]any{
			"timestamps": []int64{1111111111},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(base, overlay)
		}
	})

	b.Run("float64", func(b *testing.B) {
		base := map[string]any{
			"metrics": []float64{1.1, 2.2, 3.3},
		}
		overlay := map[string]any{
			"metrics": []float64{4.4},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(base, overlay)
		}
	})

	b.Run("bool", func(b *testing.B) {
		base := map[string]any{
			"flags": []bool{true, false, true},
		}
		overlay := map[string]any{
			"flags": []bool{false},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(base, overlay)
		}
	})
}

func BenchmarkDeepMerge_NilHandling(b *testing.B) {
	b.Run("bothNil", func(b *testing.B) {
		for range b.N {
			_ = maps.DeepMerge(nil, nil)
		}
	})

	b.Run("baseNil", func(b *testing.B) {
		overlay := map[string]any{
			"key1": "value1",
			"key2": map[string]any{"nested": "value"},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(nil, overlay)
		}
	})

	b.Run("overlayNil", func(b *testing.B) {
		base := map[string]any{
			"key1": "value1",
			"key2": map[string]any{"nested": "value"},
		}
		b.ResetTimer()
		for range b.N {
			_ = maps.DeepMerge(base, nil)
		}
	})
}

func BenchmarkDeepMerge_TypeMismatch(b *testing.B) {
	base := map[string]any{
		"field1": map[string]any{
			"nested": "value",
		},
		"field2": "string",
		"field3": []any{"a", "b"},
	}
	overlay := map[string]any{
		"field1": "now_a_string",
		"field2": map[string]any{"now": "map"},
		"field3": 123,
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_NoOverlap(b *testing.B) {
	base := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	overlay := map[string]any{
		"key4": "value4",
		"key5": "value5",
		"key6": "value6",
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}

func BenchmarkDeepMerge_FullOverlap(b *testing.B) {
	base := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	overlay := map[string]any{
		"key1": "override1",
		"key2": "override2",
		"key3": "override3",
	}

	for b.Loop() {
		_ = maps.DeepMerge(base, overlay)
	}
}
