package stream

// Reduce combines the elements of the Stream into a single value of type R by
// applying fn to an accumulator and each element in order.
//
// Reduce is a terminal operation that starts with the zero value of R as the
// initial accumulator, so it returns the zero value for an empty Stream. Use
// ReduceFrom to start from a different accumulator.
func (s *Stream[T]) Reduce[R any](fn func(R, T) R) R {
	var accumulator R
	return s.ReduceFrom(accumulator, fn)
}

// ReduceFrom combines the elements of the Stream into a single value by
// applying fn to the accumulator and each element in order, starting from the
// given initial accumulator.
//
// ReduceFrom is a terminal operation. For an empty Stream it returns the
// supplied accumulator unchanged.
func (s *Stream[T]) ReduceFrom[R any](accumulator R, fn func(R, T) R) R {
	for a := range s.iterable.All() {
		accumulator = fn(accumulator, a)
	}
	return accumulator
}
