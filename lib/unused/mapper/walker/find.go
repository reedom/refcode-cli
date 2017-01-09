package refcode

import (
	"path/filepath"
	"strings"

	"context"
	"github.com/monochromegane/go-gitignore"
	"github.com/reedom/refcode-cli/lib/fs"
	"github.com/reedom/refcode-cli/lib/refcode"
)

// FileFinder traverses directory and passes file paths through out channel.
type FileFinder interface {
	Start(root string)
}

type fileFinder struct {
	out  chan string
	opts refcode.FileFinderOpt
}

// NewFileFinder returns FileFinder object.
func NewFileFinder(out chan string, opts refcode.FileFinderOpt) FileFinder {
	return fileFinder{out, opts}
}

// Start starts directory traverse.
func (f fileFinder) Start(ctx context.Context, root string) {
	defer close(f.out)

	if len(f.opts.Includes) == 0 {
		refcode.Verbose.Println("includes is empty; exit file finder")
		return
	}

	var excludes ignoreMatchers

	// add excludes from ignore option.
	if 0 < len(f.opts.Excludes) {
		excludes = append(excludes, gitignore.NewGitIgnoreFromReader(
			root,
			strings.NewReader(strings.Join(f.opts.Excludes, "\n")),
		))
	}

	// add global gitignore.
	if f.opts.GlobalGitIgnore {
		if ignore := globalGitIgnore(root); ignore != nil {
			refcode.Verbose.Println("use ~/.gitignore")
			excludes = append(excludes, ignore)
		}
	}

	includes := f.includes(root)
	f.find(ctx, root, includes, excludes)
}

func (f fileFinder) find(ctx context.Context, root string, includes, excludes ignoreMatchers) {
	var walkFn fs.WalkFn
	walkFn = func(ctx context.Context, found fs.FoundFile) error {
		refcode.Verbose.Println("check path", found.Path())
		if found.IsDir() {
			if excludes.Match(found.Path(), true) {
				refcode.Verbose.Println("skip directory", found.Path(), "(matches with excludes/gitignore)")
				return filepath.SkipDir
			}

			refcode.Verbose.Println("enter directory", found.Path())
			excludes = append(excludes, newIgnoreMatchers(found.Path(), []string{".gitignore"})...)
			return excludes, nil
		}
		if !f.opts.FollowSymlinks && info.isSymlink() {
			refcode.Verbose.Println("skip symlink", found.Path())
			return excludes, nil
		}

		if info.isNamedPipe() {
			return excludes, nil
		}

		if excludes.Match(found.Path(), false) {
			refcode.Verbose.Println("skip file", found.Path(), "(matches with excludes/gitignore)")
			return excludes, nil
		}

		if !includes.Match(found.Path(), false) {
			refcode.Verbose.Println("skip file", found.Path(), "(does not matche with includes)")
			return excludes, nil
		}

		f.out <- found.Path()
		return excludes, nil
	}

	fs.Walk(ctx, root, walkFn)
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
