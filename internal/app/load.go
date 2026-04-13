package app

import (
	"sbl/internal/content"
	"sbl/internal/site"
)

func loadValidatedSite(siteRoot, baseURL string, requireBaseURL, includeDrafts bool) (site.Config, *content.Site, error) {
	cfg, err := site.Load(siteRoot, baseURL, requireBaseURL)
	if err != nil {
		return site.Config{}, nil, err
	}
	if err := content.ValidateLayout(siteRoot); err != nil {
		return site.Config{}, nil, err
	}

	posts, err := content.LoadPosts(siteRoot)
	if err != nil {
		return site.Config{}, nil, err
	}
	pages, err := content.LoadPages(siteRoot)
	if err != nil {
		return site.Config{}, nil, err
	}

	graph, err := content.Validate(posts, pages, includeDrafts)
	if err != nil {
		return site.Config{}, nil, err
	}

	return cfg, graph, nil
}
