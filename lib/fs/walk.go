package fs

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

// ConcurrentCount defines max concurrency in Walk job.
// You can modify the value before the first call of Walk.
var ConcurrentCount = runtime.GOMAXPROCS(0)

// sem is semaphore, it is used to control max job concurrency.
// The value will be created by the first call of Walk, based on
// the value of ConcurrentCount.
var sem chan struct{}
var once sync.Once

type WalkFn func(ctx context.Context, f FoundFile) error

// Walk walks the file tree rooted at root, calling WalkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by WalkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func Walk(ctx context.Context, root string, walkFn WalkFn) error {
	f, err := newFoundFile(root)
	if err != nil {
		return err
	}

	once.Do(func() { sem = make(chan struct{}, ConcurrentCount) })
	return walk(ctx, f, walkFn)
}

func walk(ctx context.Context, f FoundFile, walkFn WalkFn) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := walkFn(ctx, f)
	if err != nil {
		if f.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !f.IsDir() {
		return nil
	}

	names, err := readDirNames(f.Path())
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	for _, name := range names {
		filename := filepath.Join(f.Path(), name)
		info, err := newFoundFile(filename)
		if err != nil {
			continue
		}

		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(f FoundFile, wg *sync.WaitGroup) {
				defer wg.Done()
				defer func() { <-sem }()
				walk(ctx, f, walkFn)
			}(info, wg)
		default:
			walk(ctx, info, walkFn)
		}
	}
	wg.Wait()
	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
