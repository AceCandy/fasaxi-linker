package task

// ConfigOptions represents configuration options that can be applied to a task
type ConfigOptions interface {
	GetIncludePatterns() []string
	GetExcludePatterns() []string
	GetKeepDirStruct() bool
	GetOpenCache() bool
	GetMkdirIfSingle() bool
	GetDeleteDir() bool
}

// RuntimeConfig represents the parsed configuration used at runtime
type RuntimeConfig struct {
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	KeepDirStruct bool     `json:"keepDirStruct"`
	OpenCache     bool     `json:"openCache"`
	MkdirIfSingle bool     `json:"mkdirIfSingle"`
	DeleteDir     bool     `json:"deleteDir"`
}

// Ensure RuntimeConfig implements ConfigOptions interface
var _ ConfigOptions = (*RuntimeConfig)(nil)

func (r *RuntimeConfig) GetIncludePatterns() []string {
	return r.Include
}

func (r *RuntimeConfig) GetExcludePatterns() []string {
	return r.Exclude
}

func (r *RuntimeConfig) GetKeepDirStruct() bool {
	return r.KeepDirStruct
}

func (r *RuntimeConfig) GetOpenCache() bool {
	return r.OpenCache
}

func (r *RuntimeConfig) GetMkdirIfSingle() bool {
	return r.MkdirIfSingle
}

func (r *RuntimeConfig) GetDeleteDir() bool {
	return r.DeleteDir
}