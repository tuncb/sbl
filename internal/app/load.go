package app

import (
	"time"

	"sbl/internal/assets"
	"sbl/internal/content"
	"sbl/internal/site"
)

func loadValidatedSite(siteRoot, baseURL string, requireBaseURL, includeDrafts bool, report *timingReport) (site.Config, *content.Site, error) {
	start := time.Now()
	cfg, err := site.Load(siteRoot, baseURL, requireBaseURL)
	report.Add("load_site_config", time.Since(start))
	if err != nil {
		return site.Config{}, nil, err
	}
	cfg.PrismTheme, err = assets.NormalizePrismTheme(cfg.PrismTheme)
	if err != nil {
		return site.Config{}, nil, err
	}

	start = time.Now()
	if err := content.ValidateLayout(siteRoot); err != nil {
		report.Add("validate_layout", time.Since(start))
		return site.Config{}, nil, err
	}
	report.Add("validate_layout", time.Since(start))

	start = time.Now()
	posts, err := content.LoadPosts(siteRoot)
	report.Add("load_posts", time.Since(start))
	if err != nil {
		return site.Config{}, nil, err
	}

	start = time.Now()
	pages, err := content.LoadPages(siteRoot)
	report.Add("load_pages", time.Since(start))
	if err != nil {
		return site.Config{}, nil, err
	}

	start = time.Now()
	graph, err := content.Validate(posts, pages, includeDrafts)
	report.Add("validate_content", time.Since(start))
	if err != nil {
		return site.Config{}, nil, err
	}

	return cfg, graph, nil
}
