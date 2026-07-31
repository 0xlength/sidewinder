//go:build !lite

package blockstore

import (
	"bytes"

	"github.com/cockroachdb/pebble"
)

// Iterator wraps pebble.Iterator with RocksDB-like helpers used by this package.
type Iterator struct {
	*pebble.Iterator
}

// Seek positions the iterator at the first key >= key.
func (it *Iterator) Seek(key []byte) bool {
	return it.SeekGE(key)
}

// SeekToFirst positions at the first key.
func (it *Iterator) SeekToFirst() bool {
	return it.First()
}

// SeekToLast positions at the last key.
func (it *Iterator) SeekToLast() bool {
	return it.Last()
}

// ValidForPrefix reports whether the iterator is valid and the current key has prefix.
func (it *Iterator) ValidForPrefix(prefix []byte) bool {
	return it.Valid() && bytes.HasPrefix(it.Key(), prefix)
}

// Close closes the iterator.
func (it *Iterator) Close() {
	if it.Iterator != nil {
		_ = it.Iterator.Close()
		it.Iterator = nil
	}
}
