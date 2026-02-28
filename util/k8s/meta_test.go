package k8s_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/k8s-manifest-kit/pkg/util/k8s"

	. "github.com/onsi/gomega"
)

func TestSetAnnotation(t *testing.T) {
	t.Run("sets annotation on object without existing annotations", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "test"},
			},
		}

		k8s.SetAnnotation(obj, "example.io/key", "value")

		g.Expect(obj.GetAnnotations()).Should(HaveKeyWithValue("example.io/key", "value"))
	})

	t.Run("preserves existing annotations", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":        "test",
					"annotations": map[string]any{"existing": "val"},
				},
			},
		}

		k8s.SetAnnotation(obj, "new-key", "new-val")

		annotations := obj.GetAnnotations()
		g.Expect(annotations).Should(HaveKeyWithValue("existing", "val"))
		g.Expect(annotations).Should(HaveKeyWithValue("new-key", "new-val"))
	})

	t.Run("overwrites existing key", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":        "test",
					"annotations": map[string]any{"key": "old"},
				},
			},
		}

		k8s.SetAnnotation(obj, "key", "new")

		g.Expect(obj.GetAnnotations()).Should(HaveKeyWithValue("key", "new"))
	})
}

func TestSetAnnotations(t *testing.T) {
	t.Run("merges multiple annotations", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":        "test",
					"annotations": map[string]any{"existing": "val"},
				},
			},
		}

		k8s.SetAnnotations(obj, map[string]string{
			"a": "1",
			"b": "2",
		})

		annotations := obj.GetAnnotations()
		g.Expect(annotations).Should(HaveKeyWithValue("existing", "val"))
		g.Expect(annotations).Should(HaveKeyWithValue("a", "1"))
		g.Expect(annotations).Should(HaveKeyWithValue("b", "2"))
	})

	t.Run("creates annotations map when none exists", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "test"},
			},
		}

		k8s.SetAnnotations(obj, map[string]string{"key": "val"})

		g.Expect(obj.GetAnnotations()).Should(HaveKeyWithValue("key", "val"))
	})
}

func TestSetLabel(t *testing.T) {
	t.Run("sets label on object without existing labels", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "test"},
			},
		}

		k8s.SetLabel(obj, "app", "myapp")

		g.Expect(obj.GetLabels()).Should(HaveKeyWithValue("app", "myapp"))
	})

	t.Run("preserves existing labels", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":   "test",
					"labels": map[string]any{"existing": "val"},
				},
			},
		}

		k8s.SetLabel(obj, "new-key", "new-val")

		labels := obj.GetLabels()
		g.Expect(labels).Should(HaveKeyWithValue("existing", "val"))
		g.Expect(labels).Should(HaveKeyWithValue("new-key", "new-val"))
	})

	t.Run("overwrites existing key", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":   "test",
					"labels": map[string]any{"key": "old"},
				},
			},
		}

		k8s.SetLabel(obj, "key", "new")

		g.Expect(obj.GetLabels()).Should(HaveKeyWithValue("key", "new"))
	})
}

func TestSetLabels(t *testing.T) {
	t.Run("merges multiple labels", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":   "test",
					"labels": map[string]any{"existing": "val"},
				},
			},
		}

		k8s.SetLabels(obj, map[string]string{
			"a": "1",
			"b": "2",
		})

		labels := obj.GetLabels()
		g.Expect(labels).Should(HaveKeyWithValue("existing", "val"))
		g.Expect(labels).Should(HaveKeyWithValue("a", "1"))
		g.Expect(labels).Should(HaveKeyWithValue("b", "2"))
	})

	t.Run("creates labels map when none exists", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "test"},
			},
		}

		k8s.SetLabels(obj, map[string]string{"key": "val"})

		g.Expect(obj.GetLabels()).Should(HaveKeyWithValue("key", "val"))
	})
}
