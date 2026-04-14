package app

import (
	"io"
	"path/filepath"
	"time"
)

type ValidateOptions struct {
	SiteRoot      string
	BaseURL       string
	IncludeDrafts bool
	Stdout        io.Writer
	Timings       bool
}

func Validate(opts ValidateOptions) (err error) {
	start := time.Now()
	report := newTimingReport()
	defer func() {
		report.Add("total", time.Since(start))
		if opts.Timings {
			report.Print(opts.Stdout)
		}
	}()

	siteRoot, err := filepath.Abs(opts.SiteRoot)
	if err != nil {
		return err
	}
	_, _, err = loadValidatedSite(siteRoot, opts.BaseURL, false, opts.IncludeDrafts, report)
	return err
}
