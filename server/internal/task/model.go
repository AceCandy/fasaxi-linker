package task

import (
	"fmt"
	"github.com/fasaxi-linker/servergo/pkg/core"
)

// Task represents a task configuration
type Task struct {
	Name          string        `json:"name"`
	Type          string        `json:"type"` // "main" or "prune"
	PathsMapping  []PathMapping `json:"pathsMapping"`
	Include       []string      `json:"include"`
	Exclude       []string      `json:"exclude"`
	SaveMode      int           `json:"saveMode"`
	OpenCache     bool          `json:"openCache"`
	MkdirIfSingle bool          `json:"mkdirIfSingle"`
	DeleteDir     bool          `json:"deleteDir"`
	KeepDirStruct bool          `json:"keepDirStruct"`
	ScheduleType  string        `json:"scheduleType,omitempty"`
	ScheduleValue string        `json:"scheduleValue,omitempty"`
	Reverse       bool          `json:"reverse,omitempty"` // for prune
	Config        string        `json:"config"`
	IsWatching    bool          `json:"isWatching"`
	WatchError    string        `json:"watchError,omitempty"` // Watch failure reason
}

type PathMapping struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

// ToCoreOptions converts Task to core.Options
func (t *Task) ToCoreOptions() core.Options {
	pm := make(map[string][]string)
	for _, m := range t.PathsMapping {
		if _, ok := pm[m.Source]; !ok {
			pm[m.Source] = []string{}
		}
		pm[m.Source] = append(pm[m.Source], m.Dest)
	}

	opts := core.Options{
		Name:          t.Name,
		Type:          t.Type,
		PathsMapping:  pm,
		Include:       t.Include,
		Exclude:       t.Exclude,
		SaveMode:      t.SaveMode,
		OpenCache:     t.OpenCache,
		MkdirIfSingle: t.MkdirIfSingle,
		DeleteDir:     t.DeleteDir,
		KeepDirStruct: t.KeepDirStruct,
	}
	// Debug: print cache status
	if opts.OpenCache {
		fmt.Printf("DEBUG: Cache is enabled for task %s\n", t.Name)
	}
	return opts
}

// ToCoreOptionsWithConfig converts Task to core.Options with associated config
func (t *Task) ToCoreOptionsWithConfig(config ConfigOptions) core.Options {
	pm := make(map[string][]string)
	for _, m := range t.PathsMapping {
		if _, ok := pm[m.Source]; !ok {
			pm[m.Source] = []string{}
		}
		pm[m.Source] = append(pm[m.Source], m.Dest)
	}

	// Use patterns directly from config
	var includePatterns []string
	if config != nil {
		includePatterns = config.GetIncludePatterns()
	}
	
	var excludePatterns []string
	if config != nil {
		excludePatterns = config.GetExcludePatterns()
	}

	opts := core.Options{
		Name:          t.Name,
		Type:          t.Type,
		PathsMapping:  pm,
		Include:       includePatterns,
		Exclude:       excludePatterns,
		SaveMode:      t.SaveMode,
		OpenCache:     config != nil && config.GetOpenCache(),
		MkdirIfSingle: config != nil && config.GetMkdirIfSingle(),
		DeleteDir:     config != nil && config.GetDeleteDir(),
		KeepDirStruct: config != nil && config.GetKeepDirStruct(),
	}
	
	return opts
}
