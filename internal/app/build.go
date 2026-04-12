package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sbl/internal/assets"
	"sbl/internal/content"
	"sbl/internal/output"
	"sbl/internal/render"
	"sbl/internal/site"
	"sbl/internal/sws"
)

type BuildOptions struct {
	SiteRoot      string
	OutputDir     string
	BaseURL       string
	IncludeDrafts bool
	Clean         bool
	Stdout        io.Writer
}

func Build(opts BuildOptions) error {
	siteRoot, err := filepath.Abs(opts.SiteRoot)
	if err != nil {
		return err
	}
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(siteRoot, "public")
	} else if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(siteRoot, outputDir)
	}

	if opts.Clean {
		if err := os.RemoveAll(outputDir); err != nil {
			return err
		}
	}

	cfg, err := site.Load(siteRoot, opts.BaseURL, true)
	if err != nil {
		return err
	}

	posts, err := content.LoadPosts(siteRoot)
	if err != nil {
		return err
	}
	graph, err := content.Validate(posts, opts.IncludeDrafts)
	if err != nil {
		return err
	}

	engine, err := render.NewTemplateEngine(siteRoot)
	if err != nil {
		return err
	}

	staticFiles, stylesheetURL, err := assets.BuildStaticFiles(siteRoot)
	if err != nil {
		return err
	}
	for _, file := range staticFiles {
		if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
			return err
		}
	}

	postSummaries := make([]render.PostSummary, 0, len(graph.Posts))
	for _, post := range graph.Posts {
		postAssetFiles, postAssetURLs, err := assets.BuildPostAssets(post)
		if err != nil {
			return err
		}
		for _, file := range postAssetFiles {
			if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
				return err
			}
		}

		bodyHTML, generatedFiles, readingTime, err := render.RenderPostBody(post, postAssetURLs)
		if err != nil {
			return fmt.Errorf("render post %s: %w", post.Slug, err)
		}
		for _, file := range generatedFiles {
			if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
				return err
			}
		}

		pageHTML, summary, err := render.RenderPostPage(engine, cfg, stylesheetURL, post, bodyHTML, readingTime)
		if err != nil {
			return err
		}
		postSummaries = append(postSummaries, summary)
		if err := output.WriteFile(outputDir, filepath.ToSlash(filepath.Join("posts", post.Slug, "index.html")), pageHTML); err != nil {
			return err
		}
	}

	indexHTML, err := render.RenderIndexPage(engine, cfg, stylesheetURL, postSummaries)
	if err != nil {
		return err
	}
	if err := output.WriteFile(outputDir, "index.html", indexHTML); err != nil {
		return err
	}

	archiveHTML, err := render.RenderArchivePage(engine, cfg, stylesheetURL, postSummaries)
	if err != nil {
		return err
	}
	if err := output.WriteFile(outputDir, "archive/index.html", archiveHTML); err != nil {
		return err
	}

	notFoundHTML, err := render.RenderNotFoundPage(engine, cfg, stylesheetURL)
	if err != nil {
		return err
	}
	if err := output.WriteFile(outputDir, "404.html", notFoundHTML); err != nil {
		return err
	}

	tempErrorHTML, err := render.RenderTemporaryErrorPage(engine, cfg, stylesheetURL)
	if err != nil {
		return err
	}
	if err := output.WriteFile(outputDir, "50x.html", tempErrorHTML); err != nil {
		return err
	}

	feedXML, err := output.BuildFeed(cfg, graph.Posts)
	if err != nil {
		return err
	}
	if err := output.WriteFile(outputDir, "feed.xml", feedXML); err != nil {
		return err
	}

	sitemapXML, err := output.BuildSitemap(cfg, graph.Posts)
	if err != nil {
		return err
	}
	if err := output.WriteFile(outputDir, "sitemap.xml", sitemapXML); err != nil {
		return err
	}

	if err := output.WriteFile(outputDir, "robots.txt", output.BuildRobots(cfg)); err != nil {
		return err
	}

	if err := sws.Write(siteRoot, outputDir, graph.Aliases); err != nil {
		return err
	}

	if opts.Stdout != nil {
		fmt.Fprintf(opts.Stdout, "built %d posts into %s\n", len(graph.Posts), outputDir)
	}
	return nil
}
