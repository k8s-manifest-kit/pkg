package option_test

import (
	"testing"

	"github.com/k8s-manifest-kit/pkg/util/option"

	. "github.com/onsi/gomega"
)

func TestFunctionalOptionApplyTo(t *testing.T) {
	g := NewWithT(t)

	type config struct {
		value string
	}

	var opt option.Option[config] = option.FunctionalOption[config](func(target *config) {
		target.value = "applied"
	})

	configValue := config{value: "initial"}
	opt.ApplyTo(&configValue)

	g.Expect(configValue.value).To(Equal("applied"))
}

func TestFunctionalOptionSatisfiesInterface(t *testing.T) {
	g := NewWithT(t)

	type config struct {
		value int
	}

	first := option.FunctionalOption[config](func(target *config) {
		target.value++
	})
	second := option.FunctionalOption[config](func(target *config) {
		target.value += 2
	})

	configValue := config{}
	for _, opt := range []option.Option[config]{first, second} {
		opt.ApplyTo(&configValue)
	}

	g.Expect(configValue.value).To(Equal(3))
}
