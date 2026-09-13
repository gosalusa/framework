package stream

import (
	"iter"
	"slices"
)

// Stream is a lazy sequence of elements of type T.
//
// Streams wrap Go's iter.Seq[T] protocol and expose a chainable API. Use Of to
// create a Stream from a slice or New to create one from an arbitrary iterator.
// Intermediate operations (Filter, Map, FlatMap, Limit, Skip, Sort) return new
// Streams without doing any work, while terminal operations (Slice, All, Find,
// Reduce, ReduceFrom) execute the pipeline and produce a result.
type Stream[T any] struct {
	iterable iterable[T]
}

type iterable[T any] interface {
	All() iter.Seq[T]
}

type iterableSeq[T any] iter.Seq[T]

func (i iterableSeq[T]) All() iter.Seq[T] {
	return iter.Seq[T](i)
}

type iterableSlice[T any] struct {
	slice []T
}

func (i iterableSlice[T]) All() iter.Seq[T] {
	return slices.Values(i.slice)
}

// New returns a Stream that yields the elements of seq.
//
// Use New to turn any iter.Seq[T] iterator, such as a generator function, into
// a Stream. Prefer Of when building a Stream from a slice.
func New[T any](seq iter.Seq[T]) *Stream[T] {
	return &Stream[T]{
		iterable: iterableSeq[T](seq),
	}
}

// Of returns a Stream that yields the elements of s.
//
// The Stream holds a reference to s rather than a copy, so Slice on the
// resulting Stream returns the same underlying array and any mutations are
// shared with the caller's slice.
func Of[T any](s []T) *Stream[T] {
	return &Stream[T]{
		iterable: iterableSlice[T]{slice: s},
	}
}

// All returns the underlying sequence as an iter.Seq[T] iterator.
//
// All is a terminal operation; iterating the returned sequence runs the
// pipeline. It can be used with a for-range loop when a Stream is not needed.
func (s *Stream[T]) All() iter.Seq[T] {
	return s.iterable.All()
}

// Slice returns the elements of the Stream as a slice.
//
// Slice executes the pipeline and collects its output. It is a terminal
// operation. If the Stream was created with Of and has not been transformed,
// the original slice is returned without copying.
func (s *Stream[T]) Slice() []T {
	if i, ok := s.iterable.(iterableSlice[T]); ok {
		return i.slice
	}
	return slices.Collect(s.iterable.All())
}
