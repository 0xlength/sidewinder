//go:build !lite

package blockstore

import (
	"fmt"
)

// MaxRoot returns the last known root slot.
func (d *DB) MaxRoot() (uint64, error) {
	iter, err := d.CfRoot.NewIterator()
	if err != nil {
		return 0, err
	}
	defer iter.Close()
	iter.SeekToLast()
	if !iter.Valid() {
		return 0, ErrNotFound
	}
	slot, ok := ParseSlotKey(iter.Key())
	if !ok {
		return 0, fmt.Errorf("invalid key in root cf")
	}
	return slot, nil
}

// GetSlotMeta returns the shredding metadata of a given slot.
func (d *DB) GetSlotMeta(slot uint64) (*SlotMeta, error) {
	key := MakeSlotKey(slot)
	return GetBincode[SlotMeta](d.CfMeta, key[:])
}
