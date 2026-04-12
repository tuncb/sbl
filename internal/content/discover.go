package content

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func LoadPosts(root string) ([]*Post, error) {
	postsDir := filepath.Join(root, "content", "posts")
	entries, err := os.ReadDir(postsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read posts directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	posts := make([]*Post, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		indexPath := filepath.Join(postsDir, entry.Name(), "index.md")
		if _, err := os.Stat(indexPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("stat %s: %w", indexPath, err)
		}

		post, err := parsePostFile(indexPath, entry.Name())
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}
