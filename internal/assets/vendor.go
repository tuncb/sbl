package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type packageMetadata struct {
	Version string `json:"version"`
}

func BuildVendorFiles() ([]File, string, error) {
	root := moduleRoot()
	packageDir := filepath.Join(root, "node_modules", "katex")
	metadata, err := readPackageMetadata(filepath.Join(packageDir, "package.json"))
	if err != nil {
		return nil, "", fmt.Errorf("read katex metadata: %w", err)
	}

	baseDir := filepath.Join(packageDir, "dist")
	cssData, err := os.ReadFile(filepath.Join(baseDir, "katex.min.css"))
	if err != nil {
		return nil, "", fmt.Errorf("read katex css: %w", err)
	}

	targetBase := path.Join("assets", "vendor", "katex-"+metadata.Version)
	files := []File{
		{
			RelPath: path.Join(targetBase, "katex.min.css"),
			URL:     "/" + path.Join(targetBase, "katex.min.css"),
			Bytes:   cssData,
		},
	}

	fontFiles, err := readDirFiles(filepath.Join(baseDir, "fonts"))
	if err != nil {
		return nil, "", fmt.Errorf("read katex fonts: %w", err)
	}
	for rel, data := range fontFiles {
		files = append(files, File{
			RelPath: path.Join(targetBase, "fonts", rel),
			URL:     "/" + path.Join(targetBase, "fonts", rel),
			Bytes:   data,
		})
	}

	return files, "/" + path.Join(targetBase, "katex.min.css"), nil
}

func readPackageMetadata(path string) (packageMetadata, error) {
	var metadata packageMetadata
	data, err := os.ReadFile(path)
	if err != nil {
		return metadata, err
	}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return metadata, err
	}
	return metadata, nil
}

func readDirFiles(root string) (map[string][]byte, error) {
	files := map[string][]byte{}
	if err := readDiskFiles(root, files); err != nil {
		return nil, err
	}
	return files, nil
}

func moduleRoot() string {
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile()), "..", ".."))
}

func currentFile() string {
	_, file, _, _ := runtime.Caller(0)
	return file
}
