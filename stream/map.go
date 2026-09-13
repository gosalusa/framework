package stream

// Map returns a new Stream containing the results of applying the function fn
// to each element in the original Stream.
//
// The mapping function is evaluated lazily as the returned Stream is iterated.
// If the target sequence iteration is halted early by the consumer, mapping
// stops immediately.
func (s *Stream[T]) Map[R any](fn func(T) R) *Stream[R] {
	return New(func(yield func(R) bool) {
		for v := range s.iterable.All() {
			if !yield(fn(v)) {
				return
			}
		}
	})
}

// FlatMap returns a new Stream containing the elements of each slice produced
// by applying fn to every element of the original Stream.
//
// FlatMap is an intermediate operation and is evaluated lazily. The results are
// concatenated in order, and a function call that returns an empty slice
// contributes no elements.
func (s *Stream[T]) FlatMap[R any](fn func(T) []R) *Stream[R] {
	return New(func(yield func(R) bool) {
		for a := range s.iterable.All() {
			for _, v := range fn(a) {
				if !yield(v) {
					return
				}
			}
		}
	})
}
