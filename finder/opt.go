package finder

// Option is FileFinder configuration.
type Option struct {
	Includes        []string
	Excludes        []string
	GlobalGitIgnore bool
	FollowSymlinks  bool
	FollowHidden    bool
}
