package stream

// Limit returns a new Stream containing at most the first limit elements of the
// original Stream.
//
// Limit is an intermediate operation and is evaluated lazily. It stops
// consuming the upstream sequence as soon as limit elements have been yielded.
// A limit of zero yields nothing.
func (s *Stream[T]) Limit(limit int) *Stream[T] {
	return New(func(yield func(T) bool) {
		i := 0
		for v := range s.iterable.All() {
			if i >= limit {
				return
			}
			i++
			if !yield(v) {
				return
			}
		}
	})
}

// Skip returns a new Stream containing all elements of the original Stream
// except the first skip elements.
//
// Skip is an intermediate operation and is evaluated lazily. Skipping zero
// elements yields everything, and skipping at least as many elements as the
// Stream contains yields nothing.
func (s *Stream[T]) Skip(skip int) *Stream[T] {
	return New(func(yield func(T) bool) {
		i := 0
		for v := range s.iterable.All() {
			i++
			if i <= skip {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}
