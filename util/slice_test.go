package util_test

import (
	"testing"

	"github.com/k8s-manifest-kit/pkg/util"

	. "github.com/onsi/gomega"
)

func TestToAnySlice(t *testing.T) {
	g := NewWithT(t)

	g.Expect(util.ToAnySlice([]string{"first", "second"})).To(Equal([]any{"first", "second"}))
	g.Expect(util.ToAnySlice([]int{1, 2})).To(Equal([]any{1, 2}))
}
