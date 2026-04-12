package site

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(root, baseURLOverride string, requireBaseURL bool) (Config, error) {
	cfg := Config{
		Title:       "sbl",
		Description: "",
		Language:    "en",
		Author:      "",
		Navigation: []NavLink{
			{Label: "Archive", URL: "/archive/"},
		},
	}

	configPath := filepath.Join(root, "config", "site.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse site config: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read site config: %w", err)
	}

	if baseURLOverride != "" {
		cfg.BaseURL = baseURLOverride
	}
	if cfg.Title == "" {
		cfg.Title = "sbl"
	}
	if cfg.Language == "" {
		cfg.Language = "en"
	}
	if len(cfg.Navigation) == 0 {
		cfg.Navigation = []NavLink{{Label: "Archive", URL: "/archive/"}}
	}

	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if requireBaseURL && cfg.BaseURL == "" {
		return Config{}, errors.New("site base_url is required for build output")
	}
	if cfg.BaseURL != "" {
		parsed, err := url.Parse(cfg.BaseURL)
		if err != nil {
			return Config{}, fmt.Errorf("parse base_url: %w", err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return Config{}, errors.New("site base_url must be an absolute URL")
		}
	}

	return cfg, nil
}
