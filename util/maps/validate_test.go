package maps_test

import (
	"reflect"
	"testing"

	"github.com/k8s-manifest-kit/pkg/util/maps"

	. "github.com/onsi/gomega"
)

// definedMap and definedSlice are named types: nested values of these types
// do not match the exact dynamic types the clone and merge helpers
// traverse, so a cycle behind them stays a leaf. aliasMap is a true alias
// of map[string]any, keeps the ordinary dynamic type, and is traversed.
// rootValues exercises a defined map accepted at the root through the
// map[string]any parameter.
type (
	definedMap   map[string]any
	definedSlice []any
)

type (
	aliasMap   = map[string]any
	rootValues = map[string]any
)

func TestValidateAcyclic(t *testing.T) {
	t.Run("should pass nil, empty, and ordinary nested values", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(maps.ValidateAcyclic(nil)).To(Succeed())
		g.Expect(maps.ValidateAcyclic(map[string]any{})).To(Succeed())

		ordinary := map[string]any{
			"scalar": "value",
			"nested": map[string]any{
				"items": []any{
					map[string]any{"name": "a"},
					[]any{"b", true, 1.5},
				},
				"typed": []string{"x", "y"},
			},
		}

		g.Expect(maps.ValidateAcyclic(ordinary)).To(Succeed())
	})

	t.Run("should pass shared child references that do not form cycles", func(t *testing.T) {
		g := NewWithT(t)

		shared := map[string]any{"tier": "web"}

		items := make([]any, 2)
		items[0] = shared
		items[1] = items[:1]

		values := map[string]any{
			"alpha": shared,
			"beta":  shared,
			"items": items,
		}

		g.Expect(maps.ValidateAcyclic(values)).To(Succeed())

		// The input is unchanged by validation.
		g.Expect(shared["tier"]).To(Equal("web"))
		g.Expect(values["alpha"]).To(Equal(shared))
	})

	t.Run("should reject a self-referential map value", func(t *testing.T) {
		g := NewWithT(t)

		cyclic := map[string]any{}
		cyclic["self"] = cyclic

		err := maps.ValidateAcyclic(map[string]any{"cyclic": cyclic})

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(maps.ErrCycle))
		g.Expect(err.Error()).To(ContainSubstring(`"cyclic.self" refers back to "cyclic"`))
	})

	t.Run("should reject a self-referential slice value", func(t *testing.T) {
		g := NewWithT(t)

		cyclic := make([]any, 1)
		cyclic[0] = cyclic

		err := maps.ValidateAcyclic(map[string]any{"cyclic": cyclic})

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(maps.ErrCycle))
		g.Expect(err.Error()).To(ContainSubstring(`"cyclic[0]" refers back to "cyclic"`))
	})

	t.Run("should reject indirect map and slice cycles", func(t *testing.T) {
		g := NewWithT(t)

		inner := map[string]any{}
		outer := map[string]any{"inner": inner}
		inner["outer"] = outer

		mapErr := maps.ValidateAcyclic(map[string]any{"root": outer})
		g.Expect(mapErr).To(HaveOccurred())
		g.Expect(mapErr).To(MatchError(maps.ErrCycle))
		g.Expect(mapErr.Error()).To(ContainSubstring(`"root.inner.outer" refers back to "root"`))

		slice := make([]any, 1)
		holder := map[string]any{"items": slice}
		slice[0] = holder

		sliceErr := maps.ValidateAcyclic(map[string]any{"data": holder})
		g.Expect(sliceErr).To(HaveOccurred())
		g.Expect(sliceErr).To(MatchError(maps.ErrCycle))
		g.Expect(sliceErr.Error()).To(ContainSubstring(`"data.items[0]" refers back to "data"`))
	})

	t.Run("should report the first cyclic key deterministically", func(t *testing.T) {
		g := NewWithT(t)

		alpha := map[string]any{}
		alpha["self"] = alpha

		beta := map[string]any{}
		beta["self"] = beta

		err := maps.ValidateAcyclic(map[string]any{
			"beta":  beta,
			"alpha": alpha,
		})

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(`"alpha.self" refers back to "alpha"`))
		g.Expect(err.Error()).ToNot(ContainSubstring("beta"))
	})

	t.Run("should traverse a defined root map and true aliases", func(t *testing.T) {
		g := NewWithT(t)

		cyclic := map[string]any{}
		cyclic["self"] = cyclic

		rootErr := maps.ValidateAcyclic(rootValues{"cyclic": cyclic})
		g.Expect(rootErr).To(HaveOccurred())
		g.Expect(rootErr).To(MatchError(maps.ErrCycle))
		g.Expect(rootErr.Error()).To(ContainSubstring(`"cyclic.self" refers back to "cyclic"`))

		alias := aliasMap{}
		alias["self"] = alias

		aliasErr := maps.ValidateAcyclic(map[string]any{"alias": alias})
		g.Expect(aliasErr).To(HaveOccurred())
		g.Expect(aliasErr).To(MatchError(maps.ErrCycle))
		g.Expect(aliasErr.Error()).To(ContainSubstring(`"alias.self" refers back to "alias"`))
	})

	t.Run("should treat nested defined map and slice types as leaves", func(t *testing.T) {
		g := NewWithT(t)

		cyclicMap := definedMap{}
		cyclicMap["self"] = cyclicMap

		cyclicSlice := definedSlice{nil}
		cyclicSlice[0] = cyclicSlice

		values := map[string]any{
			"leafMap":   cyclicMap,
			"leafSlice": cyclicSlice,
		}

		// Defined types do not match the traversed dynamic types, so the
		// cycles behind them are not entered; the retained values are kept
		// as-is and never executed downstream.
		g.Expect(maps.ValidateAcyclic(values)).To(Succeed())
	})

	t.Run("should leave rejected input unchanged", func(t *testing.T) {
		g := NewWithT(t)

		cyclic := map[string]any{}
		cyclic["self"] = cyclic

		values := map[string]any{
			"cyclic": cyclic,
			"scalar": 1,
		}

		err := maps.ValidateAcyclic(values)

		g.Expect(err).To(HaveOccurred())
		g.Expect(values).To(HaveLen(2))
		g.Expect(values["scalar"]).To(Equal(1))

		// The cycle reference is still the identical map: identity is
		// checked by header pointer instead of recursing into the cycle.
		g.Expect(reflect.ValueOf(values["cyclic"]).Pointer()).To(Equal(reflect.ValueOf(cyclic).Pointer()))

		inner, ok := values["cyclic"].(map[string]any)
		g.Expect(ok).To(BeTrue())
		g.Expect(reflect.ValueOf(inner["self"]).Pointer()).To(Equal(reflect.ValueOf(cyclic).Pointer()))
	})
}
