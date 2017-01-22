package uniqid

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Generator interface {
	Generate(ctx context.Context, key, sub []byte, n int64) ([][]byte, error)
}

type fileStore struct {
	dir string
	a   Algorithm
	m   *sync.Mutex
}

func NewFileStore(dataDir string, a Algorithm) Generator {
	return &fileStore{dataDir, a, &sync.Mutex{}}
}

func (s *fileStore) Generate(ctx context.Context, key, sub []byte, n int64) ([][]byte, error) {
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

	if err = s.a.Load(f); err != nil {
		return nil, err
	}

	ids := make([][]byte, n)
	for i := int64(0); i < n; i++ {
		ids[i], err = s.a.NextValue()
		if err != nil {
			return nil, err
		}
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if err = s.a.Save(f); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *fileStore) getStoreDir(key []byte) string {
	hash := sha1.Sum(key)
	return filepath.Join(s.dir, fmt.Sprintf("%x", hash[:]))
}

func (s *fileStore) getFileName(sub []byte) string {
	hash := sha1.Sum(sub)
	return fmt.Sprintf("%x", hash[:])
}
