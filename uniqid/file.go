package uniqid

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Generator interface {
	Generate(ctx context.Context, key, sub []byte, n int) ([][]byte, error)
}

type seqFileStore struct {
	dir string
	m   *sync.Mutex
}

func NewFileSeq(dataDir string) Generator {
	return &seqFileStore{dataDir, &sync.Mutex{}}
}

func (s *seqFileStore) Generate(ctx context.Context, key, sub []byte, n int) ([][]byte, error) {
	s.m.Lock()
	defer s.m.Unlock()

	storeDir := s.getStoreDir(key)
	filename := s.getFileName(sub)
	fullpath := filepath.Join(storeDir, filename)
	os.MkdirAll(storeDir, 0777)

	f, err := os.OpenFile(fullpath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c uint64
	if err = binary.Read(f, binary.LittleEndian, &c); err != nil {
		if err != io.EOF {
			return nil, err
		}
	}

	ids := make([][]byte, n)
	for i := 0; i < n; i++ {
		c++
		ids[i] = strconv.AppendUint(nil, c, 10)
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if err = binary.Write(f, binary.LittleEndian, c); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *seqFileStore) getStoreDir(key []byte) string {
	hash := sha1.Sum(key)
	return filepath.Join(s.dir, fmt.Sprintf("%x", hash[:]))
}

func (s *seqFileStore) getFileName(sub []byte) string {
	hash := sha1.Sum(sub)
	return fmt.Sprintf("%x", hash[:])
}
