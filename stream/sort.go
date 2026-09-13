package stream

import (
	"slices"
)

// Sort returns a new Stream containing the elements of the original Stream
// sorted according to the three-way comparison function cmp. cmp should return
// a negative value when a is less than b, zero when they are equal, and a
// positive value when a is greater than b.
//
// Unlike the other intermediate operations, Sort is eager: it consumes the
// original Stream and sorts the collected elements at the time it is called,
// rather than when the returned Stream is iterated.
func (s *Stream[T]) Sort(cmp func(a, b T) int) *Stream[T] {
	slice := s.Slice()
	slices.SortFunc(slice, cmp)
	return Of(slice)
}
