package refcode

import (
	"path/filepath"
)

// Option bundles sub Optionurations.
type Option struct {
	Codespace  string
	DataDir    string
	Mapper     MapperOpt
	FileFinder FileFinderOpt
	Remote     RemoteOpt
	Cache      CacheOption
}

// RemoteOpt Optionures API remote of the reference code management server.
type RemoteOpt struct {
	Endpoint  string
	SecretKey string
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

// FileFinderOpt is FileFinder configuration.
type FileFinderOpt struct {
	Includes        []string
	Excludes        []string
	GlobalGitIgnore bool
	FollowSymlinks  bool
}

// CacheOption Optionures cache functionality.
type CacheOption struct {
	CachePath string
}

func (o Option) storeDir() string {
	return filepath.Join(o.DataDir, "leveldb")
}
