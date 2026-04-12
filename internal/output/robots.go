package output

import (
	"fmt"

	"sbl/internal/site"
)

func BuildRobots(cfg site.Config) []byte {
	return []byte(fmt.Sprintf("User-agent: *\nAllow: /\nSitemap: %s\n", cfg.CanonicalURL("/sitemap.xml")))
}
