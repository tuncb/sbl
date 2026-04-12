package output

import (
	"encoding/xml"
	"time"

	"sbl/internal/content"
	"sbl/internal/site"
)

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`
	Items         []item `xml:"item"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func BuildFeed(cfg site.Config, posts []*content.Post) ([]byte, error) {
	items := make([]item, 0, len(posts))
	lastBuild := ""
	if len(posts) > 0 {
		lastBuild = formatFeedDate(lastUpdated(posts[0]))
	}
	for _, post := range posts {
		link := cfg.CanonicalURL(post.CanonicalPath)
		items = append(items, item{
			Title:       post.Title,
			Link:        link,
			GUID:        link,
			PubDate:     formatFeedDate(lastUpdated(post)),
			Description: post.Summary,
		})
	}

	payload, err := xml.MarshalIndent(rss{
		Version: "2.0",
		Channel: channel{
			Title:         cfg.Title,
			Link:          cfg.CanonicalURL("/"),
			Description:   cfg.Description,
			LastBuildDate: lastBuild,
			Items:         items,
		},
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), payload...), nil
}

func lastUpdated(post *content.Post) time.Time {
	if post.Updated != nil {
		return *post.Updated
	}
	return post.Date
}

func formatFeedDate(value time.Time) string {
	return value.Format(time.RFC1123Z)
}
