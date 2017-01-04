package refcode

import (
	"path/filepath"
	"strings"

	"github.com/monochromegane/go-gitignore"
)

// FileFinder traverses directory and passes file paths through out channel.
type FileFinder interface {
	Start(root string)
}

// FileFinderOpt is FileFinder configurations.
type FileFinderOpt struct {
	Includes        []string
	Excludes        []string
	FollowSymlinks  bool
	GlobalGitIgnore bool
	Verbose         bool
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
func (f fileFinder) Start(root string) {
	defer close(f.out)
	var ignores ignoreMatchers

	if len(f.opts.Includes) == 0 {
		Verbose.Println("includes is empty; exit file finder")
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
			Verbose.Println("use ~/.gitignore")
			ignores = append(ignores, ignore)
		}
	}

	includes := f.includes(root)
	concurrentWalk(root, ignores, f.opts.FollowSymlinks, func(info fileInfo, depth int, ignores ignoreMatchers) (ignoreMatchers, error) {
		Verbose.Println("check path", info.relpath)
		if info.isDir(f.opts.FollowSymlinks) {
			if ignores.Match(info.relpath, true) {
				Verbose.Println("skip directory", info.relpath, "(matches with excludes/gitignore)")
				return ignores, filepath.SkipDir
			}

			Verbose.Println("enter directory", info.relpath)
			ignores = append(ignores, newIgnoreMatchers(info.relpath, []string{".gitignore"})...)
			return ignores, nil
		}
		if !f.opts.FollowSymlinks && info.isSymlink() {
			Verbose.Println("skip symlink", info.relpath)
			return ignores, nil
		}

		if info.isNamedPipe() {
			return ignores, nil
		}

		if ignores.Match(info.relpath, false) {
			Verbose.Println("skip file", info.relpath, "(matches with excludes/gitignore)")
			return ignores, nil
		}

		if !includes.Match(info.relpath, false) {
			Verbose.Println("skip file", info.relpath, "(does not matche with includes)")
			return ignores, nil
		}

		f.out <- info.relpath
		return ignores, nil
	})
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
