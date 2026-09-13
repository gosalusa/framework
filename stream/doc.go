// Package stream provides an immutable, lazy sequence abstraction that makes it
// easy to build and compose pipelines over slices and iterators.
//
// A Stream wraps Go's iter.Seq[T] protocol and provides a chainable API of
// intermediate operations such as Filter, Map, FlatMap, Limit, Skip, and Sort
// that return new Streams, and terminal operations such as Slice, All, Find,
// Reduce, and ReduceFrom that execute the pipeline and produce a result.
//
// Streams are lazy: intermediate operations only describe how to transform the
// data and perform no work until a terminal operation consumes them. The
// pipeline is also short-circuiting, so operations like Limit and Find stop
// iterating the upstream sequence as soon as enough elements have been seen.
//
// Create a Stream from a slice with Of, or from any iterator with New:
//
//	s := stream.Of([]int{1, 2, 3, 4, 5}).
//		Filter(func(i int) bool { return i > 2 }).
//		Map(func(i int) string { return fmt.Sprintf("item-%d", i) }).
//		Slice()
//
// See the Stream type documentation for the full list of operations.
package stream