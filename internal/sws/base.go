package sws

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sbl/embedded"
)

func loadBase(siteRoot string) (string, error) {
	override := filepath.Join(siteRoot, "deploy", "sws.base.toml")
	if data, err := os.ReadFile(override); err == nil {
		return string(data), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read SWS base config: %w", err)
	}

	data, err := fs.ReadFile(embedded.Deploy, "sws.base.toml")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
