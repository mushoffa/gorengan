package generic

// Slice represents a generic type of slice.
type Slice[T any] []T

// NewSlice returns a new [Slice] with initial given size.
func NewSlice[T any](size int) Slice[T] {
	return make(Slice[T], size)
}

// Size returns the current size of [Slice].
func (n Slice[T]) Size() int {
	return len(n)
}

// Add inserts a new data element to the last index of [Slice].
func (n *Slice[T]) Add(element T) {
	*n = append(*n, element)
}

// Append combines another slice variable to the current [Slice] with same data type.
func (n *Slice[T]) Append(elements []T) {
	*n = append(*n, elements...)
}
