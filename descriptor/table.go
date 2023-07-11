package descriptor

import "math/bits"

// Table is a data structure mapping 32 bit descriptor to objects.
//
// The data structure optimizes for memory density and lookup performance,
// trading off compute at insertion time. This is a useful compromise for the
// use cases we employ it with: objects are usually accessed a lot more
// often than they are inserted, each operation requires a table lookup so we are
// better off spending extra compute to insert objects in the table in order to
// get cheaper lookups. Memory efficiency is also crucial to support scaling
// with programs that open thousands of objects: having a high or non-linear
// memory-to-item ratio could otherwise be used as an attack vector by malicous
// applications attempting to damage performance of the host.
type Table[Descriptor ~int32 | ~uint32, Object any] struct {
	masks []uint64
	table []Object
}

// Len returns the number of objects stored in the table.
func (t *Table[Descriptor, Object]) Len() (n int) {
	// We could make this a O(1) operation if we cached the number of objects in
	// the table. More state usually means more problems, so until we have a
	// clear need for this, the simple implementation may be a better trade off.
	for _, mask := range t.masks {
		n += bits.OnesCount64(mask)
	}
	return n
}

// Grow ensures that t has enough room for n objects, potentially reallocating the
// internal buffers if their capacity was too small to hold this many objects.
func (t *Table[Descriptor, Object]) Grow(n int) {
	// Round up to a multiple of 64 since this is the smallest increment due to
	// using 64 bits masks.
	n = (n*64 + 63) / 64

	if n > len(t.masks) {
		masks := make([]uint64, n)
		copy(masks, t.masks)

		table := make([]Object, n*64)
		copy(table, t.table)

		t.masks = masks
		t.table = table
	}
}

// Insert inserts the given object to the table, returning the descriptor that
// it is mapped to.
//
// The method does not perform deduplication, it is possible for the same object
// to be inserted multiple times, each insertion will return a different
// descriptor.
func (t *Table[Descriptor, Object]) Insert(object Object) (desc Descriptor) {
	offset := 0
	for {
		// Note: this loop could be made a lot more efficient using vectorized
		// operations: 256 bits vector registers would yield a theoretical 4x
		// speed up (e.g. using AVX2).
		for index, mask := range t.masks[offset:] {
			if ^mask != 0 { // not full?
				shift := bits.TrailingZeros64(^mask)
				index += offset
				desc = Descriptor(index)*64 + Descriptor(shift)
				t.table[desc] = object
				t.masks[index] = mask | uint64(1<<shift)
				return desc
			}
		}

		offset = len(t.masks)
		n := 2 * len(t.masks)
		if n == 0 {
			n = 1
		}

		t.Grow(n)
	}
}

// Assign is similar to Insert but it inserts the object at a specific
// descriptor number. If another object was already associated with that
// number, it is returned and the boolean is set to true to indicate that
// a object was replaced.
func (t *Table[Descriptor, Object]) Assign(desc Descriptor, object Object) (prev Object, replaced bool) {
	if int(desc) >= len(t.table) {
		t.Grow(int(desc) + 1)
	}
	index := uint(desc) / 64
	shift := uint(desc) % 64
	if (t.masks[index] & (1 << shift)) != 0 {
		prev, replaced = t.table[desc], true
	}
	t.masks[index] |= 1 << shift
	t.table[desc] = object
	return
}

// Access returns a pointer to the object associated with the given
// descriptor, which may be nil if it was not found in the table.
func (t *Table[Descriptor, Object]) Access(desc Descriptor) *Object {
	if i := int(desc); i >= 0 && i < len(t.table) {
		index := uint(desc) / 64
		shift := uint(desc) % 64
		if (t.masks[index] & (1 << shift)) != 0 {
			return &t.table[i]
		}
	}
	return nil
}

// Lookup returns the object associated with the given descriptor.
func (t *Table[Descriptor, Object]) Lookup(desc Descriptor) (object Object, found bool) {
	ptr := t.Access(desc)
	if ptr != nil {
		object, found = *ptr, true
	}
	return
}

// Delete deletes the object stored at the given descriptor from the table.
func (t *Table[Descriptor, Object]) Delete(desc Descriptor) {
	if index, shift := desc/64, desc%64; int(index) < len(t.masks) {
		mask := t.masks[index]
		if (mask & (1 << shift)) != 0 {
			var zero Object
			t.table[desc] = zero
			t.masks[index] = mask & ^uint64(1<<shift)
		}
	}
}

// Range calls f for each object and its associated descriptor in the table.
// The function f might return false to interupt the iteration.
func (t *Table[Descriptor, Object]) Range(f func(Descriptor, Object) bool) {
	for i, mask := range t.masks {
		if mask == 0 {
			continue
		}
		for j := Descriptor(0); j < 64; j++ {
			if (mask & (1 << j)) == 0 {
				continue
			}
			if desc := Descriptor(i)*64 + j; !f(desc, t.table[desc]) {
				return
			}
		}
	}
}

// Reset clears the content of the table.
func (t *Table[Descriptor, Object]) Reset() {
	for i := range t.masks {
		t.masks[i] = 0
	}
	var zero Object
	for i := range t.table {
		t.table[i] = zero
	}
}
