package config

import "github.com/fasaxi-linker/servergo/internal/task"

// ParsedConfig represents the parsed configuration
type ParsedConfig struct {
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	KeepDirStruct bool     `json:"keepDirStruct"`
	OpenCache     bool     `json:"openCache"`
	MkdirIfSingle bool     `json:"mkdirIfSingle"`
	DeleteDir     bool     `json:"deleteDir"`
}

// Ensure ParsedConfig implements ConfigOptions interface
var _ task.ConfigOptions = (*ParsedConfig)(nil)

func (p *ParsedConfig) GetIncludePatterns() []string {
	return p.Include
}

func (p *ParsedConfig) GetExcludePatterns() []string {
	return p.Exclude
}

func (p *ParsedConfig) GetKeepDirStruct() bool {
	return p.KeepDirStruct
}

func (p *ParsedConfig) GetOpenCache() bool {
	return p.OpenCache
}

func (p *ParsedConfig) GetMkdirIfSingle() bool {
	return p.MkdirIfSingle
}

func (p *ParsedConfig) GetDeleteDir() bool {
	return p.DeleteDir
}