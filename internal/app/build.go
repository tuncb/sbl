package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"sbl/internal/assets"
	"sbl/internal/output"
	"sbl/internal/render"
	"sbl/internal/sws"
)

type BuildOptions struct {
	SiteRoot      string
	OutputDir     string
	BaseURL       string
	IncludeDrafts bool
	Clean         bool
	Stdout        io.Writer
	Timings       bool
}

type buildResult struct {
	OutputDir string
	PostCount int
	PageCount int
}

func Build(opts BuildOptions) (err error) {
	report := newTimingReport()
	defer func() {
		if opts.Timings {
			report.Print(opts.Stdout)
		}
	}()

	result, err := buildSite(opts, report)
	if err != nil {
		return err
	}

	printBuildSummary(opts.Stdout, result)
	return nil
}

func buildSite(opts BuildOptions, report *timingReport) (_ buildResult, err error) {
	if report != nil {
		start := time.Now()
		defer func() {
			report.Add("total", time.Since(start))
		}()
	}

	siteRoot, err := filepath.Abs(opts.SiteRoot)
	if err != nil {
		return buildResult{}, err
	}
	outputDir := resolveOutputDir(siteRoot, opts.OutputDir)
	if err := validateOutputDir(siteRoot, outputDir); err != nil {
		return buildResult{}, err
	}

	cfg, graph, err := loadValidatedSite(siteRoot, opts.BaseURL, true, opts.IncludeDrafts, report)
	if err != nil {
		return buildResult{}, err
	}

	if opts.Clean {
		if err := measureTiming(report, "clean_output", func() error {
			return os.RemoveAll(outputDir)
		}); err != nil {
			return buildResult{}, err
		}
	}

	var engine *render.Engine
	if err := measureTiming(report, "load_templates", func() error {
		engine, err = render.NewTemplateEngine(siteRoot)
		return err
	}); err != nil {
		return buildResult{}, err
	}

	var staticAssets assets.StaticAssets
	if err := measureTiming(report, "build_static_assets", func() error {
		staticFiles, staticInfo, err := assets.BuildStaticFiles(siteRoot)
		if err != nil {
			return err
		}
		for _, file := range staticFiles {
			if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
				return err
			}
		}
		staticAssets = staticInfo
		return nil
	}); err != nil {
		return buildResult{}, err
	}

	var vendorAssets assets.VendorAssets
	if err := measureTiming(report, "build_vendor_assets", func() error {
		vendorFiles, vendorInfo, err := assets.BuildVendorFiles()
		if err != nil {
			return err
		}
		for _, file := range vendorFiles {
			if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
				return err
			}
		}
		vendorAssets = vendorInfo
		return nil
	}); err != nil {
		return buildResult{}, err
	}
	postSummaries := make([]render.PostSummary, 0, len(graph.Posts))
	if err := measureTiming(report, "render_posts", func() error {
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

			bodyHTML, generatedFiles, readingTime, features, err := render.RenderPostBody(post, postAssetURLs)
			if err != nil {
				return fmt.Errorf("render post %s: %w", post.Slug, err)
			}
			for _, file := range generatedFiles {
				if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
					return err
				}
			}

			pageHTML, summary, err := render.RenderPostPage(
				engine,
				cfg,
				staticAssets.StylesheetURL,
				extraStylesheets(features, vendorAssets),
				clientRenderConfig(features, staticAssets, vendorAssets),
				post,
				bodyHTML,
				readingTime,
			)
			if err != nil {
				return err
			}
			postSummaries = append(postSummaries, summary)
			if err := output.WriteFile(outputDir, filepath.ToSlash(filepath.Join("posts", post.Slug, "index.html")), pageHTML); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return buildResult{}, err
	}

	if err := measureTiming(report, "render_pages", func() error {
		for _, page := range graph.Pages {
			pageAssetFiles, pageAssetURLs, err := assets.BuildPageAssets(page)
			if err != nil {
				return err
			}
			for _, file := range pageAssetFiles {
				if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
					return err
				}
			}

			bodyHTML, generatedFiles, _, features, err := render.RenderPageBody(page, pageAssetURLs)
			if err != nil {
				return fmt.Errorf("render page %s: %w", page.Slug, err)
			}
			for _, file := range generatedFiles {
				if err := output.WriteFile(outputDir, file.RelPath, file.Bytes); err != nil {
					return err
				}
			}

			pageHTML, err := render.RenderStandalonePage(
				engine,
				cfg,
				staticAssets.StylesheetURL,
				extraStylesheets(features, vendorAssets),
				clientRenderConfig(features, staticAssets, vendorAssets),
				page,
				bodyHTML,
			)
			if err != nil {
				return err
			}
			if err := output.WriteFile(outputDir, filepath.ToSlash(filepath.Join("pages", page.Slug, "index.html")), pageHTML); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return buildResult{}, err
	}

	if err := measureTiming(report, "render_auxiliary", func() error {
		indexHTML, err := render.RenderIndexPage(engine, cfg, staticAssets.StylesheetURL, nil, nil, postSummaries)
		if err != nil {
			return err
		}
		if err := output.WriteFile(outputDir, "index.html", indexHTML); err != nil {
			return err
		}

		archiveHTML, err := render.RenderArchivePage(engine, cfg, staticAssets.StylesheetURL, nil, nil, postSummaries)
		if err != nil {
			return err
		}
		if err := output.WriteFile(outputDir, "archive/index.html", archiveHTML); err != nil {
			return err
		}

		notFoundHTML, err := render.RenderNotFoundPage(engine, cfg, staticAssets.StylesheetURL, nil, nil)
		if err != nil {
			return err
		}
		if err := output.WriteFile(outputDir, "404.html", notFoundHTML); err != nil {
			return err
		}

		tempErrorHTML, err := render.RenderTemporaryErrorPage(engine, cfg, staticAssets.StylesheetURL, nil, nil)
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

		sitemapXML, err := output.BuildSitemap(cfg, graph.Posts, graph.Pages)
		if err != nil {
			return err
		}
		if err := output.WriteFile(outputDir, "sitemap.xml", sitemapXML); err != nil {
			return err
		}

		if err := output.WriteFile(outputDir, "robots.txt", output.BuildRobots(cfg)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return buildResult{}, err
	}

	if err := measureTiming(report, "write_deploy_config", func() error {
		return sws.Write(siteRoot, outputDir, graph.Aliases)
	}); err != nil {
		return buildResult{}, err
	}

	return buildResult{
		OutputDir: outputDir,
		PostCount: len(graph.Posts),
		PageCount: len(graph.Pages),
	}, nil
}

func printBuildSummary(out io.Writer, result buildResult) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "built %d posts and %d pages into %s\n", result.PostCount, result.PageCount, result.OutputDir)
}

func clientRenderConfig(features render.Features, staticAssets assets.StaticAssets, vendorAssets assets.VendorAssets) *render.ClientRenderConfig {
	if !features.NeedsMath && !features.NeedsMermaid && !features.NeedsCodeHighlight {
		return nil
	}
	return &render.ClientRenderConfig{
		BootstrapURL:         staticAssets.ClientRenderURL,
		KaTeXCSSURL:          vendorAssets.KaTeXCSSURL,
		KaTeXJSURL:           vendorAssets.KaTeXJSURL,
		MermaidJSURL:         vendorAssets.MermaidJSURL,
		PrismCoreJSURL:       vendorAssets.PrismCoreJSURL,
		PrismAutoloaderJSURL: vendorAssets.PrismAutoloaderJSURL,
		PrismLanguagesPath:   vendorAssets.PrismLanguagesPath,
	}
}

func extraStylesheets(features render.Features, vendorAssets assets.VendorAssets) []string {
	if !features.NeedsCodeHighlight {
		return nil
	}
	return []string{vendorAssets.PrismCSSURL}
}
