package app

import (
	"path/filepath"
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
	_, _, err = loadValidatedSite(siteRoot, opts.BaseURL, false, opts.IncludeDrafts)
	return err
}
