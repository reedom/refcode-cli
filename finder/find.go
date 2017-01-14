package finder

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/monochromegane/go-gitignore"
	"github.com/reedom/refcode-cli/log"
)

// FileFinder traverses directory and passes file paths through out channel.
type FileFinder interface {
	Start(ctx context.Context, root string)
}

// FileFinderOpt is FileFinder configuration.
type FileFinderOpt struct {
	Includes        []string
	Excludes        []string
	GlobalGitIgnore bool
	FollowSymlinks  bool
	FollowHidden    bool
}

type fileFinder struct {
	out  chan string
	opts FileFinderOpt
}

// NewFileFinder returns FileFinder object.
func NewFileFinder(out chan string, opts FileFinderOpt) FileFinder {
	return fileFinder{out, opts}
}

// Start starts directory traverse.
func (f fileFinder) Start(ctx context.Context, root string) {
	defer close(f.out)
	var ignores ignoreMatchers

	if len(f.opts.Includes) == 0 {
		log.Verbose.Println("includes is empty; exit file finder")
		return
	}

	// add ignores from ignore option.
	if 0 < len(f.opts.Excludes) {
		ignores = append(ignores, gitignore.NewGitIgnoreFromReader(
			root,
			strings.NewReader(strings.Join(f.opts.Excludes, "\n")),
		))
	}

	// add global gitignore.
	if f.opts.GlobalGitIgnore {
		if ignore := globalGitIgnore(root); ignore != nil {
			log.Verbose.Println("use ~/.gitignore")
			ignores = append(ignores, ignore)
		}
	}

	includes := f.includes(root)
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

		f.out <- info.path
		return ignores, nil
	}
	concurrentWalk(ctx, root, ignores, f.opts.FollowSymlinks, walkFn)
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
