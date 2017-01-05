package refcode

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

// CacheOption Optionures cache functionality.
type CacheOption struct {
	CachePath string
}
