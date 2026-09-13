package stream

// Filter returns a new Stream containing only the elements for which fn
// returns true.
//
// Filter is an intermediate operation and is evaluated lazily as the returned
// Stream is iterated. It only consumes upstream elements until the consumer
// stops iterating.
func (s *Stream[T]) Filter(fn func(T) bool) *Stream[T] {
	return New(func(yield func(T) bool) {
		for v := range s.iterable.All() {
			if !fn(v) {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
}

// Find returns the first element for which fn returns true and whether such an
// element was found.
//
// Find is a terminal operation that executes the pipeline and stops as soon as
// the first matching element is seen. If no element matches, it returns the
// zero value of T and false.
func (s *Stream[T]) Find(fn func(T) bool) (T, bool) {
	for v := range s.iterable.All() {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}
