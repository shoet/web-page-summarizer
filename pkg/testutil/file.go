package testutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetProjectRootDir(cwd string, maxDepth int) (string, error) {
	if maxDepth == 0 {
		return "", fmt.Errorf("reached the maximum depth")
	}
	entries, err := os.ReadDir(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to ReadDir")
	}
	for _, ent := range entries {
		if ent.Name() == "go.mod" {
			absPath, err := filepath.Abs(cwd)
			if err != nil {
				return "", fmt.Errorf("failed to filepath.Abs: %v", err)
			}
			return absPath, nil
		}
	}
	nextDir := filepath.Join(cwd, "..")
	return GetProjectRootDir(nextDir, maxDepth-1)
}
