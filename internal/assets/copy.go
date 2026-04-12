package assets

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"sbl/embedded"
	"sbl/internal/content"
)

func BuildStaticFiles(siteRoot string) ([]File, string, error) {
	files := map[string][]byte{}
	if err := readFSFiles(embedded.Static, files); err != nil {
		return nil, "", err
	}

	overrideDir := filepath.Join(siteRoot, "static")
	if err := readDiskFiles(overrideDir, files); err != nil {
		return nil, "", err
	}

	paths := make([]string, 0, len(files))
	for rel := range files {
		paths = append(paths, rel)
	}
	sort.Strings(paths)

	out := make([]File, 0, len(paths))
	stylesheetURL := ""
	for _, rel := range paths {
		data := files[rel]
		var file File
		if isRootStaticFile(rel) {
			cleaned := path.Clean(rel)
			file = File{RelPath: cleaned, URL: "/" + cleaned, Bytes: data}
		} else {
			file = NewHashedFile(rel, data)
		}
		if path.Base(rel) == "site.css" {
			stylesheetURL = file.URL
		}
		out = append(out, file)
	}

	return out, stylesheetURL, nil
}

func BuildPostAssets(post *content.Post) ([]File, map[string]string, error) {
	return buildScopedAssets("posts", post.Slug, post.SourceDir)
}

func BuildPageAssets(page *content.Page) ([]File, map[string]string, error) {
	return buildScopedAssets("pages", page.Slug, page.SourceDir)
}

func buildScopedAssets(section, slug, sourceDir string) ([]File, map[string]string, error) {
	assetsDir := filepath.Join(sourceDir, "assets")
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return nil, map[string]string{}, nil
	} else if err != nil {
		return nil, nil, fmt.Errorf("stat post assets directory: %w", err)
	}

	fileMap := make(map[string][]byte)
	if err := readDiskFiles(assetsDir, fileMap); err != nil {
		return nil, nil, err
	}

	paths := make([]string, 0, len(fileMap))
	for rel := range fileMap {
		paths = append(paths, rel)
	}
	sort.Strings(paths)

	files := make([]File, 0, len(paths))
	urls := make(map[string]string, len(paths))
	for _, rel := range paths {
		data := fileMap[rel]
		file := NewHashedFile(path.Join(section, slug, rel), data)
		files = append(files, file)
		urls["assets/"+rel] = file.URL
	}
	return files, urls, nil
}

func readFSFiles(source fs.FS, files map[string][]byte) error {
	return fs.WalkDir(source, ".", func(rel string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(source, rel)
		if err != nil {
			return err
		}
		files[path.Clean(filepath.ToSlash(rel))] = data
		return nil
	})
}

func readDiskFiles(root string, files map[string][]byte) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat %s: %w", root, err)
	}

	return filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		files[path.Clean(filepath.ToSlash(rel))] = data
		return nil
	})
}

func isRootStaticFile(rel string) bool {
	switch strings.ToLower(path.Base(rel)) {
	case "favicon.ico", "favicon.svg", "site.webmanifest":
		return path.Dir(rel) == "."
	default:
		return false
	}
}
