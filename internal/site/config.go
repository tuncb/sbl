package site

import "strings"

type NavLink struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

type Config struct {
	Title       string    `yaml:"title"`
	BaseURL     string    `yaml:"base_url"`
	Description string    `yaml:"description"`
	Language    string    `yaml:"language"`
	Author      string    `yaml:"author"`
	Navigation  []NavLink `yaml:"navigation"`
}

func (c Config) CanonicalURL(relPath string) string {
	base := strings.TrimRight(c.BaseURL, "/")
	if relPath == "" {
		return base
	}
	if strings.HasPrefix(relPath, "/") {
		return base + relPath
	}
	return base + "/" + relPath
}
