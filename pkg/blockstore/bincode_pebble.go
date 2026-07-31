//go:build !lite

package blockstore

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble"
)

func GetBincode[T any](cf *Column, key []byte) (*T, error) {
	val, closer, err := cf.DB.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()
	return ParseBincode[T](val)
}

func MultiGetBincode[T any](cf *Column, key ...[]byte) ([]*T, error) {
	vals := make([]*T, len(key))
	for i, k := range key {
		val, closer, err := cf.DB.Get(k)
		if err != nil {
			if errors.Is(err, pebble.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		parsed, err := ParseBincode[T](val)
		closer.Close()
		if err != nil {
			fmt.Printf("cannot decode %s: %s", hex.EncodeToString(k), err)
			return nil, err
		}
		vals[i] = parsed
	}
	return vals, nil
}

func getBytes(cf *Column, key []byte) ([]byte, error) {
	val, closer, err := cf.DB.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()
	out := make([]byte, len(val))
	copy(out, val)
	return out, nil
}
