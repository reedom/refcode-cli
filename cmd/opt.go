package cmd

import (
	"github.com/reedom/refcode-cli/finder"
	"github.com/reedom/refcode-cli/mapper"
)

// Option bundles sub Optionurations.
type Option struct {
	Codespace  string
	DataDir    string
	Mapper     mapper.Option
	FileFinder finder.Option
	Remote     RemoteOpt
}

// RemoteOpt Optionures API remote of the reference code management server.
type RemoteOpt struct {
	Endpoint  string
	SecretKey string
}
