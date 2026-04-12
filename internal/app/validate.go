package app

import (
	"path/filepath"

	"sbl/internal/content"
	"sbl/internal/site"
)

type ValidateOptions struct {
	SiteRoot      string
	BaseURL       string
	IncludeDrafts bool
}

func Validate(opts ValidateOptions) error {
	siteRoot, err := filepath.Abs(opts.SiteRoot)
	if err != nil {
		return err
	}
	if _, err := site.Load(siteRoot, opts.BaseURL, false); err != nil {
		return err
	}
	posts, err := content.LoadPosts(siteRoot)
	if err != nil {
		return err
	}
	pages, err := content.LoadPages(siteRoot)
	if err != nil {
		return err
	}
	_, err = content.Validate(posts, pages, opts.IncludeDrafts)
	return err
}
