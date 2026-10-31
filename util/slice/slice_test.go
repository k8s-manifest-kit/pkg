package slice_test

import (
	"testing"

	"github.com/k8s-manifest-kit/pkg/util/slice"

	. "github.com/onsi/gomega"
)

func TestToAnySlice(t *testing.T) {
	g := NewWithT(t)

	g.Expect(slice.ToAnySlice([]string{"first", "second"})).To(Equal([]any{"first", "second"}))
	g.Expect(slice.ToAnySlice([]int{1, 2})).To(Equal([]any{1, 2}))
}
