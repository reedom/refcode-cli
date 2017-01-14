package mapper

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"

	"github.com/reedom/refcode-cli/log"
	"github.com/syndtr/goleveldb/leveldb"
)

type PathProvider interface {
	Start(ctx context.Context, out chan string)
}

type IDGenerator interface {
	Generate(ctx context.Context, key, sub []byte, n int) ([][]byte, error)
}

// Mapper replaces refcode template with reference code in each souce file.
type Mapper interface {
	Run(ctx context.Context) error
}

type mapper struct {
	in    chan string
	opts  Option
	db    *leveldb.DB
	store Store
	stat  mapperStat
	pp    PathProvider
	idgen IDGenerator
}

type mapperStat struct {
	readCount   int32
	mappedCount int32
}

func (s *mapperStat) incReadCount() {
	atomic.AddInt32(&s.readCount, 1)
}

func (s *mapperStat) incMappedCount() {
	atomic.AddInt32(&s.mappedCount, 1)
}

func (s *mapperStat) getReadCount() int32 {
	return atomic.LoadInt32(&s.readCount)
}

func (s *mapperStat) getMappedCount() int32 {
	return atomic.LoadInt32(&s.mappedCount)
}

// NewMapper returns Mapper object.
func NewMapper(opts Option, pp PathProvider, idgen IDGenerator) (Mapper, error) {
	db, err := leveldb.OpenFile(opts.storeDir(), nil)
	if err != nil {
		return nil, err
	}
	store := NewStore(db)

	return mapper{
		in:    make(chan string, opts.InChannelCount),
		opts:  opts,
		db:    db,
		store: store,
		pp:    pp,
		idgen: idgen,
	}, nil
}

func (m mapper) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m mapper) GetReadCount() int32 {
	return m.stat.getReadCount()
}

func (m mapper) GetMappedCount() int32 {
	return m.stat.getMappedCount()
}

// Run starts mapper.
func (m mapper) Run(ctx context.Context) error {
	go m.pp.Start(ctx, m.in)

	done := make(chan struct{})
	go func() {
		m.run(ctx)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Verbose.Printf("ctx.Done")
		return ctx.Err()
	case <-done:
		log.Verbose.Printf("done")
		return nil
	}
}

func (m mapper) run(ctx context.Context) {
	sem := make(chan struct{}, 208)
	wg := &sync.WaitGroup{}

	for path := range m.in {
		sem <- struct{}{}
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }()
			m.handle(ctx, path)
		}(path)
	}
	wg.Wait()
}

func (m mapper) handle(ctx context.Context, path string) {
	m.stat.incReadCount()
	info, err := os.Lstat(path)
	if err != nil {
		log.ErrorLog.Printf("Failed to stat file %q, %v", path, err)
		return
	}

	lastMTime, _ := m.store.GetTime(path)
	if lastMTime.Equal(info.ModTime()) {
		log.Verbose.Printf("skip unchanged file %q, %v", path, err)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		log.ErrorLog.Printf("Failed to open file %q, %v", path, err)
		return
	}
	defer f.Close()

	marker := []byte(m.opts.Marker)
	markerCount, err := CountMarkerInContent(ctx, f, marker)
	if err != nil {
		if err == ErrBinaryFile {
			log.Verbose.Printf("skip binary file %q", path)
			return
		}
		log.Verbose.Printf("cancel transform file %q, %v", path, err)
		return
	}
	if markerCount == 0 {
		log.Verbose.Printf("skip file %q, no marker found", path)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		log.ErrorLog.Printf("Failed to seek file %q, %v", path, err)
		return
	}

	transFn, err := m.createTransFn(ctx, path, markerCount)
	if err != nil {
		log.ErrorLog.Printf("Failed to prepare refcode mapper for %q, %v", path, err)
		return
	}

	out, err := ioutil.TempFile("", "refcode")
	if err != nil {
		log.ErrorLog.Printf("Failed to create temp file for %q,  %v", path, err)
		return
	}
	defer func() {
		if out.Close() != nil {
			os.Remove(out.Name())
		}
	}()

	log.Verbose.Printf("map refcode on %q", path)
	err = TransformContent(ctx, f, out, marker, transFn)
	if err != nil {
		log.ErrorLog.Printf("Error on map refcode on %q,  %v", path, err)
		return
	}

	err = out.Close()
	if err != nil {
		log.ErrorLog.Printf("Failed to close temp file for %q,  %v", path, err)
		return
	}
	err = os.Rename(out.Name(), path)
	if err != nil {
		log.ErrorLog.Printf("Failed to replace file %q,  %v", path, err)
		return
	}

	m.stat.incMappedCount()

	err = m.store.PutTime(path, info.ModTime())
	if err != nil {
		log.ErrorLog.Printf("Failed to store mtime of %q,  %v", path, err)
		return
	}
}

var refcode int

func (m mapper) createTransFn(ctx context.Context, path string, markerCount int) (TransFn, error) {
	codes, err := m.idgen.Generate(ctx, []byte(m.opts.Codespace), nil, markerCount)
	if err != nil {
		return nil, err
	}
	i := 0
	fn := func(ctx context.Context) ([]byte, error) {
		if len(codes) <= i {
			return nil, errors.New("source file is updated")
		}
		code := codes[i]
		i++
		return code, nil
	}
	return fn, nil
}
