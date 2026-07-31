//go:build !lite

package blockstore

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/cockroachdb/pebble"
)

// requiredColumns are column families that must be present to open a blockstore.
var requiredColumns = []string{CfMeta, CfRoot, CfDataShred, CfCodeShred}

// DB is a Pebble-backed Solana blockstore.
//
// Layout: each column family is a Pebble database under <path>/<cf_name>/.
// This replaces RocksDB column families and is not compatible with Agave/Labs RocksDB ledgers.
type DB struct {
	Path string

	CfDefault   *Column
	CfMeta      *Column
	CfRoot      *Column
	CfDataShred *Column
	CfCodeShred *Column
	CfTxStatus  *Column

	columns []*Column
}

// Column is one logical column family backed by its own Pebble DB.
type Column struct {
	Name string
	DB   *pebble.DB
}

func OpenReadWrite(path string) (*DB, error) {
	return open(path, true)
}

// OpenReadOnly attaches to a blockstore in read-only mode.
func OpenReadOnly(path string) (*DB, error) {
	return open(path, false)
}

// OpenSecondary opens a blockstore read-only.
//
// Pebble has no RocksDB-style secondary instance; secondaryPath is ignored.
func OpenSecondary(path string, secondaryPath string) (*DB, error) {
	_ = secondaryPath
	return open(path, false)
}

func open(path string, write bool) (*DB, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}

	names, err := ListColumnFamilies(path)
	if err != nil {
		return nil, err
	}
	if write {
		names = mergeColumnNames(names, requiredColumns)
	}

	db := &DB{Path: path}
	opts := &pebble.Options{ReadOnly: !write}

	for _, name := range names {
		cfPath := filepath.Join(path, name)
		if write {
			if err := os.MkdirAll(cfPath, 0o755); err != nil {
				db.Close()
				return nil, err
			}
		}
		pdb, err := pebble.Open(cfPath, opts)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("open column %q: %w", name, err)
		}
		col := &Column{Name: name, DB: pdb}
		db.columns = append(db.columns, col)
		bindColumn(db, col)
	}

	if db.CfMeta == nil {
		db.Close()
		return nil, errors.New("missing column family " + CfMeta)
	}
	if db.CfRoot == nil {
		db.Close()
		return nil, errors.New("missing column family " + CfRoot)
	}
	if db.CfDataShred == nil {
		db.Close()
		return nil, errors.New("missing column family " + CfDataShred)
	}
	if db.CfCodeShred == nil {
		db.Close()
		return nil, errors.New("missing column family " + CfCodeShred)
	}

	return db, nil
}

func bindColumn(db *DB, col *Column) {
	switch col.Name {
	case CfDefault:
		db.CfDefault = col
	case CfMeta:
		db.CfMeta = col
	case CfRoot:
		db.CfRoot = col
	case CfDataShred:
		db.CfDataShred = col
	case CfCodeShred:
		db.CfCodeShred = col
	case CfTxStatus:
		db.CfTxStatus = col
	}
}

func mergeColumnNames(existing, required []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(required))
	out := make([]string, 0, len(existing)+len(required))
	for _, n := range existing {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	for _, n := range required {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// ListColumnFamilies returns column family names under path (subdirs with a Pebble CURRENT file).
func ListColumnFamilies(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		current := filepath.Join(path, e.Name(), "CURRENT")
		if _, err := os.Stat(current); err != nil {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// Columns returns all opened column families.
func (d *DB) Columns() []*Column {
	return d.columns
}

// Name returns the blockstore directory path.
func (d *DB) Name() string {
	return d.Path
}

// Flush flushes memtables for all columns.
func (d *DB) Flush() error {
	for _, col := range d.columns {
		if err := col.DB.Flush(); err != nil {
			return fmt.Errorf("flush %s: %w", col.Name, err)
		}
	}
	return nil
}

// Compact compacts all column families.
func (d *DB) Compact() error {
	// Exclusive end bound larger than any Solana blockstore key (≤16 bytes).
	end := bytes.Repeat([]byte{0xff}, 17)
	for _, col := range d.columns {
		if err := col.DB.Compact(nil, end, false); err != nil {
			return fmt.Errorf("compact %s: %w", col.Name, err)
		}
	}
	return nil
}

func (d *DB) Close() {
	for _, col := range d.columns {
		if col != nil && col.DB != nil {
			_ = col.DB.Close()
			col.DB = nil
		}
	}
	d.columns = nil
}

// NewIterator returns an iterator over the column.
func (c *Column) NewIterator() (*Iterator, error) {
	iter, err := c.DB.NewIter(nil)
	if err != nil {
		return nil, err
	}
	return &Iterator{Iterator: iter}, nil
}
