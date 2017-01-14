package finder

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type walkFunc func(info fileInfo, ignores ignoreMatchers) (ignoreMatchers, error)

func concurrentWalk(ctx context.Context, root string, ignores ignoreMatchers, followed bool, walkFn walkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	sem := make(chan struct{}, 16)
	return walk(ctx, newFileInfo(root, info), ignores, followed, walkFn, sem)
}

func walk(ctx context.Context, info fileInfo, parentIgnores ignoreMatchers, followed bool, walkFn walkFunc, sem chan struct{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ignores, walkError := walkFn(info, parentIgnores)
	if walkError != nil {
		if info.IsDir() && walkError == filepath.SkipDir {
			return nil
		}
		return walkError
	}

	if !info.isDir(followed) {
		return nil
	}

	files, err := ioutil.ReadDir(info.path)
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	for _, file := range files {
		f := newFileInfo(filepath.Join(info.path, file.Name()), file)
		select {
		case sem <- struct{}{}:
			wg.Add(1)
			go func(file fileInfo, ignores ignoreMatchers, wg *sync.WaitGroup) {
				defer wg.Done()
				defer func() { <-sem }()
				walk(ctx, file, ignores, followed, walkFn, sem)
			}(f, ignores, wg)
		default:
			walk(ctx, f, ignores, followed, walkFn, sem)
		}
	}
	wg.Wait()
	return nil
}
