package stream

import (
	"fmt"
	"iter"
	"testing"
	"unsafe"

	"github.com/go-openapi/testify/v2/assert"
)

func TestOf(t *testing.T) {
	var seq iter.Seq[int] = func(yield func(int) bool) {
		yield(10)
		yield(20)
		yield(30)
	}
	s := New(seq).Slice()
	assert.Equal(t, []int{10, 20, 30}, s)
}

func TestOfSlice(t *testing.T) {
	s := Of([]int{1, 2, 3}).Slice()
	assert.Equal(t, []int{1, 2, 3}, s)
}

func TestOfSlice_Empty(t *testing.T) {
	s := Of([]int{}).Slice()
	assert.Empty(t, s)
}

func TestStream_All(t *testing.T) {
	var collected []int
	for v := range Of([]int{1, 2, 3}).All() {
		collected = append(collected, v)
	}
	assert.Equal(t, []int{1, 2, 3}, collected)
}

func TestStream_Slice(t *testing.T) {
	s := Of([]int{4, 5, 6}).Slice()
	assert.Equal(t, []int{4, 5, 6}, s)
}

func TestStream_Chain(t *testing.T) {
	s := Of([]int{1, 2, 3, 4, 5, 6}).
		Filter(func(i int) bool {
			return i%2 == 0
		}).
		Limit(2).
		Slice()
	assert.Equal(t, []int{2, 4}, s)
}

func TestStream_Chain_FilterSkipMap(t *testing.T) {
	s := Of([]int{1, 2, 3, 4, 5}).
		Filter(func(i int) bool {
			return i > 2
		}).
		Skip(1).
		Map(func(i int) string {
			return "v"
		}).
		Slice()
	assert.Equal(t, []string{"v", "v"}, s)
}

func TestStream_Chain_FilterMapCollect(t *testing.T) {
	s := Of([]int{1, 2, 3, 4, 5}).
		Filter(func(i int) bool {
			return i >= 3
		}).
		Map(func(i int) string {
			return fmt.Sprintf("item-%d", i)
		}).
		Slice()
	assert.Equal(t, []string{"item-3", "item-4", "item-5"}, s)
}

func TestStream_Chain_FlatMapFilterLimit(t *testing.T) {
	s := Of([]int{1, 2, 3}).
		FlatMap(func(i int) []int {
			return []int{i, i * 10}
		}).
		Filter(func(i int) bool {
			return i > 5
		}).
		Limit(2).
		Slice()
	assert.Equal(t, []int{10, 20}, s)
}

func TestStream_Of_SliceSameInstance(t *testing.T) {
	source := []int{1, 2, 3}
	s := Of(source).Slice()

	assert.True(t, unsafe.SliceData(source) == unsafe.SliceData(s), "source and result slices do not point to the same object")
}

func ExampleStream() {
	s := Of([]int{1, 2, 3, 4, 5}).
		Filter(func(i int) bool {
			return i > 2
		}).
		Skip(1).
		Map(func(i int) string {
			return fmt.Sprintf("item-%d", i)
		}).
		Slice()

	fmt.Println(s)
	// Output: [item-4 item-5]
}
