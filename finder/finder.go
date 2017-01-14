package finder

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/monochromegane/go-gitignore"
	"github.com/reedom/refcode-cli/log"
)

// FileFinder traverses directory and passes file paths.
type FileFinder interface {
	Start(ctx context.Context, out chan string)
}

type fileFinder struct {
	opts Option
	root string
}

// NewFileFinder returns FileFinder object.
func NewFileFinder(opts Option, root string) FileFinder {
	return fileFinder{opts, root}
}

// Start starts directory traverse.
func (f fileFinder) Start(ctx context.Context, out chan string) {
	defer close(out)
	var ignores ignoreMatchers

	if len(f.opts.Includes) == 0 {
		log.Verbose.Println("includes is empty; exit file finder")
		return
	}

	// add ignores from ignore option.
	if 0 < len(f.opts.Excludes) {
		ignores = append(ignores, gitignore.NewGitIgnoreFromReader(
			f.root,
			strings.NewReader(strings.Join(f.opts.Excludes, "\n")),
		))
	}

	// add global gitignore.
	if f.opts.GlobalGitIgnore {
		if ignore := globalGitIgnore(f.root); ignore != nil {
			log.Verbose.Println("use ~/.gitignore")
			ignores = append(ignores, ignore)
		}
	}

	includes := f.includes(f.root)
	walkFn := func(info fileInfo, ignores ignoreMatchers) (ignoreMatchers, error) {
		log.Verbose.Println("check path", info.path)
		if info.isDir(f.opts.FollowSymlinks) {
			if ignores.Match(info.path, true) {
				log.Verbose.Println("skip directory", info.path, "(matches with excludes/gitignore)")
				return ignores, filepath.SkipDir
			}

			if !f.opts.FollowHidden && isHidden(info.Name()) {
				return ignores, filepath.SkipDir
			}

			log.Verbose.Println("enter directory", info.path)
			ignores = append(ignores, newIgnoreMatchers(info.path, []string{".gitignore"})...)
			return ignores, nil
		}
		if !f.opts.FollowSymlinks && info.isSymlink() {
			log.Verbose.Println("skip symlink", info.path)
			return ignores, nil
		}

		if info.isNamedPipe() {
			return ignores, nil
		}

		if ignores.Match(info.path, false) {
			log.Verbose.Println("skip file", info.path, "(matches with excludes/gitignore)")
			return ignores, nil
		}

		if !includes.Match(info.path, false) {
			log.Verbose.Println("skip file", info.path, "(does not matche with includes)")
			return ignores, nil
		}

		out <- info.path
		return ignores, nil
	}
	concurrentWalk(ctx, f.root, ignores, f.opts.FollowSymlinks, walkFn)
}

func (f fileFinder) includes(root string) ignoreMatchers {
	if len(f.opts.Includes) == 0 {
		return make(ignoreMatchers, 0)
	}

	return ignoreMatchers{
		gitignore.NewGitIgnoreFromReader(
			root,
			strings.NewReader(strings.Join(f.opts.Includes, "\n"))),
	}
}

func isHidden(name string) bool {
	if name == "." || name == ".." {
		return false
	}
	return 1 < len(name) && name[0] == '.'
}
