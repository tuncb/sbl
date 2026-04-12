package output

import (
	"encoding/xml"

	"sbl/internal/content"
	"sbl/internal/site"
)

type urlSet struct {
	XMLName xml.Name  `xml:"urlset"`
	Xmlns   string    `xml:"xmlns,attr"`
	URLs    []siteURL `xml:"url"`
}

type siteURL struct {
	Loc string `xml:"loc"`
}

func BuildSitemap(cfg site.Config, posts []*content.Post) ([]byte, error) {
	urls := []siteURL{
		{Loc: cfg.CanonicalURL("/")},
		{Loc: cfg.CanonicalURL("/archive/")},
	}
	for _, post := range posts {
		urls = append(urls, siteURL{Loc: cfg.CanonicalURL(post.CanonicalPath)})
	}
	payload, err := xml.MarshalIndent(urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), payload...), nil
}
