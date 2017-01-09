package refcode

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// Mapper replaces refcode template with reference code in each souce file.
type Mapper interface {
	Run(rootDir string) error
}

// MapperOpt is Mapper configuration.
type MapperOpt struct {
	Pattern string
	Replace string
	DryRun  bool

	InChannelCount int // 5000
	ParallelCount  int // 208
	WorkBufSize    int // 16*1024
}

type mapper struct {
	in      chan string
	opts    Option
	pattern *regexp.Regexp
	db      *leveldb.DB
	store   Store
}

// NewMapper returns Mapper object.
func NewMapper(opts Option) (Mapper, error) {
	pattern, err := regexp.Compile(opts.Mapper.Pattern)
	if err != nil {
		ErrorLog.Printf(`mapper.pattern = "%s"`, opts.Mapper.Pattern)
		return nil, err
	}

	db, err := leveldb.OpenFile(opts.storeDir(), nil)
	if err != nil {
		return nil, err
	}
	store := NewStore(db)

	return mapper{
		in:      make(chan string, opts.Mapper.InChannelCount),
		opts:    opts,
		pattern: pattern,
		db:      db,
		store:   store,
	}, nil
}

func (m mapper) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

// Run starts mapper.
func (m mapper) Run(rootDir string) error {
	finder := NewFileFinder(m.in, m.opts.FileFinder)
	go finder.Start(rootDir)

	errCh := make(chan error, 1)
	go func() {
		errCh <- m.run()
	}()

	return <-errCh
}

func (m mapper) run() error {
	sem := make(chan struct{}, m.opts.Mapper.ParallelCount)
	wg := &sync.WaitGroup{}

	for path := range m.in {
		sem <- struct{}{}
		wg.Add(1)
		go func(filepath string) {
			defer wg.Done()
			defer func() { <-sem }()
			m.proceed(filepath)
		}(path)
	}
	wg.Wait()
	return nil
}

func (m mapper) proceed(filepath string) {
	c := NewCodeMapper(filepath, m.pattern, m.opts.Mapper.Replace)
	c.Run(context.Background())
}

type CodeMapper interface {
	Run(ctx context.Context)
}

type codeMapper struct {
	filepath string
	pattern  *regexp.Regexp
	replace  string
}

func NewCodeMapper(filepath string, pattern *regexp.Regexp, replace string) CodeMapper {
	return codeMapper{
		filepath: filepath,
		pattern:  pattern,
		replace:  replace,
	}
}

func (c codeMapper) Run(ctx context.Context) {
	f, err := os.Open(c.filepath)
	if err != nil {
		ErrorLog.Printf("cannot open %v, %v", c.filepath, err)
		return
	}
	defer f.Close()

	_, err = CountMatch(ctx, f, c.pattern)
	if err != nil {
		if err == ErrBinaryFile {
			Verbose.Printf("skip binary file %q", c.filepath)
		} else {
			ErrorLog.Printf("prepare fail for file %q: %v", c.filepath, err)
		}
		return
	}

	f.Seek(0, io.SeekStart)
	c.replaceTemplates(ctx, f)
}

func (c codeMapper) replaceTemplates(ctx context.Context, r io.Reader) {
	out, err := ioutil.TempFile("", "refcode")
	if err != nil {
		ErrorLog.Print("failed to create temp file:", err)
		return
	}
	defer func() {
		if out.Close() == nil {
			os.Remove(out.Name())
		}
	}()

	Verbose.Print("create temp file", out.Name())

	err = TransformContent(ctx, r, out, c.pattern, c.rep)
	if err != nil {
		ErrorLog.Print("cannot write to disk:", err)
		return
	}

	out.Close()
	// os.Rename(out.Name(), c.filepath)
}

func (c codeMapper) rep([]byte) ([]byte, error) {
	return []byte("000-001"), nil
}
