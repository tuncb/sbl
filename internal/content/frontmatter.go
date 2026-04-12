package content

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type frontMatter struct {
	Title       string   `yaml:"title"`
	Date        string   `yaml:"date"`
	Updated     string   `yaml:"updated"`
	Summary     string   `yaml:"summary"`
	Draft       bool     `yaml:"draft"`
	Tags        []string `yaml:"tags"`
	Aliases     []string `yaml:"aliases"`
	Description string   `yaml:"description"`
	Image       string   `yaml:"image"`
}

func parsePostFile(path, slug string) (*Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	fmData, body, err := splitFrontMatter(data)
	if err != nil {
		return nil, fmt.Errorf("parse front matter in %s: %w", path, err)
	}

	var fm frontMatter
	if err := yaml.Unmarshal(fmData, &fm); err != nil {
		return nil, fmt.Errorf("decode front matter in %s: %w", path, err)
	}

	date, err := parseDate(fm.Date)
	if err != nil {
		return nil, fmt.Errorf("parse date in %s: %w", path, err)
	}

	var updated *time.Time
	if strings.TrimSpace(fm.Updated) != "" {
		value, err := parseDate(fm.Updated)
		if err != nil {
			return nil, fmt.Errorf("parse updated in %s: %w", path, err)
		}
		updated = &value
	}

	return &Post{
		Slug:          slug,
		SourceDir:     filepath.Dir(path),
		SourcePath:    path,
		Title:         strings.TrimSpace(fm.Title),
		Date:          date,
		Updated:       updated,
		Summary:       strings.TrimSpace(fm.Summary),
		Draft:         fm.Draft,
		Tags:          fm.Tags,
		Aliases:       fm.Aliases,
		Description:   strings.TrimSpace(fm.Description),
		Image:         strings.TrimSpace(fm.Image),
		MarkdownBody:  strings.TrimSpace(body),
		CanonicalPath: "/posts/" + slug + "/",
	}, nil
}

func parsePageFile(path, slug string) (*Page, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	fmData, body, err := splitFrontMatter(data)
	if err != nil {
		return nil, fmt.Errorf("parse front matter in %s: %w", path, err)
	}

	var fm frontMatter
	if err := yaml.Unmarshal(fmData, &fm); err != nil {
		return nil, fmt.Errorf("decode front matter in %s: %w", path, err)
	}

	return &Page{
		Slug:          slug,
		SourceDir:     filepath.Dir(path),
		SourcePath:    path,
		Title:         strings.TrimSpace(fm.Title),
		Summary:       strings.TrimSpace(fm.Summary),
		Draft:         fm.Draft,
		Aliases:       fm.Aliases,
		Description:   strings.TrimSpace(fm.Description),
		Image:         strings.TrimSpace(fm.Image),
		MarkdownBody:  strings.TrimSpace(body),
		CanonicalPath: "/pages/" + slug + "/",
	}, nil
}

func splitFrontMatter(data []byte) ([]byte, string, error) {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return nil, "", errors.New("missing opening --- line")
	}

	var frontMatterLines [][]byte
	for index := 1; index < len(lines); index++ {
		line := strings.TrimRight(string(lines[index]), "\r")
		if strings.TrimSpace(line) == "---" {
			body := bytes.Join(lines[index+1:], []byte("\n"))
			return bytes.Join(frontMatterLines, []byte("\n")), string(body), nil
		}
		frontMatterLines = append(frontMatterLines, []byte(line))
	}

	return nil, "", errors.New("missing closing --- line")
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}
