// Package array provides AGG-compatible array types and algorithms.
// This package implements POD (Plain Old Data) arrays that match the functionality
// of AGG's C++ array templates while being idiomatic Go code.
package array

// ArrayInterface defines the common interface for all array types in AGG.
// This interface allows algorithms to work with different array implementations.
type ArrayInterface[T any] interface {
	// Size returns the number of elements in the array
	Size() int

	// At returns the element at the specified index (bounds-checked)
	At(i int) T

	// Set sets the element at the specified index (bounds-checked)
	Set(i int, v T)

	// ValueAt returns the element at the specified index (may panic on out-of-bounds)
	ValueAt(i int) T
}

// MutableArrayInterface extends ArrayInterface with modification operations
type MutableArrayInterface[T any] interface {
	ArrayInterface[T]

	// Clear removes all elements but keeps capacity
	Clear()

	// RemoveAll removes all elements (equivalent to Clear)
	RemoveAll()
}

// GrowableArrayInterface extends MutableArrayInterface with growth operations
type GrowableArrayInterface[T any] interface {
	MutableArrayInterface[T]

	// Add appends an element to the end
	Add(v T)

	// PushBack appends an element to the end (equivalent to Add)
	PushBack(v T)

	// Capacity returns the current capacity
	Capacity() int
}

// SliceableArrayInterface provides access to underlying slice data
type SliceableArrayInterface[T any] interface {
	ArrayInterface[T]

	// Data returns the underlying slice
	Data() []T
}

// SerializableArrayInterface provides serialization capabilities
type SerializableArrayInterface[T any] interface {
	ArrayInterface[T]

	// ByteSize returns the size in bytes needed for serialization
	ByteSize() int

	// Serialize writes the array data to the provided byte slice
	Serialize(ptr []byte)

	// Deserialize reads array data from the provided byte slice
	Deserialize(data []byte)
}

// Ensure our main types will implement these interfaces
var (
	_ ArrayInterface[int]          = (*PodArrayAdaptor[int])(nil)
	_ ArrayInterface[int]          = (*PodAutoArray[int])(nil)
	_ MutableArrayInterface[int]   = (*PodAutoVector[int])(nil)
	_ SliceableArrayInterface[int] = (*PodArray[int])(nil)
	_ GrowableArrayInterface[int]  = (*PodVector[int])(nil)
	_ GrowableArrayInterface[int]  = (*PodBVector[int])(nil)
	_ ArrayInterface[int]          = (*RangeAdaptor[int])(nil)
)
