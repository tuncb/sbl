package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteFile(baseDir, relPath string, data []byte) error {
	cleaned := filepath.ToSlash(relPath)
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = filepath.ToSlash(filepath.Clean(cleaned))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("invalid output path %q", relPath)
	}

	fullPath := filepath.Join(baseDir, filepath.FromSlash(cleaned))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0o644)
}
