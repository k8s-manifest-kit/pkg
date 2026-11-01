package maps

import (
	"errors"
	"fmt"
	stdmaps "maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var (
	// ErrCycle is returned when a value tree contains a reference cycle
	// among the nested maps and slices the clone and merge helpers
	// traverse. Those helpers recurse without cycle detection, so callers
	// use ValidateAcyclic to reject such trees before copying them.
	ErrCycle = errors.New("cyclic values")
)

// rootPath names the top-level values map in reported key paths.
const rootPath = "<root>"

// ValidateAcyclic returns an error wrapping ErrCycle when values contain a
// reference cycle among the nested maps and slices that DeepCloneMap,
// DeepCloneValue, and DeepMerge recurse into. It follows the same
// reachability as those helpers — exact map[string]any and []any nodes —
// and treats every other type as a leaf, so valid values of any shape pass
// unchanged. The input is not modified.
//
// Only nodes on one descent path are compared, so shared references that do
// not form a cycle remain valid, including a slice that holds a prefix of
// itself. A defined map type passed as the root is traversed through the
// map parameter, while nested values of defined map or slice types do not
// match the traversed dynamic types and stay leaves; true aliases keep the
// ordinary dynamic types and are traversed. Keys are visited in sorted
// order, making the reported cycle path deterministic.
func ValidateAcyclic(values map[string]any) error {
	if len(values) == 0 {
		return nil
	}

	checker := cycleChecker{}

	return checker.checkMap(values)
}

// cycleNode identifies a map or slice during cycle detection: maps by their
// header pointer, slices by their data pointer and length. Two slices
// sharing a backing array but differing in length stay distinct, which
// matches the slice values the clone and merge helpers recurse into.
type cycleNode struct {
	pointer uintptr
	length  int
}

// cyclePathSegment is one segment of a values key path: a map key when
// isKey is set, a slice element index otherwise.
type cyclePathSegment struct {
	key   string
	index int
	isKey bool
}

// cycleChecker detects reference cycles in JSON-like value trees.
type cycleChecker struct {
	// ancestors holds the identity of every map and slice on the current
	// descent path, oldest first. The clone and merge helpers recurse
	// forever exactly when some node is reachable from itself, which makes
	// it appear on this path a second time.
	ancestors []cycleNode

	// segments holds the key path of the current descent path, oldest
	// first. It is only rendered into a string when a cycle is reported.
	segments []cyclePathSegment
}

// checkMap reports an error when a reference cycle is reachable through the
// given map. Keys are visited in sorted order so the reported cycle does
// not depend on map iteration order.
func (c *cycleChecker) checkMap(m map[string]any) error {
	node := cycleNode{pointer: reflect.ValueOf(m).Pointer(), length: len(m)}
	if slices.Contains(c.ancestors, node) {
		return c.cycleError(node)
	}

	c.ancestors = append(c.ancestors, node)

	for _, key := range slices.Sorted(stdmaps.Keys(m)) {
		c.segments = append(c.segments, cyclePathSegment{key: key, isKey: true})

		if err := c.checkValue(m[key]); err != nil {
			return err
		}

		c.segments = c.segments[:len(c.segments)-1]
	}

	c.ancestors = c.ancestors[:len(c.ancestors)-1]

	return nil
}

// checkSlice reports an error when a reference cycle is reachable through
// the given slice.
func (c *cycleChecker) checkSlice(s []any) error {
	node := cycleNode{pointer: reflect.ValueOf(s).Pointer(), length: len(s)}
	if slices.Contains(c.ancestors, node) {
		return c.cycleError(node)
	}

	c.ancestors = append(c.ancestors, node)

	for i, element := range s {
		c.segments = append(c.segments, cyclePathSegment{index: i})

		if err := c.checkValue(element); err != nil {
			return err
		}

		c.segments = c.segments[:len(c.segments)-1]
	}

	c.ancestors = c.ancestors[:len(c.ancestors)-1]

	return nil
}

// checkValue reports an error when a reference cycle is reachable through
// the given value. Only map[string]any and []any are entered, matching the
// reachability of the shared clone and merge helpers; every other type is a
// leaf.
func (c *cycleChecker) checkValue(value any) error {
	switch v := value.(type) {
	case map[string]any:
		return c.checkMap(v)
	case []any:
		return c.checkSlice(v)
	default:
		return nil
	}
}

// cycleError describes the cycle closed by the given ancestor node: where
// the closing reference lives and where it points back to.
func (c *cycleChecker) cycleError(node cycleNode) error {
	ancestor := slices.Index(c.ancestors, node)

	return fmt.Errorf(
		"%w: %q refers back to %q",
		ErrCycle,
		c.path(len(c.segments)),
		c.path(ancestor),
	)
}

// path renders the first depth segments of the current descent path as a
// dotted key path with bracketed slice indexes. It runs only when a cycle
// is reported.
func (c *cycleChecker) path(depth int) string {
	if depth == 0 {
		return rootPath
	}

	parts := make([]string, 0, depth)

	for i, segment := range c.segments[:depth] {
		switch {
		case !segment.isKey:
			parts = append(parts, "["+strconv.Itoa(segment.index)+"]")
		case i > 0:
			parts = append(parts, "."+segment.key)
		default:
			parts = append(parts, segment.key)
		}
	}

	return strings.Join(parts, "")
}
