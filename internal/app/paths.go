package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveOutputDir(siteRoot, outputDir string) string {
	if outputDir == "" {
		return filepath.Join(siteRoot, "public")
	}
	if filepath.IsAbs(outputDir) {
		return filepath.Clean(outputDir)
	}
	return filepath.Join(siteRoot, outputDir)
}

func validateOutputDir(siteRoot, outputDir string) error {
	if isFilesystemRoot(outputDir) {
		return fmt.Errorf("output directory must not be the filesystem root: %s", outputDir)
	}

	if info, err := os.Stat(outputDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("output path is not a directory: %s", outputDir)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat output directory: %w", err)
	}

	containsRoot, err := pathContains(outputDir, siteRoot)
	if err != nil {
		return err
	}
	if containsRoot {
		return fmt.Errorf("output directory must not contain the site root: %s", outputDir)
	}

	for _, reserved := range []string{
		filepath.Join(siteRoot, "config"),
		filepath.Join(siteRoot, "content"),
		filepath.Join(siteRoot, "templates"),
		filepath.Join(siteRoot, "static"),
		filepath.Join(siteRoot, "deploy"),
	} {
		insideReserved, err := pathContains(reserved, outputDir)
		if err != nil {
			return err
		}
		if insideReserved {
			return fmt.Errorf("output directory must not be inside site input directory: %s", reserved)
		}
	}

	return nil
}

func pathContains(parent, child string) (bool, error) {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	if filepath.VolumeName(parent) != filepath.VolumeName(child) {
		return false, nil
	}

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false, fmt.Errorf("compare %s and %s: %w", parent, child, err)
	}
	if rel == "." {
		return true, nil
	}

	prefix := ".." + string(os.PathSeparator)
	return rel != ".." && !strings.HasPrefix(rel, prefix), nil
}

func isFilesystemRoot(path string) bool {
	cleaned := filepath.Clean(path)
	volume := filepath.VolumeName(cleaned)
	if volume != "" {
		return cleaned == volume+string(os.PathSeparator)
	}
	return cleaned == string(os.PathSeparator)
}
