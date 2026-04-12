package content

import (
	"strings"
	"time"
)

type Post struct {
	Slug          string
	SourceDir     string
	SourcePath    string
	Title         string
	Date          time.Time
	Updated       *time.Time
	Summary       string
	Draft         bool
	Tags          []string
	Aliases       []string
	Description   string
	Image         string
	MarkdownBody  string
	CanonicalPath string
}

type Page struct {
	Slug          string
	SourceDir     string
	SourcePath    string
	Title         string
	Summary       string
	Draft         bool
	Aliases       []string
	Description   string
	Image         string
	MarkdownBody  string
	CanonicalPath string
}

type Site struct {
	Posts     []*Post
	Pages     []*Page
	Canonical map[string]string
	Aliases   map[string]string
}

type ValidationErrors struct {
	Messages []string
}

func (v *ValidationErrors) Add(message string) {
	v.Messages = append(v.Messages, message)
}

func (v *ValidationErrors) Error() string {
	return strings.Join(v.Messages, "\n")
}

func (v *ValidationErrors) ErrOrNil() error {
	if len(v.Messages) == 0 {
		return nil
	}
	return v
}
