package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func RepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func CopyFixture(t *testing.T, name string) string {
	t.Helper()
	source := filepath.Join(RepoRoot(t), "testdata", name)
	target := filepath.Join(t.TempDir(), "site")
	if err := copyDir(source, target); err != nil {
		t.Fatalf("copy fixture %s: %v", name, err)
	}
	return target
}

func ReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(data)
}

func MustGlobOne(t *testing.T, pattern string) string {
	t.Helper()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one match for %s, got %d", pattern, len(matches))
	}
	return matches[0]
}

func copyDir(source, target string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0o644)
	})
}
