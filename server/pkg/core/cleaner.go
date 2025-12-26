package core

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// FileInfo holds path and inode
type FileInfo struct {
	Path  string
	Inode uint64
}

// GetInodes scans directories and returns map of inodes
func GetInodes(paths []string) (map[uint64]bool, error) {
	inodes := make(map[uint64]bool)
	for _, root := range paths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err // or ignore?
			}
			if !d.IsDir() {
				info, err := d.Info()
				if err != nil {
					return nil
				}
				stat, ok := info.Sys().(*syscall.Stat_t)
				if ok {
					inodes[stat.Ino] = true
				}
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			// Log error but continue?
			fmt.Printf("Error scanning %s: %v\n", root, err)
		}
	}
	return inodes, nil
}

// ScanFiles returns all files in directories with metadata
func ScanFiles(paths []string) ([]FileInfo, error) {
	var files []FileInfo
	for _, root := range paths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				info, err := d.Info()
				if err != nil {
					return nil
				}
				stat, ok := info.Sys().(*syscall.Stat_t)
				if ok {
					files = append(files, FileInfo{
						Path:  path,
						Inode: stat.Ino,
					})
				}
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error scanning %s: %v\n", root, err)
		}
	}
	return files, nil
}

// GetPruneFiles identifies files to be deleted
func GetPruneFiles(opts Options) ([]string, error) {
	// Source paths = keys of PathsMapping
	var sourcePaths []string
	for k := range opts.PathsMapping {
		sourcePaths = append(sourcePaths, k)
	}
	
	// Dest paths = values of PathsMapping
	var destPaths []string
	for _, v := range opts.PathsMapping {
		destPaths = append(destPaths, v...)
	}

	// 1. Get Source Inodes
	sourceInodes, err := GetInodes(sourcePaths)
	if err != nil {
		return nil, err
	}

	// 2. Scan Dest Files
	destFiles, err := ScanFiles(destPaths)
	if err != nil {
		return nil, err
	}

	var toDelete []string

	// 3. Filter
	for _, f := range destFiles {
		// If dest file matches Exclude -> Skip (Supported logic inside)
		// But Supported checks "Is this file Supported?". 
		// If Supported returns false (meaning it IS excluded or NOT included), 
		// then in "prune", do we ignore it (keep it safe) or delete it?
		// JS logic:
		// .filter((item) => { return !inodes.includes(item.inode) }) // Orphan
		// .filter((item) => { return supported(...) }) // Only prune files that match the rules
		
		// If a file is Orphan AND it is "Supported" (i.e. we manage this type of file), THEN we delete it.
		// If it is Orphan but "Excluded" (e.g. .DS_Store), we leave it alone.
		
		isOrphan := !sourceInodes[f.Inode]
		if isOrphan {
			if Supported(f.Path, opts.Include, opts.Exclude) {
				toDelete = append(toDelete, f.Path)
			}
		}
	}

	return toDelete, nil
}

// DeleteEmptyDirs deletes empty directories recursively
func DeleteEmptyDirs(paths []string) error {
	// Simple implementation using find command like JS version
	// Or we can implement in Go. JS version finds common parent.
	if len(paths) == 0 {
		return nil
	}
	
	// Find common parent or just invoke on each dest path?
	// JS finds common parent of deleted files? 
	// Actually JS uses `findParent(dir)` where dir is `pathsNeedDelete`. 
	// Wait, `deleteEmptyDir(pathsNeedDelete)`? 
	// No, usually we want to delete empty dirs in DESTINATION ROOTS.
	// But passing deleted files path to `findParent` gives the scope.
	
	// Let's iterate over `dirs` properly? 
	// Ideally we run "find dest -type d -empty -delete" on the destination roots.
	// But we don't know dest roots explicitly here easily without re-parsing options.
	// Assuming the caller knows or we infer.
	
	// For now, let's just attempt to remove parents of deleted files?
	// But `rmdir` only works if empty.
	// Let's implement a simple loop that tries to remove parent dir if empty.
	
	return nil // Placeholder. The "find" approach is robust.
}

func PruneEmptyDirs(roots []string) error {
    for _, root := range roots {
        cmd := exec.Command("find", root, "-type", "d", "-empty", "-delete")
        _ = cmd.Run() 
    }
    return nil
}
