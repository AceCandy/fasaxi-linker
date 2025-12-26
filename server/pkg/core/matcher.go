package core

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Supported checks if a file path is supported based on include and exclude patterns.
// It tries to mimic micromatch behavior:
// 1. Filter out hidden files (starting with .)
// 2. If include is empty, it assumes match all (unless excluded).
// 3. Path must match at least one include pattern.
// 4. Path must NOT match any exclude pattern.
func Supported(path string, include []string, exclude []string) bool {
	// Filter out hidden files (files and directories starting with .)
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return false
	}

	// Normalize path separators to forward slashes for globbing
	// path = filepath.ToSlash(path) // Optional, doublestar handles OS separators usually, but consistent is better.
	// Actually doublestar recommends forward slashes for patterns.

	// Check exclusion first
	for _, pattern := range exclude {
		// For patterns without path separators, match against basename
		// For patterns with path separators, match against full path
		if strings.Contains(pattern, "/") {
			if match, _ := doublestar.PathMatch(pattern, path); match {
				return false
			}
		} else {
			// Match against basename for simple patterns like *.tmp
			if match, _ := doublestar.Match(pattern, base); match {
				return false
			}
		}
	}

	// If no include patterns are provided, we assume everything is included (unless excluded above)
	// This matches the original logic where empty include defaults to ['**']
	if len(include) == 0 {
		return true
	}

	for _, pattern := range include {
		// For patterns without path separators, match against basename
		// For patterns with path separators, match against full path
		if strings.Contains(pattern, "/") {
			if match, _ := doublestar.PathMatch(strings.ToLower(pattern), strings.ToLower(path)); match {
				return true
			}
		} else {
			// Match against basename for simple patterns like *.js
			if match, _ := doublestar.Match(strings.ToLower(pattern), strings.ToLower(base)); match {
				return true
			}
		}
	}

	return false
}
