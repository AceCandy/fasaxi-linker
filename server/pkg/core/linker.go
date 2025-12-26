package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetOriginalDestPath calculates the destination directory path
func GetOriginalDestPath(sourceFile, source, dest string, keepDirStruct bool, mkdirIfSingle bool) (string, error) {
	currentDir := filepath.Dir(sourceFile)
	currentName := filepath.Base(sourceFile)

	absSource, err := filepath.Abs(source)
	if err != nil {
		return "", err
	}
	absCurrentDir, err := filepath.Abs(currentDir)
	if err != nil {
		return "", err
	}

	relativePath, err := filepath.Rel(absSource, absCurrentDir)
	if err != nil {
		return "", err
	}

	if mkdirIfSingle && (relativePath == "." || relativePath == "") {
		ext := filepath.Ext(currentName)
		relativePath = strings.TrimSuffix(currentName, ext)
	}

	// Logic port from JS: relativePath.split(sep).slice(Number(keepDirStruct) - 1).join(sep)
	// If keepDirStruct is true (1), slice(0) -> keep all.
	// If keepDirStruct is false (0), slice(-1) -> keep last one.
	parts := strings.Split(relativePath, string(os.PathSeparator))
	
	// Handle "." case from Rel
	if len(parts) == 1 && parts[0] == "." {
		parts = []string{}
	}

	var finalParts []string
	if keepDirStruct {
		finalParts = parts
	} else {
		if len(parts) > 0 {
			finalParts = parts[len(parts)-1:]
		} else {
			finalParts = []string{}
		}
	}

	return filepath.Join(dest, filepath.Join(finalParts...)), nil
}

// Link creates a hard link
func Link(sourceFile, destDir string) (string, error) {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	targetFile := filepath.Join(destDir, filepath.Base(sourceFile))

	// Check if target exists
	if _, err := os.Stat(targetFile); err == nil {
		return targetFile, fmt.Errorf("file exists: %s", targetFile)
	}

	// Create hard link
	if err := os.Link(sourceFile, targetFile); err != nil {
		return targetFile, err
	}

	return targetFile, nil
}
