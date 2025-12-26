package core

// Options defines the task configuration
type Options struct {
	Name          string              `json:"name"`
	Type          string              `json:"type"` // "main" or "prune"
	PathsMapping  map[string][]string `json:"pathsMapping"`
	Include       []string            `json:"include"`
	Exclude       []string            `json:"exclude"`
	SaveMode      int                 `json:"saveMode"` // 0: keepDirStruct, 1: flatten? (Check logic)
	OpenCache     bool                `json:"openCache"`
	MkdirIfSingle bool                `json:"mkdirIfSingle"`
	DeleteDir     bool                `json:"deleteDir"` // for prune
	KeepDirStruct bool                `json:"keepDirStruct"`
}

// Stats holds execution statistics
type Stats struct {
	SuccessCount int
	FailCount    int
	FailFiles    map[string][]string
}
