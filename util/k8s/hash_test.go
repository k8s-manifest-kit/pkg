package k8s_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/k8s-manifest-kit/pkg/util/k8s"

	. "github.com/onsi/gomega"
)

func TestContentHash(t *testing.T) {
	t.Run("produces deterministic hash for the same object", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":      "test-config",
					"namespace": "default",
				},
				"data": map[string]any{
					"key": "value",
				},
			},
		}

		hash1 := k8s.ContentHash(obj)
		hash2 := k8s.ContentHash(obj)

		g.Expect(hash1).ShouldNot(BeEmpty())
		g.Expect(hash1).Should(Equal(hash2))
	})

	t.Run("different objects produce different hashes", func(t *testing.T) {
		g := NewWithT(t)

		obj1 := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "config-a",
				},
				"data": map[string]any{
					"key": "value-a",
				},
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "config-b",
				},
				"data": map[string]any{
					"key": "value-b",
				},
			},
		}

		hash1 := k8s.ContentHash(obj1)
		hash2 := k8s.ContentHash(obj2)

		g.Expect(hash1).ShouldNot(Equal(hash2))
	})

	t.Run("objects with different annotations produce different hashes", func(t *testing.T) {
		g := NewWithT(t)

		obj1 := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "test",
					"annotations": map[string]any{
						"source": "renderer-a",
					},
				},
			},
		}

		obj2 := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "test",
					"annotations": map[string]any{
						"source": "renderer-b",
					},
				},
			},
		}

		hash1 := k8s.ContentHash(obj1)
		hash2 := k8s.ContentHash(obj2)

		g.Expect(hash1).ShouldNot(Equal(hash2))
	})

	t.Run("returns sha256-prefixed hex digest", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "test",
				},
			},
		}

		hash := k8s.ContentHash(obj)

		// "sha256:" (7 chars) + 64 hex chars = 71
		g.Expect(hash).Should(HaveLen(71))
		g.Expect(hash).Should(MatchRegexp("^sha256:[0-9a-f]{64}$"))
	})
}
