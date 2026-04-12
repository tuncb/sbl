package content

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func LoadPosts(root string) ([]*Post, error) {
	return loadCollection(root, "posts", parsePostFile)
}

func LoadPages(root string) ([]*Page, error) {
	return loadCollection(root, "pages", parsePageFile)
}

func loadCollection[T any](root, section string, parse func(path, slug string) (*T, error)) ([]*T, error) {
	contentDir := filepath.Join(root, "content", section)
	entries, err := os.ReadDir(contentDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s directory: %w", section, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	items := make([]*T, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		indexPath := filepath.Join(contentDir, entry.Name(), "index.md")
		if _, err := os.Stat(indexPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("stat %s: %w", indexPath, err)
		}

		item, err := parse(indexPath, entry.Name())
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
