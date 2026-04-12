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

func Validate(posts []*Post, includeDrafts bool) (*Site, error) {
	errs := &ValidationErrors{}
	active := make([]*Post, 0, len(posts))

	for _, post := range posts {
		validatePostMetadata(post, errs)
		if !post.Draft || includeDrafts {
			active = append(active, post)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		if !active[i].Date.Equal(active[j].Date) {
			return active[i].Date.After(active[j].Date)
		}
		return active[i].Slug < active[j].Slug
	})

	canonical := make(map[string]*Post, len(active))
	aliases := make(map[string]string)

	for _, post := range active {
		if existing, exists := canonical[post.CanonicalPath]; exists {
			errs.Add(fmt.Sprintf("canonical path collision: %s and %s both map to %s", existing.SourcePath, post.SourcePath, post.CanonicalPath))
			continue
		}
		canonical[post.CanonicalPath] = post
		for _, rawAlias := range post.Aliases {
			alias, err := NormalizeURLPath(rawAlias)
			if err != nil {
				errs.Add(fmt.Sprintf("invalid alias %q in %s: %v", rawAlias, post.SourcePath, err))
				continue
			}
			if alias == "" || !strings.HasPrefix(alias, "/") {
				errs.Add(fmt.Sprintf("alias %q in %s must be an absolute site path", rawAlias, post.SourcePath))
				continue
			}
			if _, exists := canonical[alias]; exists {
				errs.Add(fmt.Sprintf("alias collision: %s conflicts with canonical route %s", alias, alias))
				continue
			}
			if target, exists := aliases[alias]; exists {
				errs.Add(fmt.Sprintf("alias collision: %s is claimed by both %s and %s", alias, target, post.CanonicalPath))
				continue
			}
			aliases[alias] = post.CanonicalPath
		}
	}

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

	for _, post := range active {
		validateLocalAssets(post, errs)
		validateInternalLinks(post, knownPaths, errs)
	}

	if err := errs.ErrOrNil(); err != nil {
		return nil, err
	}

	return &Site{
		Posts:     active,
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

func validateLocalAssets(post *Post, errs *ValidationErrors) {
	for _, ref := range CollectLocalAssetRefs(post.MarkdownBody) {
		fullPath := filepath.Join(post.SourceDir, filepath.FromSlash(ref))
		if _, err := os.Stat(fullPath); err != nil {
			errs.Add(fmt.Sprintf("missing asset %q referenced by %s", ref, post.SourcePath))
		}
	}
}

func validateInternalLinks(post *Post, knownPaths map[string]struct{}, errs *ValidationErrors) {
	for _, rawLink := range CollectInternalLinks(post.MarkdownBody) {
		route, err := NormalizeURLPath(rawLink)
		if err != nil {
			errs.Add(fmt.Sprintf("invalid internal link %q in %s: %v", rawLink, post.SourcePath, err))
			continue
		}
		if strings.HasPrefix(route, "/assets/") {
			continue
		}
		if _, exists := knownPaths[route]; !exists {
			errs.Add(fmt.Sprintf("broken internal link %q in %s", route, post.SourcePath))
		}
	}
}
