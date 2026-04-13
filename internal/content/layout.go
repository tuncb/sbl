package content

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ValidateLayout(root string) error {
	contentDir := filepath.Join(root, "content")
	info, err := os.Stat(contentDir)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("site content directory does not exist: %s", contentDir)
	}
	if err != nil {
		return fmt.Errorf("stat site content directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("site content path is not a directory: %s", contentDir)
	}

	postsExists, err := sectionExists(contentDir, "posts")
	if err != nil {
		return err
	}
	pagesExists, err := sectionExists(contentDir, "pages")
	if err != nil {
		return err
	}
	if !postsExists && !pagesExists {
		return fmt.Errorf(
			"site content directory must include at least one of %s or %s",
			filepath.Join(contentDir, "posts"),
			filepath.Join(contentDir, "pages"),
		)
	}

	return nil
}

func sectionExists(contentDir, section string) (bool, error) {
	sectionDir := filepath.Join(contentDir, section)
	info, err := os.Stat(sectionDir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat %s directory: %w", section, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("site %s path is not a directory: %s", section, sectionDir)
	}
	return true, nil
}
