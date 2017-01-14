package mapper

import (
	"path/filepath"
)

// Option is Mapper configuration.
type Option struct {
	Codespace string
	DataDir   string

	Marker        string
	ReplaceFormat string
	DryRun        bool

	InChannelCount int // 5000
	ParallelCount  int // 208
	WorkBufSize    int // 16*1024
}

func (o Option) storeDir() string {
	return filepath.Join(o.DataDir, "store")
}
