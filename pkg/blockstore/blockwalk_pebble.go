//go:build !lite

package blockstore

import (
	"fmt"
	"sort"

	"github.com/0xlength/sidewinder/pkg/shred"
	"k8s.io/klog/v2"
)

// BlockWalk walks blocks in ascending order over multiple blockstore databases.
type BlockWalk struct {
	handles       []WalkHandle // sorted
	shredRevision int

	root *Iterator
}

func NewBlockWalk(handles []WalkHandle, shredRevision int) (*BlockWalk, error) {
	if err := sortWalkHandles(handles, shredRevision); err != nil {
		return nil, err
	}
	return &BlockWalk{
		handles:       handles,
		shredRevision: shredRevision,
	}, nil
}

// Seek skips ahead to a specific slot.
// The caller must call BlockWalk.Next after Seek.
func (m *BlockWalk) Seek(slot uint64) bool {
	for len(m.handles) > 0 {
		h := m.handles[0]
		if slot < h.Start {
			return false
		}
		if slot <= h.Stop {
			h.Start = slot
			return true
		}
		m.pop()
	}
	return false
}

// SlotsAvailable returns the number of contiguous slots that lay ahead.
func (m *BlockWalk) SlotsAvailable() (total uint64) {
	if len(m.handles) == 0 {
		return 0
	}
	start := m.handles[0].Start
	for _, h := range m.handles {
		if h.Start > start {
			return
		}
		stop := h.Stop + 1
		total += stop - start
		start = stop
	}
	return
}

// Next seeks to the next slot.
func (m *BlockWalk) Next() (meta *SlotMeta, ok bool) {
	if len(m.handles) == 0 {
		return nil, false
	}
	h := m.handles[0]
	if m.root == nil {
		iter, err := h.DB.CfRoot.NewIterator()
		if err != nil {
			klog.Errorf("FATAL: cannot open root iterator: %s", err)
			return nil, false
		}
		m.root = iter
		key := MakeSlotKey(h.Start)
		m.root.Seek(key[:])
	}
	if !m.root.Valid() {
		m.pop()
		return m.Next()
	}

	slot, ok := ParseSlotKey(m.root.Key())
	if !ok {
		klog.Exitf("Invalid slot key: %x", m.root.Key())
	}
	if slot > h.Stop {
		m.pop()
		return m.Next()
	}
	h.Start = slot

	var err error
	meta, err = h.DB.GetSlotMeta(slot)
	if err != nil {
		klog.Errorf("FATAL: invalid slot meta at slot %d, aborting CAR generation: %s", slot, err)
		return nil, false
	}

	m.root.Next()

	return meta, true
}

func (m *BlockWalk) Current() *DB {
	if len(m.handles) == 0 {
		return nil
	}
	return m.handles[0].DB
}

// Entries returns the entries at the current cursor.
// Caller must have made an ok call to BlockWalk.Next before calling this.
func (m *BlockWalk) Entries(meta *SlotMeta) ([][]shred.Entry, error) {
	h := m.handles[0]
	mapping, err := h.DB.GetEntries(meta, m.shredRevision)
	if err != nil {
		return nil, err
	}
	batches := make([][]shred.Entry, len(mapping))
	for i, batch := range mapping {
		batches[i] = batch.Entries
	}
	return batches, nil
}

func (m *BlockWalk) pop() {
	m.root.Close()
	m.root = nil
	m.handles[0].DB.Close()
	m.handles = m.handles[1:]
}

func (m *BlockWalk) Close() {
	if m.root != nil {
		m.root.Close()
		m.root = nil
	}
	for _, h := range m.handles {
		h.DB.Close()
	}
	m.handles = nil
}

type WalkHandle struct {
	DB    *DB
	Start uint64
	Stop  uint64 // inclusive
}

func sortWalkHandles(h []WalkHandle, shredRevision int) error {
	for i, db := range h {
		start, err := getLowestCompletedSlot(db.DB, shredRevision)
		if err != nil {
			return err
		}
		stop, err := db.DB.MaxRoot()
		if err != nil {
			return err
		}
		h[i] = WalkHandle{
			Start: start,
			Stop:  stop,
			DB:    db.DB,
		}
	}
	sort.Slice(h, func(i, j int) bool {
		return h[i].Start < h[j].Start
	})
	return nil
}

func getLowestCompletedSlot(d *DB, shredRevision int) (uint64, error) {
	iter, err := d.CfMeta.NewIterator()
	if err != nil {
		return 0, err
	}
	defer iter.Close()
	iter.SeekToFirst()

	const maxTries = 32
	for i := 0; iter.Valid() && i < maxTries; i++ {
		slot, ok := ParseSlotKey(iter.Key())
		if !ok {
			return 0, fmt.Errorf(
				"getLowestCompletedSlot(%s): choked on invalid slot key: %x", d.Name(), iter.Key())
		}

		meta, err := ParseBincode[SlotMeta](iter.Value())
		if err != nil {
			return 0, fmt.Errorf(
				"getLowestCompletedSlot(%s): choked on invalid meta for slot %d", d.Name(), slot)
		}

		if _, err = d.GetEntries(meta, shredRevision); err == nil {
			return slot, nil
		}

		iter.Next()
	}

	return 0, fmt.Errorf("failed to find a valid complete slot in DB: %s", d.Name())
}
