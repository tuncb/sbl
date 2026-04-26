package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func Validate(posts []*Post, pages []*Page, includeDrafts bool) (*Site, error) {
	errs := &ValidationErrors{}
	activePosts := make([]*Post, 0, len(posts))
	activePages := make([]*Page, 0, len(pages))

	for _, post := range posts {
		validatePostMetadata(post, errs)
		if !post.Draft || includeDrafts {
			activePosts = append(activePosts, post)
		}
	}
	for _, page := range pages {
		validatePageMetadata(page, errs)
		if !page.Draft || includeDrafts {
			activePages = append(activePages, page)
		}
	}

	sort.Slice(activePosts, func(i, j int) bool {
		if !activePosts[i].Date.Equal(activePosts[j].Date) {
			return activePosts[i].Date.After(activePosts[j].Date)
		}
		return activePosts[i].Slug < activePosts[j].Slug
	})
	sort.Slice(activePages, func(i, j int) bool {
		return activePages[i].Slug < activePages[j].Slug
	})

	canonical := make(map[string]string, len(activePosts)+len(activePages))
	aliases := make(map[string]string)

	registerPostCanonicals(activePosts, canonical, aliases, errs)
	registerPageCanonicals(activePages, canonical, aliases, errs)

	knownPaths := map[string]struct{}{
		"/":            {},
		"/archive/":    {},
		"/feed.xml":    {},
		"/sitemap.xml": {},
		"/robots.txt":  {},
	}
	for route := range canonical {
		knownPaths[route] = struct{}{}
	}
	for alias := range aliases {
		knownPaths[alias] = struct{}{}
	}

	for _, post := range activePosts {
		validateLocalAssets(post.SourceDir, post.SourcePath, post.MarkdownBody, errs)
		validateInternalLinks(post.SourcePath, post.MarkdownBody, post.MarkdownLine, knownPaths, errs)
	}
	for _, page := range activePages {
		validateLocalAssets(page.SourceDir, page.SourcePath, page.MarkdownBody, errs)
		validateInternalLinks(page.SourcePath, page.MarkdownBody, page.MarkdownLine, knownPaths, errs)
	}

	if err := errs.ErrOrNil(); err != nil {
		return nil, err
	}

	return &Site{
		Posts:     activePosts,
		Pages:     activePages,
		Canonical: canonical,
		Aliases:   aliases,
	}, nil
}

func validatePostMetadata(post *Post, errs *ValidationErrors) {
	if !slugPattern.MatchString(post.Slug) {
		errs.Add(fmt.Sprintf("invalid slug %q in %s", post.Slug, post.SourcePath))
	}
	if post.Title == "" {
		errs.Add(fmt.Sprintf("missing title in %s", post.SourcePath))
	}
	if post.Date.IsZero() {
		errs.Add(fmt.Sprintf("missing or invalid date in %s", post.SourcePath))
	}
	if post.Summary == "" {
		errs.Add(fmt.Sprintf("missing summary in %s", post.SourcePath))
	}
	if post.Updated != nil && post.Updated.Before(post.Date) {
		errs.Add(fmt.Sprintf("updated date must not be earlier than date in %s", post.SourcePath))
	}
}

func validatePageMetadata(page *Page, errs *ValidationErrors) {
	if !slugPattern.MatchString(page.Slug) {
		errs.Add(fmt.Sprintf("invalid slug %q in %s", page.Slug, page.SourcePath))
	}
	if page.Title == "" {
		errs.Add(fmt.Sprintf("missing title in %s", page.SourcePath))
	}
}

func validateLocalAssets(sourceDir, sourcePath, markdown string, errs *ValidationErrors) {
	for _, ref := range CollectLocalAssetRefs(markdown) {
		fullPath := filepath.Join(sourceDir, filepath.FromSlash(ref))
		if _, err := os.Stat(fullPath); err != nil {
			errs.Add(fmt.Sprintf("missing asset %q referenced by %s", ref, sourcePath))
		}
	}
}

func validateInternalLinks(sourcePath, markdown string, markdownLine int, knownPaths map[string]struct{}, errs *ValidationErrors) {
	for _, link := range collectInternalLinks(markdown) {
		route, err := NormalizeURLPath(link.URL)
		sourceLine := markdownLine + link.Line - 1
		if err != nil {
			errs.Add(fmt.Sprintf("invalid internal link %q in %s:%d: %v", link.URL, sourcePath, sourceLine, err))
			continue
		}
		if strings.HasPrefix(route, "/assets/") {
			continue
		}
		if _, exists := knownPaths[route]; !exists {
			errs.Add(fmt.Sprintf("broken internal link %q in %s:%d (markdown target %q)", route, sourcePath, sourceLine, link.URL))
		}
	}
}

func registerPostCanonicals(posts []*Post, canonical map[string]string, aliases map[string]string, errs *ValidationErrors) {
	for _, post := range posts {
		if existing, exists := canonical[post.CanonicalPath]; exists {
			errs.Add(fmt.Sprintf("canonical path collision: %s and %s both map to %s", existing, post.SourcePath, post.CanonicalPath))
			continue
		}
		canonical[post.CanonicalPath] = post.SourcePath
		registerAliases(post.SourcePath, post.Aliases, post.CanonicalPath, canonical, aliases, errs)
	}
}

func registerPageCanonicals(pages []*Page, canonical map[string]string, aliases map[string]string, errs *ValidationErrors) {
	for _, page := range pages {
		if existing, exists := canonical[page.CanonicalPath]; exists {
			errs.Add(fmt.Sprintf("canonical path collision: %s and %s both map to %s", existing, page.SourcePath, page.CanonicalPath))
			continue
		}
		canonical[page.CanonicalPath] = page.SourcePath
		registerAliases(page.SourcePath, page.Aliases, page.CanonicalPath, canonical, aliases, errs)
	}
}

func registerAliases(sourcePath string, rawAliases []string, destination string, canonical map[string]string, aliases map[string]string, errs *ValidationErrors) {
	for _, rawAlias := range rawAliases {
		alias, err := NormalizeURLPath(rawAlias)
		if err != nil {
			errs.Add(fmt.Sprintf("invalid alias %q in %s: %v", rawAlias, sourcePath, err))
			continue
		}
		if alias == "" || !strings.HasPrefix(alias, "/") {
			errs.Add(fmt.Sprintf("alias %q in %s must be an absolute site path", rawAlias, sourcePath))
			continue
		}
		if _, exists := canonical[alias]; exists {
			errs.Add(fmt.Sprintf("alias collision: %s conflicts with canonical route %s", alias, alias))
			continue
		}
		if target, exists := aliases[alias]; exists {
			errs.Add(fmt.Sprintf("alias collision: %s is claimed by both %s and %s", alias, target, destination))
			continue
		}
		aliases[alias] = destination
	}
}
