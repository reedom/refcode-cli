package mapper

import (
	"encoding/binary"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// Store represents store(model).
type Store interface {
	// GetTime returns time value for the key.
	// If returns ErrNotFound if the key is not found in the store.
	GetTime(key string) (time.Time, error)
	// PutTime stores key and time pair to the store.
	PutTime(key string, t time.Time) error
}

type store struct {
	db *leveldb.DB
}

// NewStore returns new store object.
func NewStore(db *leveldb.DB) Store {
	return store{db}
}

// GetTime returns time value for the key.
// If returns ErrNotFound if the key is not found in the store.
func (s store) GetTime(key string) (time.Time, error) {
	v, err := s.db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return time.Time{}, ErrNotFound
	}

	ts := binary.LittleEndian.Uint64(v)
	return time.Unix(0, int64(ts)), nil
}

// PutTime stores key and time pair to the store.
func (s store) PutTime(key string, t time.Time) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(t.UnixNano()))
	return s.db.Put([]byte(key), buf[:], nil)
}
